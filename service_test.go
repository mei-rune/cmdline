package cmdline

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitVersion(t *testing.T) {
	tests := []struct {
		name            string
		cmdline         string
		expectedexe     string
		expectedversion string
	}{
		{
			name:    "blank",
			cmdline: "",
		},
		{
			name:            "python2.7",
			cmdline:         "python2.7",
			expectedexe:     "python",
			expectedversion: "2.7",
		},
		{
			name:            "python2",
			cmdline:         "python2",
			expectedexe:     "python",
			expectedversion: "2",
		},
		{
			name:            "ruby2.3",
			cmdline:         "ruby2.3",
			expectedexe:     "ruby",
			expectedversion: "2.3",
		},
		{
			name:            "ruby2",
			cmdline:         "ruby2",
			expectedexe:     "ruby",
			expectedversion: "2",
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
		name     string
		cmdline  string
		expected interface{}
	}{
		{
			name:     "empty",
			cmdline:  "",
			expected:  &CommandLine{},
		},
		{
			name: "single arg executable",
			cmdline: strings.Join([]string{
				"./my-server.sh",
			}, " "),
			expected: &CommandLine{
				ExecutePath: "./my-server.sh",
				Args: []string{},
			},
		},
		{
			name: "sudo",
			cmdline: strings.Join([]string{
				"sudo", "-E", "-u", "dog", "/usr/local/bin/myApp", "-items=0,1,2,3", "-foo=bar",
			}, " "),
			expected: &CommandLine{
				ExecutePath: "sudo",
				Args: []string{
					"-E", "-u", "dog", "/usr/local/bin/myApp", "-items=0,1,2,3", "-foo=bar",
				},
				Sub: &SubCommand{
					Command: "/usr/local/bin/myApp",
					Args: []string{
						 "-items=0,1,2,3", "-foo=bar",
						},
				},
			},
		},
		{
			name: "python flask argument",
			cmdline: strings.Join([]string{
				"/opt/python/2.7.11/bin/python2.7", "flask", "run", "--host=0.0.0.0",
			}, " "),
			expected: &CommandLine{
				ExecutePath: "/opt/python/2.7.11/bin/python2.7",
				Args: []string{
					"flask", "run", "--host=0.0.0.0",
				},
				Python: &PythonArgs{
					FilePath: "flask",
					Args: []string{
						 "run", "--host=0.0.0.0",
						},
				},
			},
		},
		{
			name: "python - flask argument in path",
			cmdline: strings.Join([]string{
				"/opt/python/2.7.11/bin/python2.7", "/opt/dogweb/bin/flask", "run", "--host=0.0.0.0", "--without-threads",
			}, " "),
			expected: &CommandLine{
				ExecutePath: "/opt/python/2.7.11/bin/python2.7",
				Args: []string{
					"/opt/dogweb/bin/flask", "run", "--host=0.0.0.0", "--without-threads",
				},
				Python: &PythonArgs{
					FilePath: "/opt/dogweb/bin/flask",
					Args: []string{
						 "run", "--host=0.0.0.0", "--without-threads",
						},
				},
			},
		},
		{
			name: "python - module hello",
			cmdline: strings.Join([]string{
				"python3", "-m", "hello",
			}, " "),
			expected: &CommandLine{
				ExecutePath: "python3",
				Args: []string{
					"-m", "hello",
				},
				Python: &PythonArgs{
					FilePath: "hello",
					Args: []string{},
				},
			},
		},
		{
			name: "ruby - td-agent",
			cmdline: strings.Join([]string{
				"ruby", "/usr/sbin/td-agent", "--log", "/var/log/td-agent/td-agent.log", "--daemon", "/var/run/td-agent/td-agent.pid",
			}, " "),
			expected: &CommandLine{
				ExecutePath: "ruby",
				Args: []string{
					 "/usr/sbin/td-agent", "--log", "/var/log/td-agent/td-agent.log", "--daemon", "/var/run/td-agent/td-agent.pid",
				},
				Ruby: &RubyArgs{
					FilePath: "/usr/sbin/td-agent",
					Args: []string{
						"--log", "/var/log/td-agent/td-agent.log", "--daemon", "/var/run/td-agent/td-agent.pid",
					},
				},
			},
		},
		{
			name: "java using the -jar flag to define the service",
			cmdline: strings.Join([]string{
				"java", "-Xmx4000m", "-Xms4000m", "-XX:ReservedCodeCacheSize=256m", "-jar", "/opt/sheepdog/bin/myservice.jar",
			}, " "),

			expected: &CommandLine{
				ExecutePath: "java",
				Args: []string{
					"-Xmx4000m", "-Xms4000m", "-XX:ReservedCodeCacheSize=256m", "-jar", "/opt/sheepdog/bin/myservice.jar",
				},
				Java: &JavaArgs{
					ClassName: "/opt/sheepdog/bin/myservice.jar",
					Args: []string{},
				},
			},
		},
		{
			name: "java class name as service",
			cmdline: strings.Join([]string{
				"java", "-Xmx4000m", "-Xms4000m", "-XX:ReservedCodeCacheSize=256m", "com.datadog.example.HelloWorld",
			}, " "),
			expected: &CommandLine{
				ExecutePath: "java",
				Args: []string{
					"-Xmx4000m", "-Xms4000m", "-XX:ReservedCodeCacheSize=256m", "com.datadog.example.HelloWorld",
				},
				Java: &JavaArgs{
					ClassName: "com.datadog.example.HelloWorld",
					Args: []string{},
				},
			},
		},
		{
			name: "java kafka",
			cmdline: strings.Join([]string{
				"java", "-Xmx4000m", "-Xms4000m", "-XX:ReservedCodeCacheSize=256m", "kafka.Kafka",
			}, " "),
			expected: &CommandLine{
				ExecutePath: "java",
				Args: []string{
					"-Xmx4000m", "-Xms4000m", "-XX:ReservedCodeCacheSize=256m", "kafka.Kafka",
				},
				Java: &JavaArgs{
					ClassName: "kafka.Kafka",
					Args: []string{},
				},
			},
		},
		{
			name: "java parsing for org.apache projects with cassandra as the service",
			cmdline: strings.Join([]string{
				"/usr/bin/java", "-Xloggc:/usr/share/cassandra/logs/gc.log", "-ea", "-XX:+HeapDumpOnOutOfMemoryError", "-Xss256k", "-Dlogback.configurationFile=logback.xml",
				"-Dcassandra.logdir=/var/log/cassandra", "-Dcassandra.storagedir=/data/cassandra",
				"-cp", "/etc/cassandra:/usr/share/cassandra/lib/HdrHistogram-2.1.9.jar:/usr/share/cassandra/lib/cassandra-driver-core-3.0.1-shaded.jar",
				"org.apache.cassandra.service.CassandraDaemon",
			}, " "),
			expected: &CommandLine{
				ExecutePath: "/usr/bin/java",
				Args: []string{
					"-Xloggc:/usr/share/cassandra/logs/gc.log", "-ea", "-XX:+HeapDumpOnOutOfMemoryError", "-Xss256k", "-Dlogback.configurationFile=logback.xml",
				"-Dcassandra.logdir=/var/log/cassandra", "-Dcassandra.storagedir=/data/cassandra",
				"-cp", "/etc/cassandra:/usr/share/cassandra/lib/HdrHistogram-2.1.9.jar:/usr/share/cassandra/lib/cassandra-driver-core-3.0.1-shaded.jar",
				"org.apache.cassandra.service.CassandraDaemon",
				},
				Java: &JavaArgs{
					ClassName: "org.apache.cassandra.service.CassandraDaemon",
					Args: []string{},
				},
			},
		},
		{
			name:     "java space in java executable path",
			cmdline:  "\"/home/dd/my java dir/java\" com.dog.cat",
			expected: &CommandLine{
				ExecutePath: "/home/dd/my java dir/java",
				Args: []string{
					"com.dog.cat",
				},
				Java: &JavaArgs{
					ClassName: "com.dog.cat",
					Args: []string{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			command, err := ParseCommandLine(tt.cmdline)
			if err != nil {
				t.Error(err)
				return
			}

			assert.Equal(t, tt.expected, command)
		})
	}
}
