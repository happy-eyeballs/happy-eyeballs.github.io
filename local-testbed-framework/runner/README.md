# Executor

This Go program can execute test case. It uploads
and runs shell scripts on several hosts according to configuration files. The
executor is capable of connecting to multiple hosts via SSH simultaneously.

## Usage

```text
Usage:
  runner [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  exec        Execute arbitrary test cases
  format      Format the output of a test run in a pretty format
  help        Help about any command

Flags:
  -h, --help                 help for runner
  -l, --log-file string      Path to the file where this program's logging output will be written to (default stdout)
  -v, --verbosity int8[=1]   Verbosity (log level) of the program's log output (default 2)

Use "runner [command] --help" for more information about a command.
```

### `exec` command

Executes test cases against iCloud Private Relay. Needs a configuration file
that contains all available hosts and how to connect to them. A sample can be
found in the YAML file [config.sample.yml](config.sample.yml).

```text
Usage:
  runner exec [flags]

Flags:
  -c, --config string            Path to the configuration file (default "config.yml")
  -h, --help                     help for exec
      --repeat-duration string   Duration for how long to repeat the selected test cases. Valid time units are "ns", "us", "ms", "s", "m", "h"
      --repeat-num int           Number of times to repeat the selected test cases (default 1)
      --repeat-spacing string    Minimum time that must elapse before the next iteration gets started. Valid time units are "ns", "us", "ms", "s", "m", "h"
  -d, --test-cases-dir string    Path to the directory storing all available test cases (default "tests")
  -t, --tests string             Comma-seperated list of test cases (default "*")
```

### `format` command

Simple tool to render the outputs of test cases in a human-readable format.

```text
Usage:
  runner format [flags]

Flags:
  -d, --artifacts-dir string   Path to the directory of the test run's artifacts
  -h, --help                   help for format
  -f, --outputs-file string    Name of the file within the artifacts directory containing the outputs of the test run (default "outputs.json")
      --show-all               Additionally show the output of the special '_all' test case
      --show-stderr            Also show the output printed to stderr
```
