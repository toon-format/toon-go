# Contributing to toon-go

## Development Setup

The spec conformance fixtures live in the `tests/spec` submodule, so clone recursively:

```bash
git clone --recurse-submodules https://github.com/toon-format/toon-go.git
cd toon-go
go test ./...
```

## Pull Requests

Spec behavior is tested through the spec fixtures – a missing case goes to [toon-format/spec](https://github.com/toon-format/spec) as a fixture, and so do changes to the format itself. A test under `tests/` is only for API the spec does not describe, such as struct tags, `Unmarshal`, and options. Use [Conventional Commits](https://www.conventionalcommits.org/) for commit messages.

## Maintainers

- [@bpradana](https://github.com/bpradana)
- [@johannschopplich](https://github.com/johannschopplich)

All maintainers have equal decision-making power. Open an issue before a major architectural change.
