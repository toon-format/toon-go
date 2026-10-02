# Example Programs with READMEs

Example programs with README files – descriptions, run commands, and expected output – are out of scope.

## Why this is out of scope

A README next to example code repeats the code, and nothing checks it. When the API or the encoder output changes, the program stops compiling or prints something new, but the README keeps the old output – and expected TOON output drifts first, because spec releases change it.

Go has a checked format for this: a testable example – an `Example` function with an `// Output:` comment – runs under `go test`, fails when the output changes, and renders on [pkg.go.dev](https://pkg.go.dev/github.com/toon-format/toon-go) next to the function it documents. Explanations of the examples belong there.

## Prior requests

- [#12](https://github.com/toon-format/toon-go/issues/12) – "Docs: add README.md to examples directory"
