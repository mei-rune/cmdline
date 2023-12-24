// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package parser

import (
	"errors"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/mattn/go-shellwords"
)

type serviceExtractorFn func(cmdline *CommandLine) error

const (
	javaJarFlag      = "-jar"
	javaJarExtension = ".jar"
	javaApachePrefix = "org.apache."
)

// List of binaries that usually have additional process context of whats running
var binsWithContext = map[string]serviceExtractorFn{
	"python":    parseCommandContextPython,
	"python2.7": parseCommandContextPython,
	"python3":   parseCommandContextPython,
	"python3.7": parseCommandContextPython,
	"ruby2.3":   parseCommandContextRuby,
	"ruby":      parseCommandContextRuby,
	"java":      parseCommandContextJava,
	"java.exe":  parseCommandContextJava,
	"sudo":      parseCommandContext,
}

type CommandLine struct {
	ExecutePath string
	Args        []string

	Sub    *SubCommand
	Ruby   *RubyArgs
	Python *PythonArgs
	Java   *JavaArgs
}

type SubCommand struct {
	Command string
	Args        []string
}

type RubyArgs struct {
	FilePath string
	Args        []string
}

type PythonArgs struct {
	FilePath string
	Args        []string
}

type JavaArgs struct {
	ClassName string
	JmxEnable bool
	JmxPort         string
	JmxSsl          bool
	JmxAuthenticate bool

	Args        []string
}

func ParseCommandLine(s string) (*CommandLine, error) {
	if len(s) == 0 {
		return &CommandLine{}, nil
	}

	_, args, err := shellwords.ParseWithEnvs(s)
	if err != nil {
		return nil, err
	}
	if len(args) == 0 {
		return nil, errors.New("invalid command - `" + s + "`")
	}

	exe := args[0]
	// trim any quotes from the executable
	exe = strings.Trim(exe, "\"")
	return Parse(exe, args[1:])
}

func Parse(exe string, args []string) (*CommandLine, error) {
	c := &CommandLine{
		ExecutePath: exe,
		Args:        args,
	}

	exe = removeFilePath(exe)

	if ext := filepath.Ext(exe); strings.ToLower(ext) == ".exe" {
		exe = strings.TrimSuffix(exe, ext)
	}

	if contextFn, ok := binsWithContext[exe]; ok {
		return c, contextFn(c)
	}

	baseExe, _ := splitVersion(exe)
	if contextFn, ok := binsWithContext[baseExe]; ok {
		return c, contextFn(c)
	}

	// // trim trailing file extensions
	// if i := strings.LastIndex(exe, "."); i > 0 {
	// 	exe = exe[:i]
	// }

	return c, nil
}

func removeFilePath(s string) string {
	if s != "" {
		return filepath.Base(s)
	}
	return s
}

func splitVersion(s string) (string, string) {
	runes := []rune(s)
	for index := len(runes) - 1; index >= 0; index-- {
		if !unicode.IsDigit(runes[index]) && runes[index] != '.' {
			return s[:index+1], s[index+1:]
		}
	}
	return "", s
}

// In most cases, the best context is the first non-argument / environment variable, if it exists
func parseCommandContext(cmdline *CommandLine) error {
	var prevArgIsFlag bool

	for idx, a := range cmdline.Args {
		hasFlagPrefix, isEnvVariable := strings.HasPrefix(a, "-"), strings.ContainsRune(a, '=')
		shouldSkipArg := prevArgIsFlag || hasFlagPrefix || isEnvVariable

		if !shouldSkipArg {
			cmdline.Sub = &SubCommand{
				Command: a,
				Args: cmdline.Args[idx+1:],
			}
			return nil
		}

		prevArgIsFlag = hasFlagPrefix
	}
	return errors.New("scriptfile not found")
}


// In most cases, the best context is the first non-argument / environment variable, if it exists
func parseCommandContextRuby(cmdline *CommandLine) error {
	var prevArgIsFlag bool

	for idx, a := range cmdline.Args {
		hasFlagPrefix, isEnvVariable := strings.HasPrefix(a, "-"), strings.ContainsRune(a, '=')
		shouldSkipArg := prevArgIsFlag || hasFlagPrefix || isEnvVariable

		if !shouldSkipArg {
			cmdline.Ruby = &RubyArgs{
				FilePath: a,
				Args: cmdline.Args[idx+1:],
			}
			return nil
		}

		prevArgIsFlag = hasFlagPrefix
	}
	return errors.New("scriptfile not found")
}

func parseCommandContextPython(cmdline *CommandLine) error {
	var (
		prevArgIsFlag bool
		moduleFlag    bool
	)

	for idx, a := range cmdline.Args {
		hasFlagPrefix, isEnvVariable := strings.HasPrefix(a, "-"), strings.ContainsRune(a, '=')

		shouldSkipArg := prevArgIsFlag || hasFlagPrefix || isEnvVariable

		if !shouldSkipArg || moduleFlag {
			cmdline.Python = &PythonArgs{
				FilePath: a,
				Args: cmdline.Args[idx+1:],
			}
			return nil
		}

		if hasFlagPrefix && a == "-m" {
			moduleFlag = true
		}

		prevArgIsFlag = hasFlagPrefix
	}

	return errors.New("scriptfile not found")
}

func parseCommandContextJava(cmdline *CommandLine) error {
	var jmxremoteEnable bool
	var jmxremotePort string
	var jmxremoteSsl bool
	var jmxremoteAuthenticate bool

	prevArgIsFlag := false

	for idx, a := range cmdline.Args {
		hasFlagPrefix := strings.HasPrefix(a, "-")
		includesAssignment := strings.ContainsRune(a, '=') ||
			strings.HasPrefix(a, "-X") ||
			strings.HasPrefix(a, "-javaagent:") ||
			strings.HasPrefix(a, "-verbose:")
		shouldSkipArg := prevArgIsFlag || hasFlagPrefix || includesAssignment
		if !shouldSkipArg {
			cmdline.Java = &JavaArgs{
				ClassName: a,
				Args: cmdline.Args[idx+1:],
				JmxEnable: jmxremoteEnable,
				JmxPort: jmxremotePort,
				JmxSsl: jmxremoteSsl,
				JmxAuthenticate: jmxremoteAuthenticate,
			}
			return nil
		}

		if strings.HasPrefix(a, "-Dcom.sun.management.jmxremote=") {
			s := strings.TrimPrefix(a, "-Dcom.sun.management.jmxremote=")
			jmxremoteEnable = strings.ToLower(s) == "true"
		}

		if strings.HasPrefix(a, "-Dcom.sun.management.jmxremote.port=") {
			jmxremotePort = strings.TrimPrefix(a, "-Dcom.sun.management.jmxremote.port=")
		}

		if strings.HasPrefix(a, "-Dcom.sun.management.jmxremote.ssl") {
			s := strings.TrimPrefix(a, "-Dcom.sun.management.jmxremote.ssl")
			jmxremoteSsl = strings.ToLower(s) == "true"
		}

		if strings.HasPrefix(a, "-Dcom.sun.management.jmxremote.authenticate") {
			s := strings.TrimPrefix(a, "-Dcom.sun.management.jmxremote.authenticate")
			jmxremoteAuthenticate = strings.ToLower(s) == "true"
		}

		prevArgIsFlag = hasFlagPrefix && !includesAssignment && a != javaJarFlag
	}
	return errors.New("classname not found")
}
