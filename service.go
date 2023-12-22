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

type serviceExtractorFn func(args []string) (*CommandLine, error)

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
	"ruby2.3":   parseCommandContext,
	"ruby":      parseCommandContext,
	"java":      parseCommandContextJava,
	"java.exe":  parseCommandContextJava,
	"sudo":      parseCommandContext,
}

type CommandLine struct {
	Line        string
	ExecutePath string
	Args        []string

	Service string

	JavaArgs *JavaArgs
}

type JavaArgs struct {
	ClassName string
	JmxEnable bool
	JmxPort   int
}

func ParseCommandLine(s string) (*CommandLine, error) {
	if len(s) == 0 {
		return &CommandLine{
			Line: s,
		}, nil
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
	exe = removeFilePath(exe)

	if ext := filepath.Ext(exe); strings.ToLower(ext) == ".exe" {
		exe = strings.TrimSuffix(exe, ext)
	}

	if contextFn, ok := binsWithContext[exe]; ok {
		return contextFn(cmd[1:])
	}

	baseExe, _ := splitVersion(exe)
	if contextFn, ok := binsWithContext[exe]; ok {
		return contextFn(cmd[1:])
	}

	// trim trailing file extensions
	if i := strings.LastIndex(exe, "."); i > 0 {
		exe = exe[:i]
	}

	return &CommandLine{
		Line:    cmd,
		Service: exe,
	}, nil
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
func parseCommandContext(args []string) string {
	var prevArgIsFlag bool

	for _, a := range args {
		hasFlagPrefix, isEnvVariable := strings.HasPrefix(a, "-"), strings.ContainsRune(a, '=')
		shouldSkipArg := prevArgIsFlag || hasFlagPrefix || isEnvVariable

		if !shouldSkipArg {
			if c := trimColonRight(removeFilePath(a)); isRuneLetterAt(c, 0) {
				return c
			}
		}

		prevArgIsFlag = hasFlagPrefix
	}

	return ""
}

func parseCommandContextPython(args []string) string {
	var (
		prevArgIsFlag bool
		moduleFlag    bool
	)

	for _, a := range args {
		hasFlagPrefix, isEnvVariable := strings.HasPrefix(a, "-"), strings.ContainsRune(a, '=')

		shouldSkipArg := prevArgIsFlag || hasFlagPrefix || isEnvVariable

		if !shouldSkipArg || moduleFlag {
			if c := trimColonRight(removeFilePath(a)); isRuneLetterAt(c, 0) {
				return c
			}
		}

		if hasFlagPrefix && a == "-m" {
			moduleFlag = true
		}

		prevArgIsFlag = hasFlagPrefix
	}

	return ""
}

func parseCommandContextJava(args []string) string {
	prevArgIsFlag := false

	for _, a := range args {
		hasFlagPrefix := strings.HasPrefix(a, "-")
		includesAssignment := strings.ContainsRune(a, '=') ||
			strings.HasPrefix(a, "-X") ||
			strings.HasPrefix(a, "-javaagent:") ||
			strings.HasPrefix(a, "-verbose:")
		shouldSkipArg := prevArgIsFlag || hasFlagPrefix || includesAssignment
		if !shouldSkipArg {
			arg := removeFilePath(a)

			if arg = trimColonRight(arg); isRuneLetterAt(arg, 0) {
				if strings.HasSuffix(arg, javaJarExtension) {
					return arg[:len(arg)-len(javaJarExtension)]
				}

				if strings.HasPrefix(arg, javaApachePrefix) {
					// take the project name after the package 'org.apache.' while stripping off the remaining package
					// and class name
					arg = arg[len(javaApachePrefix):]
					if idx := strings.Index(arg, "."); idx != -1 {
						return arg[:idx]
					}
				}
				if idx := strings.LastIndex(arg, "."); idx != -1 && idx+1 < len(arg) {
					// take just the class name without the package
					return arg[idx+1:]
				}

				return arg
			}
		}

		prevArgIsFlag = hasFlagPrefix && !includesAssignment && a != javaJarFlag
	}

	return ""
}
