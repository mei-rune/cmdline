// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitVersion(t *testing.T) {
	tests := []struct {
		name               string
		cmdline            string
		expectedexe string
		expectedversion string
	}{
		{
			name:               "blank",
			cmdline:            "",
		},
		{
			name:               "python2.7",
			cmdline:           "python2.7",
			expectedexe: "python",
			expectedversion: "2.7",
		},
		{
			name:               "python2",
			cmdline:           "python2",
			expectedexe: "python",
			expectedversion: "2",
		},
		{
			name:               "ruby2.3",
			cmdline:           "ruby2.3",
			expectedexe: "ruby",
			expectedversion: "2.3",
		},
		{
			name:               "ruby2",
			cmdline:            "ruby2",
			expectedexe:        "ruby",
			expectedversion:    "2",
		},
	}

	for _, tt := range tests {
	  exe, version := splitVersion(tt.cmdline)
	  if tt.expectedexe != exe {
	   	t.Error("[exe] want", tt.expectedexe, "got", exe)
	  }
	  if tt.expectedversion != version {
	   	t.Error("[version] want", tt.expectedversion, "got", version)
	  }
	}
}

func TestExtractServiceMetadata(t *testing.T) {
	tests := []struct {
		name               string
		cmdline            []string
		expected string
	}{
		{
			name:               "empty",
			cmdline:            []string{},
			expected: "",
		},
		{
			name:               "blank",
			cmdline:            []string{""},
			expected: "",
		},
		{
			name: "single arg executable",
			cmdline: []string{
				"./my-server.sh",
			},
			expected: "process_context:my-server",
		},
		{
			name: "sudo",
			cmdline: []string{
				"sudo", "-E", "-u", "dog", "/usr/local/bin/myApp", "-items=0,1,2,3", "-foo=bar",
			},
			expected: "process_context:myApp",
		},
		{
			name: "python flask argument",
			cmdline: []string{
				"/opt/python/2.7.11/bin/python2.7", "flask", "run", "--host=0.0.0.0",
			},
			expected: "process_context:flask",
		},
		{
			name: "python - flask argument in path",
			cmdline: []string{
				"/opt/python/2.7.11/bin/python2.7", "/opt/dogweb/bin/flask", "run", "--host=0.0.0.0", "--without-threads",
			},
			expected: "process_context:flask",
		},
		{
			name: "python flask in single argument",
			cmdline: []string{
				"/opt/python/2.7.11/bin/python2.7 flask run --host=0.0.0.0",
			},
			expected: "process_context:flask",
		},
		{
			name: "python - module hello",
			cmdline: []string{
				"python3", "-m", "hello",
			},
			expected: "process_context:hello",
		},
		{
			name: "ruby - td-agent",
			cmdline: []string{
				"ruby", "/usr/sbin/td-agent", "--log", "/var/log/td-agent/td-agent.log", "--daemon", "/var/run/td-agent/td-agent.pid",
			},
			expected: "process_context:td-agent",
		},
		{
			name: "java using the -jar flag to define the service",
			cmdline: []string{
				"java", "-Xmx4000m", "-Xms4000m", "-XX:ReservedCodeCacheSize=256m", "-jar", "/opt/sheepdog/bin/myservice.jar",
			},
			expected: "process_context:myservice",
		},
		{
			name: "java class name as service",
			cmdline: []string{
				"java", "-Xmx4000m", "-Xms4000m", "-XX:ReservedCodeCacheSize=256m", "com.datadog.example.HelloWorld",
			},
			expected: "process_context:HelloWorld",
		},
		{
			name: "java kafka",
			cmdline: []string{
				"java", "-Xmx4000m", "-Xms4000m", "-XX:ReservedCodeCacheSize=256m", "kafka.Kafka",
			},
			expected: "process_context:Kafka",
		},
		{
			name: "java parsing for org.apache projects with cassandra as the service",
			cmdline: []string{
				"/usr/bin/java", "-Xloggc:/usr/share/cassandra/logs/gc.log", "-ea", "-XX:+HeapDumpOnOutOfMemoryError", "-Xss256k", "-Dlogback.configurationFile=logback.xml",
				"-Dcassandra.logdir=/var/log/cassandra", "-Dcassandra.storagedir=/data/cassandra",
				"-cp", "/etc/cassandra:/usr/share/cassandra/lib/HdrHistogram-2.1.9.jar:/usr/share/cassandra/lib/cassandra-driver-core-3.0.1-shaded.jar",
				"org.apache.cassandra.service.CassandraDaemon",
			},
			expected: "process_context:cassandra",
		},
		{
			name: "java space in java executable path",
			cmdline: []string{
				"/home/dd/my java dir/java", "com.dog.cat",
			},
			expected: "process_context:cat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			command, err := ParseCommandLine(tt.cmdline)
			if err != nil {
				t.Error(err)
				return
			}

			assert.Equal(t, tt.expected, command.Service)
		})
	}
}
