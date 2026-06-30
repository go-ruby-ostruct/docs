# go-ruby-ostruct documentation

**Ruby's OpenStruct data core in pure Go — MRI-compatible, no cgo.**

`go-ruby-ostruct/ostruct` is a faithful, pure-Go (zero cgo) reimplementation of the
data structure underneath Ruby's `OpenStruct` (`require "ostruct"`), matching
reference Ruby (MRI 4.0.5) byte-for-byte on the `inspect` format, `to_h` insertion
ordering, and `delete_field` semantics. The module path is
`github.com/go-ruby-ostruct/ostruct`.

It implements an **ordered, `Symbol`-keyed attribute table** with the accessors,
conversions, comparison, and inspection MRI exposes. It was **extracted from
rbgo's internals into a reusable standalone library**: the module is standalone
and importable by any Go program, and it is the backend bound into
[go-embedded-ruby](https://github.com/go-embedded-ruby/ruby) by `rbgo` as a native
module — just like [go-ruby-regexp](https://github.com/go-ruby-regexp) and
[go-ruby-erb](https://github.com/go-ruby-erb). The dependency runs the other way:
this library has **no dependency on the Ruby runtime**.

!!! success "Status: complete — MRI byte-exact"
    The ordered attribute table and its MRI-faithful behavior — `Get` / `Set`, `Index` / `SetIndex` (`[]` / `[]=`), `ToH`, `EachPair`, `Members`, `Dig`, `RespondToField`, `DeleteField`, `Equal` / `Eql`, and `Inspect` / `String`. The `inspect` format, `to_h` insertion ordering, and `delete_field` semantics are validated **byte-for-byte** against the system `ruby` at 100% coverage, `gofmt` + `go vet` clean, CI green across the six 64-bit Go targets and three OSes.

## What stays in the host

`OpenStruct`'s defining feature — turning *any* method name into an attribute read
or write — is dynamic and lives in the host (`rbgo`): `method_missing`,
`define_method`, and `respond_to_missing?`. That glue is implemented **in terms of
this table**: a reader call becomes `Get`, a `name=` call becomes `Set`, and
`respond_to_missing?` consults `RespondToField`. This package owns the table and
its MRI-faithful behavior; the host owns the dynamic dispatch.

## Quick taste

```go
o := ostruct.New(
	ostruct.Pair{Key: "name", Value: "John"},
	ostruct.Pair{Key: "age", Value: 70},
)
o.Inspect()                     // #<OpenStruct name="John", age=70>
o.Get("name")                   // "John"  (the method_missing reader target)
old, _ := o.DeleteField("age")  // 70       (NameError if absent)
```

```text
OpenStruct.new(name: "John", age: 70).inspect  #=> #<OpenStruct name="John", age=70>
OpenStruct.new(b: 1, a: 2, c: 3).to_h.keys      #=> [:b, :a, :c]   (insertion order)
```

## Repositories

| Repo | What it is |
| --- | --- |
| [`ostruct`](https://github.com/go-ruby-ostruct/ostruct) | the library — Ruby's OpenStruct data core in pure Go |
| [`docs`](https://github.com/go-ruby-ostruct/docs) | this documentation site (MkDocs Material, versioned with mike) |
| [`go-ruby-ostruct.github.io`](https://github.com/go-ruby-ostruct/go-ruby-ostruct.github.io) | the organization landing page (Hugo) |
| [`brand`](https://github.com/go-ruby-ostruct/brand) | logo and brand assets |

## Principles

- **Pure Go, `CGO_ENABLED=0`** — trivial cross-compilation, a single static
  binary, no C toolchain.
- **MRI byte-exact.** The `inspect` format, `to_h` ordering, and `delete_field`
  semantics match reference Ruby exactly, validated by a differential oracle
  against the `ruby` binary.
- **The data core, not the dynamic glue.** This package owns the ordered table;
  `OpenStruct`'s dynamic attribute dispatch stays in the host, implemented in
  terms of this table.
- **Standalone & reusable.** Extracted from rbgo's internals; no dependency on
  the Ruby runtime — the dependency runs the other way.
- **100% test coverage** is the target, enforced as a CI gate, across 6 arches
  and 3 OSes.

## Where to go next

- [Why pure Go](why.md) — why this slice of Ruby is deterministic enough to live
  as a standalone, interpreter-independent Go library.
- [Usage & API](api.md) — the public surface and worked examples.
- [Roadmap](roadmap.md) — what is done and what is downstream by design.

Source lives at [github.com/go-ruby-ostruct/ostruct](https://github.com/go-ruby-ostruct/ostruct).
