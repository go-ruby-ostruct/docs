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

**go vs YJIT verdict:** the pure-Go library now **beats MRI + YJIT on all five
operations** — `construct` (~53× faster than YJIT), `write` / dynamic member add
(~27× faster), `read` (0.50× YJIT), `index` (0.91× YJIT, a thin win), and — after
the 2026-07-03 optimization below — `to_h` (**0.69× YJIT**, 0.56× MRI), which was
previously the module's one loss (~6.4× YJIT). The `to_h` fix stores the table's
entries in an insertion-ordered slice so serialisation is a single slice copy with
no per-key re-hash; see the `to_h-40` note.

#### construct-40 — `OpenStruct.new(hash)` from a 40-field hash

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 1271.8 | 0.02× |
| MRI | 67886.0 | 1.00× |
| MRI + YJIT | 67754.0 | 1.00× |
| JRuby | 28397.6 | 0.42× |
| TruffleRuby | 162990.7 | 2.40× |

Construction is where MRI's `OpenStruct` is famously slow: `new(hash)` defines a
singleton accessor method **per field** via `define_method`, so 40 fields cost ~68
µs. The pure-Go build is an ordered slice fill plus an index insert — **~53× faster
than MRI and YJIT** (1271.8 / 67754.0 = **0.019× YJIT**). YJIT cannot help: the cost
is in metaprogramming (method-table churn), not interpreted bytecode.

#### write-40 — dynamic member add (`os.f = v` on a fresh struct)

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 3575.9 | 0.04× |
| MRI | 100180.0 | 1.00× |
| MRI + YJIT | 97715.0 | 0.98× |
| JRuby | 40042.9 | 0.40× |
| TruffleRuby | 178721.7 | 1.78× |

Growing a fresh struct one **new** attribute at a time is the worst case for MRI's
`method_missing` + `define_singleton_method` writer path (~100 µs for 40 adds). The
Go writer just appends the entry and records its position in the index — **~27×
faster than YJIT** (3575.9 / 97715.0 = **0.037×**). This is the single biggest win.

#### read-40 — attribute read (`method_missing` reader path)

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 1249.9 | 0.43× |
| MRI | 2891.0 | 1.00× |
| MRI + YJIT | 2481.0 | 0.86× |
| JRuby | 2545.5 | 0.88× |
| TruffleRuby | 1214.6 | 0.42× |

Once the accessors exist, reads are cheaper, but a pure-Go map probe into the index
plus a slice load still **beats YJIT** (1249.9 / 2481.0 = **0.50× YJIT**; 0.43×
MRI). YJIT recovers some ground over plain MRI (0.86×) by compiling the accessor,
but does not catch the Go table.

#### index-40 — `[]=` write-through then `[]` read-back

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 2775.1 | 0.53× |
| MRI | 5203.5 | 1.00× |
| MRI + YJIT | 3044.5 | 0.59× |
| JRuby | 2524.4 | 0.49× |
| TruffleRuby | 381.4 | 0.07× |

The bracket accessors (`[]`/`[]=`) are ordinary method calls that hash the key each
time. Go **edges YJIT** here (2775.1 / 3044.5 = **0.91× YJIT**; 0.53× MRI) — a thin
win. TruffleRuby's Graal JIT is dramatically faster on this tight steady-state loop
(0.07×), the one op where a warmed JIT clearly leads; JRuby also beats Go slightly
(0.49×). This is a steady-state hot loop the JITs are built for.

#### to_h-40 — ordered serialisation to a Hash

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 89.5 | 0.56× |
| MRI | 160.0 | 1.00× |
| MRI + YJIT | 130.5 | 0.82× |
| JRuby | 176.0 | 1.10× |
| TruffleRuby | 191.2 | 1.19× |

Formerly the module's **one loss**, `to_h` now **beats every reference runtime**,
including MRI + YJIT (89.5 / 130.5 = **0.69× YJIT**; 0.56× MRI). The fix
([go-ruby-ostruct/ostruct#1](https://github.com/go-ruby-ostruct/ostruct/pull/1)):
the table's entries are stored in an **insertion-ordered slice of pairs** (each key
a Symbol) with a separate `Symbol → position` index. `to_h` was previously
`824.7 ns/op` (~6.4× YJIT) because it walked the ordered keys and did **one map
probe per key** to re-pair each value with its position — an O(n) re-hash of an
order the struct already knew. With the values already sitting in insertion order,
`ToH` is a single `make`+`copy` with **zero per-key hashing**; a fresh slice is
still returned, so callers may mutate it exactly as Ruby's `to_h` returns an
independent Hash — matching MRI's tight C hash-copy. Random access (`[]`, readers)
stays a single map probe, so `construct`/`write`/`read`/`index` are unaffected.
This is a sub-microsecond row, so treat the ratio as order-of-magnitude — but the
direction (a ~9× speedup, now under YJIT) is stable across runs.

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
