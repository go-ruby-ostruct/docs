# frozen_string_literal: true
# SPDX-License-Identifier: BSD-3-Clause
#
# Cross-runtime workload for OpenStruct, identical (same inputs, same iteration
# counts) to benchmarks/go/main.go. A fixed 40-field struct exercises the five
# representative operations: construct-from-hash, attribute read (method_missing),
# attribute write / dynamic member add, to_h, and the []/[]= bracket accessors.
require "ostruct"
require_relative "_harness"

N = 40
NAMES = (0...N).map { |i| :"f#{i}" }        # :f0 .. :f39, insertion order
VALS  = (0...N).map { |i| i * 31 + 7 }      # deterministic values
SEED  = NAMES.each_with_index.to_h { |k, i| [k, VALS[i]] }

def digest(os)
  os.to_h.map { |k, v| "#{k}=#{v}" }.join("|")
end

base = OpenStruct.new(SEED)                  # pre-built struct for read/to_h/index

# --- Correctness digests (checked byte-identical to MRI before timing) --------
printf("VERIFY\tconstruct\t%s\n", digest(OpenStruct.new(SEED)))

read_sum = 0
NAMES.each { |n| read_sum += base.public_send(n) }
printf("VERIFY\tread\t%d\n", read_sum)

w = OpenStruct.new
N.times { |i| w.public_send("#{NAMES[i]}=", VALS[i]) }
printf("VERIFY\twrite\t%s\n", digest(w))

printf("VERIFY\ttoh\t%s\n", digest(base))

idx_sum = 0
N.times do |i|
  base[NAMES[i]] = VALS[i]
  idx_sum += base[NAMES[i]]
end
printf("VERIFY\tindex\t%d\n", idx_sum)

# --- Timed benchmarks ---------------------------------------------------------
bench("construct-40", 2000) { OpenStruct.new(SEED) }

bench("read-40", 2000) do
  s = 0
  NAMES.each { |n| s += base.public_send(n) }
  s
end

bench("write-40", 2000) do
  o = OpenStruct.new
  N.times { |i| o.public_send("#{NAMES[i]}=", VALS[i]) }
  o
end

bench("to_h-40", 2000) { base.to_h }

bench("index-40", 2000) do
  s = 0
  N.times do |i|
    base[NAMES[i]] = VALS[i]
    s += base[NAMES[i]]
  end
  s
end
