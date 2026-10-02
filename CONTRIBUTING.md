# Contributing to toon-go

## Development Setup

The spec conformance fixtures live in the `tests/spec` submodule, so clone recursively:

```bash
git clone --recurse-submodules https://github.com/toon-format/toon-go.git
cd toon-go
go mod download
go test ./...
```

CI runs the tests on Go 1.23, 1.24, and 1.25 and lints with `golangci-lint` v2.1. The minimum Go version is the one in `go.mod`.

## Coding Standards

Format with `go fmt` and check with `go vet` before committing. Encoding and decoding behavior must match the [TOON specification](https://github.com/toon-format/spec/blob/main/SPEC.md).

## Pull Requests

Add tests for every behavior change and use [Conventional Commits](https://www.conventionalcommits.org/) for commit messages. Changes to the format itself belong in [toon-format/spec](https://github.com/toon-format/spec).

## Maintainers

- [@bpradana](https://github.com/bpradana)
- [@johannschopplich](https://github.com/johannschopplich)

All maintainers have equal decision-making power. Open an issue before a major architectural change.

## License

By contributing, you agree that your contributions are licensed under the MIT License.
