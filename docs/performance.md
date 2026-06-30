# Performance

`go-ruby-ostruct/ostruct` is the pure-Go library that
[`rbgo`](https://github.com/go-embedded-ruby/ruby) binds for Ruby's `OpenStruct`.
This page records the **methodology** of the comparative benchmark of that module
against the reference Ruby runtimes, part of the ecosystem-wide per-module parity
suite.

!!! note "Methodology only"
    No measured numbers are published here yet. This page documents *how* the
    benchmark is run so the result is reproducible and apples-to-apples; the
    measured table will be added once the run is recorded, the same way it is for
    the sibling modules — never fabricated.

## What is measured

The **same** Ruby script — building an `OpenStruct`, reading and writing
attributes through `method_missing`, `to_h`, `dig`, and `inspect` over a
representative record — is run under every runtime. `rbgo`'s number reflects
**this pure-Go library doing the table work** (the dynamic dispatch is the host's,
the table operations are this library's); every other column is that interpreter's
own `ostruct` stdlib. So the comparison is the **Ruby-visible operation**,
apples-to-apples across interpreters. The script prints a deterministic checksum
and its output is checked **byte-identical to MRI** before timing.

## How it is run

- **Method:** best-of-N wall time (best, not mean, to suppress scheduler noise);
  single-shot processes, no warm-up beyond the script's own loop.
- **Runtimes:** `ruby` (MRI, the oracle) and `ruby --yjit`; `jruby` (on the JVM);
  `truffleruby` (GraalVM). JVM/Graal rows are timed **cold, single-shot**, so they
  carry runtime startup on every run — read them as one-shot `ruby file.rb` costs,
  the same way `rbgo` and MRI are measured, not as steady-state JIT numbers.
- The benchmark script and harness live in rbgo's repo under
  [`bench/modules/`](https://github.com/go-embedded-ruby/ruby/tree/main/bench/modules)
  (`ostruct.rb` + `run.sh`). Reproduce with the same
  `RBGO=./rbgo TRUFFLE=truffleruby bash bench/modules/run.sh N` invocation used
  across the ecosystem.

## Honest framing

Rows that complete in well under ~200 ms carry the most relative noise; their
ratios should be read as order-of-magnitude. Any numbers added here will be real
measured numbers from a dated run — nothing cherry-picked.
