# GoShip ⚓️

Bring fun into Go development with GoShip and GitHub Actions!

GoShip runs your Go tests, generates coverage reports, vets your code, and checks formatting — then adds beautiful results to the GitHub step summary. Like a friendly shipmate keeping your code in shipshape condition.

## Usage

Add GoShip to your workflow after `actions/checkout`:

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:

jobs:
  ci:
    runs-on: ubuntu-latest
    permissions:
      contents: read

    steps:
      - uses: actions/checkout@v6

      - name: GoShip
        uses: saintedlama/goship@v1
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
```

All checks are enabled by default. Disable any you don't need:

```yaml
      - name: GoShip
        uses: saintedlama/goship@v1
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
          test: true      # run go test
          coverage: true  # collect and report coverage
          vet: true       # run go vet  (fails on issues)
          fmt: true       # run gofmt   (fails on unformatted files)
```

## Inputs

| Input | Default | Description |
|---|---|---|
| `github-token` | `github.token` | Token for GitHub API access |
| `working-directory` | `.` | Directory to run checks in |
| `test` | `true` | Run `go test` and add results to the step summary |
| `coverage` | `true` | Collect coverage and add a report to the step summary |
| `vet` | `true` | Run `go vet`; fails the action if issues are found |
| `fmt` | `true` | Run `gofmt`; fails the action if files need formatting |

## Step summary

GoShip writes a rich Markdown report directly to the GitHub step summary:

- 🔭 **Tests** — per-package results, pass/fail counts, elapsed time, and failure details
- 🗺️ **Coverage** — total and per-package percentages with visual indicators
- 🧭 **Vet** — all findings listed with file positions and messages
- 🏄🏾 **Fmt** — list of files that need formatting

## Contributing

### Prerequisites

- Go 1.26+
- Docker (for container builds)

### Build

```bash
make build        # compiles to bin/action
```

### Test

```bash
make test         # runs all unit tests
go test ./...     # equivalent
```

### Run locally

```bash
make run                                      # run against the current directory
WORKING_DIRECTORY=/path/to/your/go/project make run   # run against another project
```

The binary reads the same environment variables that the GitHub Actions runner injects, so you can also set them directly:

```bash
WORKING_DIRECTORY=. TEST=true COVERAGE=true VET=true FMT=true ./bin/action
```

### Docker

```bash
docker build -t goship .
docker run --rm \
  -e WORKING_DIRECTORY=/src \
  -v /path/to/your/go/project:/src \
  goship
```
