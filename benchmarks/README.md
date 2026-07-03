<!-- SPDX-License-Identifier: BSD-3-Clause -->
# `go-ruby-ostruct` library-level benchmark harness

Reproducible, cross-runtime benchmark of the **pure-Go `go-ruby-ostruct` library**
against the reference Ruby runtimes (MRI, MRI + YJIT, JRuby, TruffleRuby). It
measures the **library primitive** through its Go API, isolated from the rbgo
interpreter, so the numbers answer: *is the pure-Go implementation as fast as the
reference runtime's own `OpenStruct`?*

## Layout

- `go/`            — self-contained Go driver; `go.mod` pins the published library
  by pseudo-version (no `replace`).
- `ruby/ostruct.rb` — the equivalent workload; `ruby/_harness.rb` is the shared timer.
- `run.sh`         — verifies Go output == MRI, then runs every available runtime
  and prints one Markdown table per sub-benchmark (ns/op + ratio vs MRI).

## Run

```sh
bash benchmarks/run.sh
```

Environment knobs: `OUTER` (timed passes, default 25), `WARM` (untimed warm-up
passes, default 3), and `RUBY`/`JRUBY`/`TRUFFLERUBY` to select runtime binaries.

## Workload

A fixed 40-field OpenStruct (`:f0..:f39`, deterministic integer values `i*31+7`)
exercises the five representative operations:

- **construct-40** — `OpenStruct.new(hash)` from the 40-field hash.
- **read-40** — read every attribute (MRI's `method_missing` reader path).
- **write-40** — grow a fresh empty struct one **dynamic member** at a time.
- **to_h-40** — serialise the table to an ordered Hash.
- **index-40** — `[]=` write-through then `[]` read-back over all 40 fields.

## Method

Each process runs `WARM` untimed passes (to let the JVM/GraalVM JITs warm up),
then `OUTER` timed passes of a fixed inner loop, timed with a monotonic clock;
the **best** pass is reported as **ns/op**. Interpreter start-up is outside the
timed region. The Go driver and the Ruby script build **identical inputs** and
their `VERIFY` digests are checked **byte-identical to MRI** before any timing.
Results are published, dated, in `../docs/performance.md`.
