[![Go Report card](https://goreportcard.com/badge/github.com/wolphin-project/prometheus-export)](https://goreportcard.com/report/github.com/wolphin-project/prometheus-export)

# Prometheus v2 time series exporter

This tool is designed to extract time series data from a Prometheus (v2) database
without going through the REST API.

## Query format

Querying is done through a minimum and maximum time (``--mint``, ``--maxt``, RFC 3339 datetimes),
and metric names (``--name``). Using the ``--select`` or ``-S`` flags which receive
``label:value`` pairs as arguments only return the records matching those values.
Further filtering is done after the database query through the ``-I`` flags to
keep only select labels (``*``, the default, catches all labels). The ``-E`` flags exclude
labels you know you don't want when using wildcard include.

## Output

Both ``json`` and ``csv`` are available as output formats. JSON output is done through
[JSON Lines](http://jsonlines.org/), with one JSON object per line. CSV output is done as
a standard CSV file, with a header for each included label (CSV output does not work with
wildcard label inclusion because it requires a fixed column format), with the three first
columns being ``__name__``, ``__time__`` and ``__value__``.

Script-related info/error/warning messages happen on stderr.

## Example

### JSON

```bash
$ ./prometheus-export --name container_cpu_usage_seconds_total -I id -f json
{"__name__":"container_cpu_usage_seconds_total","__time__":"1508254713272","__value__":"12.018311","id":"/docker/c0d2c0ebdee961957d2859d7a9909262cec9a8dd8001108fded93ff9e97144ae"}
{"__name__":"container_cpu_usage_seconds_total","__time__":"1508255433272","__value__":"12.829795","id":"/docker/c0d2c0ebdee961957d2859d7a9909262cec9a8dd8001108fded93ff9e97144ae"}
[…]
```

### CSV


```bash
$ ./prometheus-export --name container_cpu_usage_seconds_total -I id -f csv
__name__,__time__,__value__,id
container_cpu_usage_seconds_total,1508254713272,12.018311,/docker/c0d2c0ebdee961957d2859d7a9909262cec9a8dd8001108fded93ff9e97144ae
container_cpu_usage_seconds_total,1508255433272,12.829795,/docker/c0d2c0ebdee961957d2859d7a9909262cec9a8dd8001108fded93ff9e97144ae
[…]
```

### Filtering


```bash
$ ./prometheus-export --name container_cpu_usage_seconds_total -I id -S "image:google/cadvisor:v0.27.1"
{"__name__":"container_cpu_usage_seconds_total","__time__":"1507883805647","__value__":"0.928787","id":"/docker/05ea8b0a1283d9bf75e815ae3c0c0472db71e2dd13ae7ed9cfe6386fad6f8e58"}
{"__name__":"container_cpu_usage_seconds_total","__time__":"1507884645649","__value__":"3.301307","id":"/docker/05ea8b0a1283d9bf75e815ae3c0c0472db71e2dd13ae7ed9cfe6386fad6f8e58"}
[…]
```

The JSON output can be fed to [jq](https://stedolan.github.io/jq/) or other tools which understand more complex expressions.

## Usage

```
NAME:
   prometheus-export - Prometheus v2 DB exporter

USAGE:
   prometheus-export [global options] command [command options] [arguments...]

VERSION:
   1.0

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --format value, -f value         export format; one of csv, json (default: "json")
   --mint value, --min value        lower bound of the observed time period (default: "1970-01-01T00:00:00Z")
   --maxt value, --max value        upper bound of the observed time period (default: "2050-01-01T00:00:00Z")
   --name value, -N value           time series names to export
   --select value, -S value         additional filters with -S "label:value"
   --path value, -p value           path to the directory of the Prometheus Database (default: "/data/prometheus")
   --out value, -o value            output file path (- for stdout) (default: "-")
   --include-label value, -I value  labels to include in the export (series without those labels will not be kept)
   --exclude-label value, -E value  labels to exclude from the export
   --help, -h                       show help
   --version, -v                    print the version
```


## License

Apache License 2.0, see [LICENSE](./LICENSE).
