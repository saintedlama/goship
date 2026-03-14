# GoShip

Bring fun into Go development with GoShip and GitHub Actions!

🔭 test
🗺️ coverage
🧭 vet
🏄🏾 fmt

GoShip runs your Go tests, generates coverage reports, vets your code, and checks your formatting and adds results to the GitHub Steps summary. It's like having a friendly shipmate who keeps your code in shipshape condition while you focus on writing great Go code.

## Usage

```yaml
name: CI

on:
  push:
    branches:
      - master
  pull_request:

jobs:
  build-and-lint:
    runs-on: ubuntu-latest
    permissions:
      checks: write
      contents: read

    steps:
      - name: Checkout
        uses: actions/checkout@v6

      - name: Set up Go
        uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
          cache: true

      - name: Download dependencies
        run: go mod download

      - name: Build
        run: go build ./...

      - name: GoShip
        uses: saintedlama/goship@v1
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
          test: true      # Default: true
          coverage: true  # Default: true
          vet: true       # Default: true
          fmt: true       # Default: true
```

## References

- Use https://docs.github.com/en/actions/reference/workflows-and-actions/variables to learn about GitHub Actions variables.
- Use https://docs.github.com/en/actions/how-tos/create-and-publish-actions to learn about creating and publishing GitHub Actions.