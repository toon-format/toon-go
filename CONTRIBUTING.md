# Contributing to toon-go

## Development Setup

The spec conformance fixtures live in the `tests/spec` submodule, so clone recursively:

```bash
git clone --recurse-submodules https://github.com/toon-format/toon-go.git
cd toon-go
go test ./...
```

## Pull Requests

Add tests for every behavior change and use [Conventional Commits](https://www.conventionalcommits.org/) for commit messages. Changes to the format itself belong in [toon-format/spec](https://github.com/toon-format/spec).

## Maintainers

- [@bpradana](https://github.com/bpradana)
- [@johannschopplich](https://github.com/johannschopplich)

All maintainers have equal decision-making power. Open an issue before a major architectural change.
