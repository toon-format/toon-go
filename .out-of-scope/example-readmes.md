# READMEs in `examples/`

The programs in `examples/` don't get README files with descriptions, run commands, and expected output.

## Why this is out of scope

A README next to example code repeats what the code says, and nothing checks it. When the API or the encoder output changes, the program stops compiling or prints something new, but the README keeps showing the old output. Expected TOON output is the part that drifts first, because spec releases change it.

This repo already shows the drift. The README example and `examples/basic/main.go` both pass `toon.WithLengthMarkers(true)`, and spec v4.1, where the `[#N]` length marker no longer exists, means fixing each one by hand. Per-example READMEs with expected output would add five more places to fix on every spec release.

Go has a checked format for this. A testable example – an `ExampleMarshal` function in a `_test.go` file with an `// Output:` comment – runs under `go test`, fails when the output changes, and renders on [pkg.go.dev](https://pkg.go.dev/github.com/toon-format/toon-go) next to the function it documents. Explanations of the examples belong there.

## Prior requests

- [#12](https://github.com/toon-format/toon-go/issues/12) – "Docs: add README.md to examples directory"
