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

## Library-level benchmark (Go API vs runtimes) — 2026-07-03

This section measures the **pure-Go library directly, through its Go API** — not
the `rbgo` interpreter path recorded above. It isolates the library primitive
from Ruby-interpreter dispatch, answering the parity question head-on: *is the
pure-Go implementation as fast as the reference runtime's own `OpenStruct`?* The
**same workload, same inputs, same iteration counts** run through the Go library
and through each reference runtime's stdlib; the drivers' digests were checked
**byte-identical to MRI** before any timing.

- **Host:** Apple M4 Max (`Mac16,5`, arm64), macOS 26.5.1 — **date 2026-07-03**.
- **Runtimes:** Go 1.26.4 · MRI `ruby 4.0.5 +PRISM` · MRI + YJIT · JRuby 10.1.0.0
  (OpenJDK 25) · TruffleRuby 34.0.1 (GraalVM CE Native).
- **Workload:** a fixed 40-field OpenStruct (`:f0..:f39`, deterministic integer
  values `i*31+7`) driving the five representative operations below.
- **Method:** each process runs 3 untimed warm-up passes, then 25 timed passes of
  a fixed inner loop, timed with a monotonic clock; the **best** pass is reported
  as **ns/op** (lower is better). `vs MRI` < 1.00× means *faster than MRI*.
  Interpreter start-up is outside the timed region, so these are operation costs,
  not `ruby file.rb` process costs.

**go vs YJIT verdict:** the pure-Go library **beats MRI + YJIT on four of the five
operations** — `construct` (~67× faster than YJIT), `write` / dynamic member add
(~33× faster), `read` (0.46× YJIT), and `index` (0.93× YJIT, a thin win). It
**loses only on `to_h`** (~5.7× YJIT), which is the module's one optimization
target — see the note below.

#### construct-40 — `OpenStruct.new(hash)` from a 40-field hash

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 1024.0 | 0.01× |
| MRI | 69597.0 | 1.00× |
| MRI + YJIT | 68999.5 | 0.99× |
| JRuby | 28278.6 | 0.41× |
| TruffleRuby | 159135.0 | 2.29× |

Construction is where MRI's `OpenStruct` is famously slow: `new(hash)` defines a
singleton accessor method **per field** via `define_method`, so 40 fields cost ~70
µs. The pure-Go build is an ordered map fill — **~68× faster than MRI and YJIT**
(1024.0 / 68999.5 = **0.015× YJIT**). YJIT cannot help: the cost is in
metaprogramming (method-table churn), not interpreted bytecode.

#### write-40 — dynamic member add (`os.f = v` on a fresh struct)

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 3060.0 | 0.03× |
| MRI | 102103.0 | 1.00× |
| MRI + YJIT | 100253.0 | 0.98× |
| JRuby | 39565.1 | 0.39× |
| TruffleRuby | 182686.0 | 1.79× |

Growing a fresh struct one **new** attribute at a time is the worst case for MRI's
`method_missing` + `define_singleton_method` writer path (~100 µs for 40 adds). The
Go writer just appends to the key slice and stores in the map — **~33× faster than
YJIT** (3060.0 / 100253.0 = **0.031×**). This is the single biggest win.

#### read-40 — attribute read (`method_missing` reader path)

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 1179.4 | 0.41× |
| MRI | 2844.5 | 1.00× |
| MRI + YJIT | 2557.0 | 0.90× |
| JRuby | 2743.2 | 0.96× |
| TruffleRuby | 1435.5 | 0.50× |

Once the accessors exist, reads are cheaper, but a pure-Go map lookup still
**beats YJIT** (1179.4 / 2557.0 = **0.46× YJIT**; 0.41× MRI). YJIT recovers some
ground over plain MRI (0.90×) by compiling the accessor, but does not catch the Go
table.

#### index-40 — `[]=` write-through then `[]` read-back

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 2925.9 | 0.57× |
| MRI | 5166.0 | 1.00× |
| MRI + YJIT | 3152.5 | 0.61× |
| JRuby | 2538.0 | 0.49× |
| TruffleRuby | 510.0 | 0.10× |

The bracket accessors (`[]`/`[]=`) are ordinary method calls that hash the key each
time. Go **edges YJIT** here (2925.9 / 3152.5 = **0.93× YJIT**; 0.57× MRI) — a thin
win. TruffleRuby's Graal JIT is dramatically faster on this tight steady-state loop
(0.10×), the one op where a warmed JIT clearly leads; JRuby also beats Go slightly
(0.49×). This is a steady-state hot loop the JITs are built for.

#### to_h-40 — ordered serialisation to a Hash

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 767.1 | 4.62× |
| MRI | 166.0 | 1.00× |
| MRI + YJIT | 134.5 | 0.81× |
| JRuby | 295.5 | 1.78× |
| TruffleRuby | 192.0 | 1.16× |

The **one operation where the pure-Go library loses**: MRI's `to_h` is a tight C
copy of an already-built Hash, whereas the Go `ToH` allocates a fresh `[]Pair` and
does one map lookup per key to rebuild insertion order — **~5.7× YJIT** (767.1 /
134.5), ~4.6× MRI. This is the module's remaining optimization target: cache the
ordered pairs (or expose a zero-copy view over the internal `keys`/`table`) so
`ToH` avoids the per-key re-hash. It is a sub-microsecond row, so treat the ratio
as order-of-magnitude — but the direction is stable across runs.

!!! note "Reproduce"
    The harness is committed under
    [`benchmarks/`](https://github.com/go-ruby-ostruct/docs/tree/main/benchmarks):
    a self-contained Go driver (`go/`, pins the published library via
    `go.mod` pseudo-version), the equivalent `ruby/ostruct.rb` workload, and
    `run.sh`. Run `bash benchmarks/run.sh`; it first checks the Go driver's output
    **byte-identical to MRI**, then times every available runtime. Env
    `OUTER`/`WARM` tune the pass budget and `RUBY`/`JRUBY`/`TRUFFLERUBY` select the
    runtime binaries.

!!! warning "Warm-up budget & noise — honest framing"
    Numbers reflect a **fixed warm-process budget** (3 warm-up + 25 timed passes
    in one process). The JVM/GraalVM JITs (JRuby, TruffleRuby) may need a larger
    warm-up to reach steady state, so their columns can **understate** peak
    throughput on the longer loops and **overstate** it on the shortest ones where
    they happen to have warmed (see `index-40`). Sub-microsecond rows (`to_h-40`)
    carry the most relative noise; treat those ratios as order-of-magnitude. Every
    number here is a **real measured value** from the dated run above — nothing is
    fabricated, estimated, or cherry-picked. The go-ruby column is the pure-Go
    library; every other column is that interpreter's own `ostruct` stdlib doing
    the equivalent work.
