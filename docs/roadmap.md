# Roadmap

`go-ruby-ostruct/ostruct` is grown **test-first**, each capability differential-tested
against MRI rather than built in isolation. The `OpenStruct` data core — the
deterministic, interpreter-independent slice extracted from rbgo's internals — is
**complete**.

| Stage | What | Status |
| --- | --- | --- |
| Ordered attribute table | The data structure underneath `OpenStruct`: an ordered, `Symbol`-keyed table built from `New(pairs...)`, preserving insertion order exactly as MRI's `to_h` does. | **Done** |
| Accessors | `Get` / `Set` (the `method_missing` reader / `name=` writer targets), `Index` / `SetIndex` (`[]` / `[]=`), and `RespondToField` (the table half of `respond_to_missing?`). | **Done** |
| Conversions & traversal | `ToH` (Symbol keys, insertion order), `EachPair`, `Members`, and `Dig(keys...)` routed through a `Digger` interface. | **Done** |
| delete_field & comparison | `DeleteField` returns the old value and raises `NameError` when absent; `Equal` / `Eql` implement `==` / `eql?`. | **Done** |
| MRI-exact inspect | `Inspect` / `String` render `#<OpenStruct …>` byte-for-byte through an `Inspector` interface with a built-in renderer for common shapes. | **Done** |
| Differential oracle & coverage | The `inspect` format, `to_h` insertion ordering, and `delete_field` semantics compared against the system `ruby` byte-for-byte; 100% coverage, gofmt + go vet clean, green across all six 64-bit Go arches and three OSes. | **Done** |

## Documented out-of-scope boundaries

These are **deliberate**, recorded so the module's surface is unambiguous:

- **No dynamic dispatch.** `OpenStruct`'s defining feature — turning *any* method
  name into a read or write — is dynamic and needs a live Ruby binding. It lives
  in the host (`rbgo`), implemented in terms of this table (`Get` / `Set` /
  `RespondToField`), not here.
- **Reference is reference Ruby (MRI).** Byte-for-byte conformance targets MRI's
  behaviour; differences across MRI releases are matched to the reference used by
  the differential oracle.
- **Standalone & reusable.** The module has no dependency on the Ruby runtime; the
  dependency runs the other way.

See [Usage & API](api.md) for the surface and [Why pure Go](why.md) for the
deterministic/interpreter split.
