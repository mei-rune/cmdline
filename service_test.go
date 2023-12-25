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
		isWindows bool
		name      string
		cmdline   string
		expected  interface{}
	}{
		{
			name:     "empty",
			cmdline:  "",
			expected: &CommandLine{},
		},
		{
			name: "single arg executable",
			cmdline: strings.Join([]string{
				"./my-server.sh",
			}, " "),
			expected: &CommandLine{
				ExecutePath: "./my-server.sh",
				Args:        []string{},
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
					Args:     []string{},
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
					Args:      []string{},
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
					Args:      []string{},
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
					Args:      []string{},
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
					Args:      []string{},
				},
			},
		},
		{
			name:    "java space in java executable path",
			cmdline: "\"/home/dd/my java dir/java\" com.dog.cat",
			expected: &CommandLine{
				ExecutePath: "/home/dd/my java dir/java",
				Args: []string{
					"com.dog.cat",
				},
				Java: &JavaArgs{
					ClassName: "com.dog.cat",
					Args:      []string{},
				},
			},
		},

		{
			isWindows: true,
			name:      "windows java and -Dcom.sun.management.jmxremote",
			cmdline:   "D:\\data\\hengwei_dev\\runtime_env\\jre\\bin\\java.exe -Xmx4096m -cp D:\\data\\hengwei_dev\\lib\\commons\\EasyXls-1.1.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\HikariCP-3.4.5.jar;D:\\data\\hengwei_dev\\lib\\commons\\JavaEWAH-0.7.9.jar;D:\\data\\hengwei_dev\\lib\\commons\\SparseBitSet-1.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\accessors-smart-1.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\activation-1.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\activemq-client-5.13.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\aopalliance-1.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\aopalliance-repackaged-2.4.0-b31.jar;D:\\data\\hengwei_dev\\lib\\commons\\apache-mime4j-0.6.jar;D:\\data\\hengwei_dev\\lib\\commons\\argparse4j-0.4.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\asm-4.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\asm-tree-4.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\asm-util-4.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-all-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-anim-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-awt-util-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-bridge-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-codec-1.14.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-constants-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-css-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-dom-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-ext-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-extension-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-gui-util-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-gvt-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-i18n-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-parser-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-rasterizer-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-rasterizer-ext-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-script-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-shared-resources-1.14.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-slideshow-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-squiggle-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-squiggle-ext-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-svg-dom-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-svgbrowser-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-svggen-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-svgpp-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-svgrasterizer-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-swing-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-transcoder-1.14.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-ttf2svg-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-util-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-xml-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\bcpkix-jdk15on-1.68.jar;D:\\data\\hengwei_dev\\lib\\commons\\bcprov-jdk15on-1.68.jar;D:\\data\\hengwei_dev\\lib\\commons\\bcprov-jdk16-1.46.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-beanutils-1.9.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-codec-1.6.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-collections-3.2.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-collections4-4.4.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-compress-1.20.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-csv-1.8.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-io-2.11.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-jexl-2.1.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-lang-2.6.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-lang3-3.3.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-logging-1.1.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-math3-3.6.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-pool2-2.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\curvesapi-1.06.jar;D:\\data\\hengwei_dev\\lib\\commons\\easyexcel-3.1.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\easyexcel-core-3.1.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\easyexcel-support-3.1.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\ehcache-3.9.9.jar;D:\\data\\hengwei_dev\\lib\\commons\\extreme-client-3.8.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\extreme-commons-3.8.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\extreme-model-3.8.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\extreme-report-3.8.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\extreme-rest-3.8.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\extreme-share-3.8.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\fastjson-1.2.83.jar;D:\\data\\hengwei_dev\\lib\\commons\\flyway-core-6.3.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\fontbox-2.0.22.jar;D:\\data\\hengwei_dev\\lib\\commons\\fr.opensagres.poi.xwpf.converter.core-2.0.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\fr.opensagres.poi.xwpf.converter.pdf-2.0.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\fr.opensagres.xdocreport.itext.extension-2.0.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\freemarker-2.3.30.jar;D:\\data\\hengwei_dev\\lib\\commons\\geronimo-j2ee-management_1.1_spec-1.0.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\geronimo-jms_1.1_spec-1.1.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\graphics2d-0.30.jar;D:\\data\\hengwei_dev\\lib\\commons\\grizzly-framework-2.3.23.jar;D:\\data\\hengwei_dev\\lib\\commons\\grizzly-http-2.3.23.jar;D:\\data\\hengwei_dev\\lib\\commons\\grizzly-http-server-2.3.23.jar;D:\\data\\hengwei_dev\\lib\\commons\\gson-2.3.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\guava-19.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\guice-3.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\guice-multibindings-3.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\hamcrest-core-1.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\hawtbuf-1.11.jar;D:\\data\\hengwei_dev\\lib\\commons\\hk2-api-2.4.0-b31.jar;D:\\data\\hengwei_dev\\lib\\commons\\hk2-locator-2.4.0-b31.jar;D:\\data\\hengwei_dev\\lib\\commons\\hk2-utils-2.4.0-b31.jar;D:\\data\\hengwei_dev\\lib\\commons\\httpclient-4.3.6.jar;D:\\data\\hengwei_dev\\lib\\commons\\httpcore-4.3.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\influxdb-java-2.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\itext-2.1.7.jar;D:\\data\\hengwei_dev\\lib\\commons\\jackson-annotations-2.9.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\jackson-core-2.9.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\jackson-core-asl-1.9.12.jar;D:\\data\\hengwei_dev\\lib\\commons\\jackson-databind-2.9.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\jackson-jaxrs-1.9.12.jar;D:\\data\\hengwei_dev\\lib\\commons\\jackson-mapper-asl-1.9.12.jar;D:\\data\\hengwei_dev\\lib\\commons\\jackson-xc-1.9.12.jar;D:\\data\\hengwei_dev\\lib\\commons\\java-jwt-3.3.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\javacsv-2.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\javassist-3.12.1.GA.jar;D:\\data\\hengwei_dev\\lib\\commons\\javassist-3.18.1-GA.jar;D:\\data\\hengwei_dev\\lib\\commons\\javax.annotation-api-1.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\javax.inject-1.jar;D:\\data\\hengwei_dev\\lib\\commons\\javax.inject-2.4.0-b31.jar;D:\\data\\hengwei_dev\\lib\\commons\\javax.ws.rs-api-2.0.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\jaxb-impl-2.2.5-2.jar;D:\\data\\hengwei_dev\\lib\\commons\\jaxrs-api-3.0.2.Final.jar;D:\\data\\hengwei_dev\\lib\\commons\\jcip-annotations-1.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\jcl-over-slf4j-1.7.30.jar;D:\\data\\hengwei_dev\\lib\\commons\\jedis-2.6.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\jersey-client-2.22.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\jersey-common-2.22.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\jersey-container-grizzly2-http-2.22.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\jersey-guava-2.22.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\jersey-media-jaxb-2.22.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\jersey-server-2.22.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\jmockit-1.7.jar;D:\\data\\hengwei_dev\\lib\\commons\\jsch-0.1.50.jar;D:\\data\\hengwei_dev\\lib\\commons\\json-smart-2.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\jsqlparser-1.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\jsr250-api-1.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\jtds-1.3.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\junit-4.11.jar;D:\\data\\hengwei_dev\\lib\\commons\\jxl-2.6.12.jar;D:\\data\\hengwei_dev\\lib\\commons\\logback-classic-1.2.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\logback-core-1.2.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\mail-1.4.4.jar;D:\\data\\hengwei_dev\\lib\\commons\\mibble-parser-2.9.3.fix17.jar;D:\\data\\hengwei_dev\\lib\\commons\\mybatis-3.2.8.jar;D:\\data\\hengwei_dev\\lib\\commons\\mybatis-guice-3.6.jar;D:\\data\\hengwei_dev\\lib\\commons\\netty-3.6.4.Final.jar;D:\\data\\hengwei_dev\\lib\\commons\\okhttp-2.4.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\okio-1.4.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\org.eclipse.jgit-3.4.1.201406201815-r.jar;D:\\data\\hengwei_dev\\lib\\commons\\org.eclipse.jgit.http.server-3.4.1.201406201815-r.jar;D:\\data\\hengwei_dev\\lib\\commons\\org.eclipse.jgit.junit-3.4.1.201406201815-r.jar;D:\\data\\hengwei_dev\\lib\\commons\\org.eclipse.jgit.ui-3.4.1.201406201815-r.jar;D:\\data\\hengwei_dev\\lib\\commons\\osgi-resource-locator-1.0.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\pagehelper-5.1.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\pdfbox-2.0.22.jar;D:\\data\\hengwei_dev\\lib\\commons\\pdfbox-app-2.0.25.jar;D:\\data\\hengwei_dev\\lib\\commons\\pinyin4j-2.6.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\poi-5.0.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\poi-ooxml-5.0.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\poi-ooxml-full-5.2.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\poi-ooxml-lite-5.0.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\poi-ooxml-schemas-4.1.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\poi-ooxml-schemas-extra-5.1.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\poi-scratchpad-5.0.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\poi-tl-1.11.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\postgresql-42.2.18.jre7.jar;D:\\data\\hengwei_dev\\lib\\commons\\resteasy-guice-3.0.2.Final.jar;D:\\data\\hengwei_dev\\lib\\commons\\resteasy-jackson-provider-3.0.2.Final.jar;D:\\data\\hengwei_dev\\lib\\commons\\resteasy-jaxb-provider-3.0.2.Final.jar;D:\\data\\hengwei_dev\\lib\\commons\\resteasy-jaxrs-3.0.2.Final.jar;D:\\data\\hengwei_dev\\lib\\commons\\resteasy-multipart-provider-3.0.2.Final.jar;D:\\data\\hengwei_dev\\lib\\commons\\resteasy-netty-3.0.2.Final.jar;D:\\data\\hengwei_dev\\lib\\commons\\retrofit-1.9.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\scannotation-1.0.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\screw-core-1.0.5.jar;D:\\data\\hengwei_dev\\lib\\commons\\serializer-2.7.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\servlet-api-2.5.jar;D:\\data\\hengwei_dev\\lib\\commons\\slf4j-api-1.7.5.jar;D:\\data\\hengwei_dev\\lib\\commons\\snmp4j-1.10.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\stax2-api-4.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\syslog4j-0.9.30.jar;D:\\data\\hengwei_dev\\lib\\commons\\validation-api-1.1.0.Final.jar;D:\\data\\hengwei_dev\\lib\\commons\\woodstox-core-5.2.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\xalan-2.7.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\xml-apis-1.4.01.jar;D:\\data\\hengwei_dev\\lib\\commons\\xml-apis-ext-1.3.04.jar;D:\\data\\hengwei_dev\\lib\\commons\\xmlbeans-4.0.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\xmlgraphics-commons-2.4.jar;D:\\data\\hengwei_dev\\lib\\commons\\xmlsec-2.2.1.jar;D:\\data\\hengwei_dev\\lib\\server_biz\\commons-lang-2.6.jar;D:\\data\\hengwei_dev\\lib\\server_biz\\extreme-biz-3.8.1.jar;D:\\data\\hengwei_dev\\lib\\server_biz\\extreme-migration-3.8.1.jar;D:\\data\\hengwei_dev\\lib\\server_biz\\hsqldb.jar;D:\\data\\hengwei_dev\\lib\\server_biz\\jackcess-2.1.0.jar;D:\\data\\hengwei_dev\\lib\\server_biz\\ucanaccess-2.0.9.5.jar -Dcom.sun.management.jmxremote -Dcom.sun.management.jmxremote.port=0 -Dcom.sun.management.jmxremote.authenticate=false -Dcom.sun.management.jmxremote.ssl=false -Dcom.sun.management.jmxremote.local.only=false -Dconf=D:\\data\\hengwei_dev/conf/global.properties com.tpt.nm.Server",
			expected: &CommandLine{
				ExecutePath: "D:\\data\\hengwei_dev\\runtime_env\\jre\\bin\\java.exe",
				Args: []string{
					"-Xmx4096m",
					"-cp",
					"D:\\data\\hengwei_dev\\lib\\commons\\EasyXls-1.1.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\HikariCP-3.4.5.jar;D:\\data\\hengwei_dev\\lib\\commons\\JavaEWAH-0.7.9.jar;D:\\data\\hengwei_dev\\lib\\commons\\SparseBitSet-1.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\accessors-smart-1.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\activation-1.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\activemq-client-5.13.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\aopalliance-1.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\aopalliance-repackaged-2.4.0-b31.jar;D:\\data\\hengwei_dev\\lib\\commons\\apache-mime4j-0.6.jar;D:\\data\\hengwei_dev\\lib\\commons\\argparse4j-0.4.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\asm-4.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\asm-tree-4.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\asm-util-4.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-all-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-anim-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-awt-util-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-bridge-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-codec-1.14.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-constants-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-css-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-dom-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-ext-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-extension-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-gui-util-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-gvt-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-i18n-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-parser-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-rasterizer-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-rasterizer-ext-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-script-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-shared-resources-1.14.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-slideshow-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-squiggle-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-squiggle-ext-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-svg-dom-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-svgbrowser-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-svggen-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-svgpp-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-svgrasterizer-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-swing-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-transcoder-1.14.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-ttf2svg-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-util-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\batik-xml-1.13.jar;D:\\data\\hengwei_dev\\lib\\commons\\bcpkix-jdk15on-1.68.jar;D:\\data\\hengwei_dev\\lib\\commons\\bcprov-jdk15on-1.68.jar;D:\\data\\hengwei_dev\\lib\\commons\\bcprov-jdk16-1.46.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-beanutils-1.9.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-codec-1.6.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-collections-3.2.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-collections4-4.4.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-compress-1.20.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-csv-1.8.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-io-2.11.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-jexl-2.1.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-lang-2.6.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-lang3-3.3.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-logging-1.1.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-math3-3.6.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\commons-pool2-2.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\curvesapi-1.06.jar;D:\\data\\hengwei_dev\\lib\\commons\\easyexcel-3.1.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\easyexcel-core-3.1.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\easyexcel-support-3.1.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\ehcache-3.9.9.jar;D:\\data\\hengwei_dev\\lib\\commons\\extreme-client-3.8.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\extreme-commons-3.8.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\extreme-model-3.8.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\extreme-report-3.8.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\extreme-rest-3.8.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\extreme-share-3.8.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\fastjson-1.2.83.jar;D:\\data\\hengwei_dev\\lib\\commons\\flyway-core-6.3.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\fontbox-2.0.22.jar;D:\\data\\hengwei_dev\\lib\\commons\\fr.opensagres.poi.xwpf.converter.core-2.0.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\fr.opensagres.poi.xwpf.converter.pdf-2.0.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\fr.opensagres.xdocreport.itext.extension-2.0.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\freemarker-2.3.30.jar;D:\\data\\hengwei_dev\\lib\\commons\\geronimo-j2ee-management_1.1_spec-1.0.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\geronimo-jms_1.1_spec-1.1.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\graphics2d-0.30.jar;D:\\data\\hengwei_dev\\lib\\commons\\grizzly-framework-2.3.23.jar;D:\\data\\hengwei_dev\\lib\\commons\\grizzly-http-2.3.23.jar;D:\\data\\hengwei_dev\\lib\\commons\\grizzly-http-server-2.3.23.jar;D:\\data\\hengwei_dev\\lib\\commons\\gson-2.3.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\guava-19.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\guice-3.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\guice-multibindings-3.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\hamcrest-core-1.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\hawtbuf-1.11.jar;D:\\data\\hengwei_dev\\lib\\commons\\hk2-api-2.4.0-b31.jar;D:\\data\\hengwei_dev\\lib\\commons\\hk2-locator-2.4.0-b31.jar;D:\\data\\hengwei_dev\\lib\\commons\\hk2-utils-2.4.0-b31.jar;D:\\data\\hengwei_dev\\lib\\commons\\httpclient-4.3.6.jar;D:\\data\\hengwei_dev\\lib\\commons\\httpcore-4.3.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\influxdb-java-2.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\itext-2.1.7.jar;D:\\data\\hengwei_dev\\lib\\commons\\jackson-annotations-2.9.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\jackson-core-2.9.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\jackson-core-asl-1.9.12.jar;D:\\data\\hengwei_dev\\lib\\commons\\jackson-databind-2.9.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\jackson-jaxrs-1.9.12.jar;D:\\data\\hengwei_dev\\lib\\commons\\jackson-mapper-asl-1.9.12.jar;D:\\data\\hengwei_dev\\lib\\commons\\jackson-xc-1.9.12.jar;D:\\data\\hengwei_dev\\lib\\commons\\java-jwt-3.3.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\javacsv-2.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\javassist-3.12.1.GA.jar;D:\\data\\hengwei_dev\\lib\\commons\\javassist-3.18.1-GA.jar;D:\\data\\hengwei_dev\\lib\\commons\\javax.annotation-api-1.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\javax.inject-1.jar;D:\\data\\hengwei_dev\\lib\\commons\\javax.inject-2.4.0-b31.jar;D:\\data\\hengwei_dev\\lib\\commons\\javax.ws.rs-api-2.0.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\jaxb-impl-2.2.5-2.jar;D:\\data\\hengwei_dev\\lib\\commons\\jaxrs-api-3.0.2.Final.jar;D:\\data\\hengwei_dev\\lib\\commons\\jcip-annotations-1.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\jcl-over-slf4j-1.7.30.jar;D:\\data\\hengwei_dev\\lib\\commons\\jedis-2.6.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\jersey-client-2.22.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\jersey-common-2.22.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\jersey-container-grizzly2-http-2.22.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\jersey-guava-2.22.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\jersey-media-jaxb-2.22.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\jersey-server-2.22.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\jmockit-1.7.jar;D:\\data\\hengwei_dev\\lib\\commons\\jsch-0.1.50.jar;D:\\data\\hengwei_dev\\lib\\commons\\json-smart-2.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\jsqlparser-1.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\jsr250-api-1.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\jtds-1.3.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\junit-4.11.jar;D:\\data\\hengwei_dev\\lib\\commons\\jxl-2.6.12.jar;D:\\data\\hengwei_dev\\lib\\commons\\logback-classic-1.2.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\logback-core-1.2.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\mail-1.4.4.jar;D:\\data\\hengwei_dev\\lib\\commons\\mibble-parser-2.9.3.fix17.jar;D:\\data\\hengwei_dev\\lib\\commons\\mybatis-3.2.8.jar;D:\\data\\hengwei_dev\\lib\\commons\\mybatis-guice-3.6.jar;D:\\data\\hengwei_dev\\lib\\commons\\netty-3.6.4.Final.jar;D:\\data\\hengwei_dev\\lib\\commons\\okhttp-2.4.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\okio-1.4.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\org.eclipse.jgit-3.4.1.201406201815-r.jar;D:\\data\\hengwei_dev\\lib\\commons\\org.eclipse.jgit.http.server-3.4.1.201406201815-r.jar;D:\\data\\hengwei_dev\\lib\\commons\\org.eclipse.jgit.junit-3.4.1.201406201815-r.jar;D:\\data\\hengwei_dev\\lib\\commons\\org.eclipse.jgit.ui-3.4.1.201406201815-r.jar;D:\\data\\hengwei_dev\\lib\\commons\\osgi-resource-locator-1.0.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\pagehelper-5.1.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\pdfbox-2.0.22.jar;D:\\data\\hengwei_dev\\lib\\commons\\pdfbox-app-2.0.25.jar;D:\\data\\hengwei_dev\\lib\\commons\\pinyin4j-2.6.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\poi-5.0.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\poi-ooxml-5.0.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\poi-ooxml-full-5.2.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\poi-ooxml-lite-5.0.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\poi-ooxml-schemas-4.1.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\poi-ooxml-schemas-extra-5.1.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\poi-scratchpad-5.0.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\poi-tl-1.11.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\postgresql-42.2.18.jre7.jar;D:\\data\\hengwei_dev\\lib\\commons\\resteasy-guice-3.0.2.Final.jar;D:\\data\\hengwei_dev\\lib\\commons\\resteasy-jackson-provider-3.0.2.Final.jar;D:\\data\\hengwei_dev\\lib\\commons\\resteasy-jaxb-provider-3.0.2.Final.jar;D:\\data\\hengwei_dev\\lib\\commons\\resteasy-jaxrs-3.0.2.Final.jar;D:\\data\\hengwei_dev\\lib\\commons\\resteasy-multipart-provider-3.0.2.Final.jar;D:\\data\\hengwei_dev\\lib\\commons\\resteasy-netty-3.0.2.Final.jar;D:\\data\\hengwei_dev\\lib\\commons\\retrofit-1.9.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\scannotation-1.0.3.jar;D:\\data\\hengwei_dev\\lib\\commons\\screw-core-1.0.5.jar;D:\\data\\hengwei_dev\\lib\\commons\\serializer-2.7.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\servlet-api-2.5.jar;D:\\data\\hengwei_dev\\lib\\commons\\slf4j-api-1.7.5.jar;D:\\data\\hengwei_dev\\lib\\commons\\snmp4j-1.10.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\stax2-api-4.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\syslog4j-0.9.30.jar;D:\\data\\hengwei_dev\\lib\\commons\\validation-api-1.1.0.Final.jar;D:\\data\\hengwei_dev\\lib\\commons\\woodstox-core-5.2.1.jar;D:\\data\\hengwei_dev\\lib\\commons\\xalan-2.7.2.jar;D:\\data\\hengwei_dev\\lib\\commons\\xml-apis-1.4.01.jar;D:\\data\\hengwei_dev\\lib\\commons\\xml-apis-ext-1.3.04.jar;D:\\data\\hengwei_dev\\lib\\commons\\xmlbeans-4.0.0.jar;D:\\data\\hengwei_dev\\lib\\commons\\xmlgraphics-commons-2.4.jar;D:\\data\\hengwei_dev\\lib\\commons\\xmlsec-2.2.1.jar;D:\\data\\hengwei_dev\\lib\\server_biz\\commons-lang-2.6.jar;D:\\data\\hengwei_dev\\lib\\server_biz\\extreme-biz-3.8.1.jar;D:\\data\\hengwei_dev\\lib\\server_biz\\extreme-migration-3.8.1.jar;D:\\data\\hengwei_dev\\lib\\server_biz\\hsqldb.jar;D:\\data\\hengwei_dev\\lib\\server_biz\\jackcess-2.1.0.jar;D:\\data\\hengwei_dev\\lib\\server_biz\\ucanaccess-2.0.9.5.jar",
					"-Dcom.sun.management.jmxremote",
					"-Dcom.sun.management.jmxremote.port=0",
					"-Dcom.sun.management.jmxremote.authenticate=false",
					"-Dcom.sun.management.jmxremote.ssl=false",
					"-Dcom.sun.management.jmxremote.local.only=false",
					"-Dconf=D:\\data\\hengwei_dev/conf/global.properties",
					"com.tpt.nm.Server",
				},
				Java: &JavaArgs{
					ClassName: "com.tpt.nm.Server",
					Args:      []string{},
					JmxEnable: true,
					JmxPort:   "0",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			command, err := ParseCommandLine(tt.isWindows, tt.cmdline)
			if err != nil {
				t.Error(err)
				return
			}

			assert.Equal(t, tt.expected, command)
		})
	}
}
