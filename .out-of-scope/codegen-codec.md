# Code-Generated Codec

toon-go has one codec: `toon.Marshal` and `toon.Unmarshal` on Go values via reflection. A second, code-generated path – as a `codegen` subpackage or as an alternative implementation inside this repo – is out of scope.

## Why this is out of scope

- **Every spec rule would live twice.** Quoting, tabular detection, and strict-mode errors would each have to land in the reflection path and in the generator, so every spec release becomes two migrations, and any drift between them is a conformance bug.
- **Users pay with a build step.** Generated codecs mean `go generate` and checked-in generated files for every encoded struct.
- **The speed-up doesn't matter where TOON is used.** TOON is "designed primarily for LLM prompts and contexts where every token costs" ([Purpose and Scope](https://github.com/toon-format/spec/blob/main/SPEC.md#purpose-and-scope)), and nanoseconds saved per marshal disappear next to the model call the output feeds.

A faster Go implementation is welcome as its own project; listing it is a PR against [`docs/ecosystem/implementations.md`](https://github.com/toon-format/toon/blob/main/docs/ecosystem/implementations.md) in toon-format/toon.

## Prior requests

- [#17](https://github.com/toon-format/toon-go/issues/17) – "Proposal: Add high-performance codegen-based Go implementation as an alternative"
