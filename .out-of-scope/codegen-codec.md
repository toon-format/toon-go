# Code-Generated Codec

toon-go has one codec: `toon.Marshal` and `toon.Unmarshal` on Go values via reflection. A second, code-generated path – as a `codegen` subpackage or as an alternative implementation inside this repo – is out of scope.

## Why this is out of scope

- **Every spec rule would live twice.** Quoting ([§7.2](https://github.com/toon-format/spec/blob/main/SPEC.md#72-quoting-rules-for-string-values-encoding)), tabular detection ([§9.3](https://github.com/toon-format/spec/blob/main/SPEC.md#93-arrays-of-objects--tabular-form)), strict-mode errors ([§14](https://github.com/toon-format/spec/blob/main/SPEC.md#14-strict-mode-errors-and-diagnostics-authoritative-checklist)) – each would need to land in the reflection path, in the generator, and in the code it emits. The spec moves, so every spec release would be two migrations, and any drift between them is a conformance bug that only one path has.
- **Users pay with a build step.** Generated codecs mean `go generate` and checked-in generated files for every struct that gets encoded.
- **The speed-up doesn't matter where TOON is used.** The spec describes TOON as "designed primarily for LLM prompts and contexts where every token costs" ([Purpose and Scope](https://github.com/toon-format/spec/blob/main/SPEC.md#purpose-and-scope)). The proposal in #17 measured ~42 ns/op versus ~160 ns/op per marshal – a difference that disappears next to the model call the output feeds.

A faster Go implementation is welcome as its own project. Listing it as a community implementation is a PR against [`docs/ecosystem/implementations.md`](https://github.com/toon-format/toon/blob/main/docs/ecosystem/implementations.md) in toon-format/toon.

## Prior requests

- [#17](https://github.com/toon-format/toon-go/issues/17) – "Proposal: Add high-performance codegen-based Go implementation as an alternative"
