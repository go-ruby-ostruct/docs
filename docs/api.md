# Usage & API

The public API lives at the module root (`github.com/go-ruby-ostruct/ostruct`). It
is **Ruby-shaped but Go-idiomatic**: the methods mirror `OpenStruct`'s, while the
surface follows Go conventions — explicit errors, value types, no global state.
Keys accept a `Symbol` or `string` (interned via `ToSym`); values are held as
opaque `any`.

!!! success "Status: implemented"
    The library is built and importable as `github.com/go-ruby-ostruct/ostruct`,
    bound into `rbgo` as a native module; see [Roadmap](roadmap.md).

## Install

```sh
go get github.com/go-ruby-ostruct/ostruct
```

## Worked example

```go
o := ostruct.New(
	ostruct.Pair{Key: "name", Value: "John"},
	ostruct.Pair{Key: "age", Value: 70},
)

o.Inspect()                     // #<OpenStruct name="John", age=70>
o.Get("name")                   // "John"  (a reader → method_missing target)
o.Set("city", "Lyon")           // a name= writer → method_missing target
o.Index("age")                  // 70      ([])
for _, p := range o.ToH() {     // to_h: Symbol keys, insertion order
	_ = p.Key
}
old, err := o.DeleteField("age")  // 70, nil   (NameError if absent)
```

## API

| Go | Ruby |
| --- | --- |
| `New(pairs...)` | `OpenStruct.new(hash)` |
| `Get(name)` / `Set(name, v)` | reader / `name=` writer (the `method_missing` target) |
| `Index(name)` / `SetIndex(name, v)` | `[]` / `[]=` |
| `ToH()` | `to_h` (Symbol keys, insertion order) |
| `EachPair(fn)` | `each_pair` |
| `Members()` | `members` |
| `Dig(keys...)` | `dig(*keys)` |
| `RespondToField(name)` | the table half of `respond_to_missing?` |
| `DeleteField(name)` | `delete_field` (returns old value; `NameError` if absent) |
| `Equal(o)` / `Eql(o)` | `==` / `eql?` |
| `Inspect()` / `String()` | `inspect` / `to_s` |

`Inspect` and `Dig` route through the `Inspector` and `Digger` interfaces so the
host's Ruby values render and dig exactly as in MRI, with a built-in renderer
covering the common scalar/collection shapes for deterministic, ruby-free testing.

## MRI-faithful samples

```text
OpenStruct.new(name: "John", age: 70).inspect  #=> #<OpenStruct name="John", age=70>
OpenStruct.new.inspect                          #=> #<OpenStruct>
OpenStruct.new(b: 1, a: 2, c: 3).to_h.keys      #=> [:b, :a, :c]   (insertion order)
OpenStruct.new(a: {b: {c: 1}}).dig(:a, :b, :c)  #=> 1
o.delete_field(:age)                            #=> 70             (old value)
OpenStruct.new.delete_field(:x)                 #=> NameError: no field 'x' in #<OpenStruct>
OpenStruct.new(name: "John") == OpenStruct.new(name: "John")  #=> true
```

## MRI conformance

Correctness is defined by reference Ruby. A **differential oracle** runs a wide
corpus through both the system `ruby` and this library and compares the `inspect`
output, `to_h` ordering, and `delete_field` results **byte-for-byte** — not
approximated from memory. The oracle tests skip themselves where `ruby` is not on
`PATH` (e.g. the qemu arch lanes), so the cross-arch builds still validate the
library.

## Relationship to Ruby

`go-ruby-ostruct/ostruct` is **standalone and reusable**, and is the backend bound
into [go-embedded-ruby](https://github.com/go-embedded-ruby/ruby) by `rbgo` as a
native module — the same way [go-ruby-regexp](https://github.com/go-ruby-regexp)
and [go-ruby-erb](https://github.com/go-ruby-erb) are bound. The dynamic accessors
(`method_missing`, `respond_to_missing?`) stay in the host and are implemented in
terms of this table; this library has no dependency on the Ruby runtime.
