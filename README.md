# pseudonymous
[![MegaLinter](https://github.com/diz-unimr/pseudonymous/actions/workflows/mega-linter.yml/badge.svg)](https://github.com/diz-unimr/pseudonymous/actions/workflows/mega-linter.yml) ![go](https://github.com/diz-unimr/pseudonymous/actions/workflows/build.yml/badge.svg) ![docker](https://github.com/diz-unimr/pseudonymous/actions/workflows/release.yml/badge.svg) [![codecov](https://codecov.io/github/diz-unimr/pseudonymous/branch/main/graph/badge.svg?token=waEHzvF9pf)](https://codecov.io/github/diz-unimr/pseudonymous)
> Pseudonymize FHIR🔥 resources from and to MongoDB databases

This is a command line tool to run pseudonymization / anonymization on FHIR resources via the
[FHIR® Pseudonymizer](https://github.com/miracum/fhir-pseudonymizer).
It currently supports reading from a MongoDB database. Data from all collections are send to the FHIR pseudonymization
service and written back to a target database and its collections accordingly.

## Usage

The project name is the only _required_ command line flag to provide when running the tool.
It is used by convention as a suffix to determine the source database (`idat_fhir_[project]`).
The target database name is set to `psn_fhir_[project]`.

Configuration properties are set via a YAML file which defaults to `app.yaml`.

```
Usage:
  pseudonymous [flags]

Flags:
  -c, --config string    config file (default is ./app.yaml)
  -h, --help             help for pseudonymous
  -p, --project string   project name (required)
```

## Installation

Binary releases and docker images are available under
[releases](https://github.com/diz-unimr/pseudonymous/releases).

## Configuration properties

| Name                                | Default                    | Description                                                        |
|-------------------------------------|----------------------------|--------------------------------------------------------------------|
| `app.log-level`                     | info                       | Log level (error,warn,info,debug)                                  |
| `app.concurrency`                   | 5                          | Number of concurrent threads                                       |
| `gpas.url`                          |                            | URL to the gPAS service for auto-creating pseudonymization domains |
| `fhir.provider.mongodb.connection`  | mongodb://localhost        | MongoDB connection string                                          |
| `fhir.provider.mongodb.batch-size`  | 5000                       | Batch size when reading data from the source database              |
| `fhir.pseudonymizer.url`            | http://localhost:5000/fhir | FHIR® Pseudonymizer endpoint                                       |
| `fhir.pseudonymizer.retry.count`    | 10                         | Retry count                                                        |
| `fhir.pseudonymizer.retry.timeout`  | 10                         | Retry timeout                                                      |
| `fhir.pseudonymizer.retry.wait`     | 5                          | Retry wait between retries                                         |
| `fhir.pseudonymizer.retry.max-wait` | 20                         | Retry maximum wait                                                 |

### Environment variables

Override configuration properties by providing environment variables with their respective names.
Upper case env variables are supported as well as underscores (`_`) instead of `.` and `-`.

## License

[AGPL-3.0](https://www.gnu.org/licenses/agpl-3.0.en.html)
