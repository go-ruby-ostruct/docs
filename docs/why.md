# Why pure Go

`go-ruby-ostruct/ostruct` reimplements the data structure underneath Ruby's
`OpenStruct` in **pure Go, with cgo disabled**. The slice of Ruby it covers — an
ordered attribute table and its accessors, conversions, comparison, and
inspection — is **deterministic and interpreter-independent**: given its inputs,
the result is a pure function of those inputs. That is exactly the part that can —
and should — live as a standalone Go library, separate from the interpreter.

## The data core vs. the dynamic glue

`OpenStruct`'s headline behavior is dynamic: *any* method name becomes an attribute
read or write. That dispatch — `method_missing`, `define_method`,
`respond_to_missing?` — needs a live Ruby binding and is **not** in this library;
it lives in the host (`rbgo`). But the dispatch is thin: a reader call resolves to
`Get`, a `name=` writer to `Set`, and `respond_to_missing?` to `RespondToField`.
Everything underneath — the ordered table, insertion-order `to_h`, `delete_field`
semantics, the MRI-exact `inspect` format — is deterministic and lives here.

## Extracted from rbgo, reusable by anyone

This library began life inside
[go-embedded-ruby](https://github.com/go-embedded-ruby/ruby)'s `rbgo`. It has been
**extracted into a reusable standalone library** so that:

- any Go program can import `github.com/go-ruby-ostruct/ostruct` directly, with no
  Ruby runtime;
- the dependency runs the *other* way — `rbgo` binds this module as a native
  module (the same pattern as [go-ruby-regexp](https://github.com/go-ruby-regexp)
  and [go-ruby-erb](https://github.com/go-ruby-erb)), rather than this module
  depending on the interpreter;
- the behavior is pinned by a **differential oracle** against the system `ruby`,
  independent of any one consumer.

## Why pure Go matters here

Because the library is CGO-free and dependency-free, it:

- cross-compiles to every Go target with no C toolchain, and links into a single
  static binary;
- has **no dependency on the Ruby runtime** — the dependency runs the other way;
- can be differentially tested against the `ruby` binary wherever one is on
  `PATH`, while the cross-arch lanes (where `ruby` is absent) still validate the
  library itself, since the built-in value renderer keeps `inspect` deterministic
  and ruby-free.

See [Usage & API](api.md) for the surface and [Roadmap](roadmap.md) for what is
in scope.
