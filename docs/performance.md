# Performance

`go-ruby-ostruct/ostruct` is the pure-Go library that
[`rbgo`](https://github.com/go-embedded-ruby/ruby) binds for Ruby's `OpenStruct`.
This page records the **methodology** of the comparative benchmark of that module
against the reference Ruby runtimes, part of the ecosystem-wide per-module parity
suite.

## Result (best of 5, ms)

Measured 2026-06-30 on **Apple M4 Max**, macOS (darwin/arm64), Go 1.26.4, with
`ruby 4.0.5 +PRISM`, `jruby 10.1.0.0` (OpenJDK 25) and `truffleruby 34.0.1`
(GraalVM CE Native). The cross-runtime workload builds an `OpenStruct` from a
40-field hash, then reads / mutates / `dig`s / `to_h`s it; checksum byte-identical
to MRI before timing.

| Runtime | time | vs MRI |
| --- | ---: | ---: |
| **rbgo** (go-ruby-ostruct) | 860 | **0.52×** |
| MRI (ruby 4.0.5) | 1660 | 1.00× |
| MRI + YJIT | 1690 | 1.02× |
| JRuby 10.1.0.0 | 2160 | 1.30× |
| TruffleRuby 34.0.1 | 4200 | 2.53× |

rbgo runs on **go-ruby-ostruct** and is **~1.9× faster than MRI** here (0.52×):
the pure-Go field table builds, reads and serialises more cheaply than MRI's
`OpenStruct`, whose per-attribute access goes through `method_missing` and dynamic
singleton-method definition in Ruby. The compiled table operations dominate.

!!! note "Honest framing"
    JRuby and TruffleRuby are timed **cold, single-shot**, so they carry JVM /
    Graal startup on every run — read them as one-shot `ruby file.rb` costs, the
    same way `rbgo` and MRI are measured, not as steady-state JIT numbers. These
    are **real measured numbers** from the 2026-06-30 run (Apple M4 Max;
    `ruby 4.0.5 +PRISM`, `jruby 10.1.0.0`, `truffleruby 34.0.1`) — nothing is
    fabricated or cherry-picked.

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
