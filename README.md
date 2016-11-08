# jsonline
json to influxdb line protocol

This reads a line delimited input stream from stdin and writes a influxdb line
protocol compatible strings to stdout.

You can use this by piping logs with json strings to this app like

```shell
tail -f /var/log/example.logg |jsonline --name happy
```

## usage
```shell
NAME:
   jsonline - translates json objects to influxdb line protocol

USAGE:
   jsonline [global options] command [command options] [arguments...]

VERSION:
   0.1.2

COMMANDS:
     stream, s  streams from  stdin
     help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```
