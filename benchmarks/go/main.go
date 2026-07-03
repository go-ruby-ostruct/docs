// Copyright (c) the go-ruby-* authors
// SPDX-License-Identifier: BSD-3-Clause
//
// Library-level cross-runtime benchmark driver for go-ruby-ostruct/ostruct.
// Builds the SAME deterministic 40-field workload as ruby/ostruct.rb and times
// each representative OpenStruct operation through the pure-Go library's Go API.
package main

import (
	"fmt"
	"strconv"

	"github.com/go-ruby-ostruct/ostruct"
)

// nFields is the fixed representative field count, identical to ruby/ostruct.rb.
const nFields = 40

// fieldName is the i-th attribute name ("f0".."f39").
func fieldName(i int) string { return "f" + strconv.Itoa(i) }

// fieldVal is the deterministic i-th value, identical to the Ruby workload.
func fieldVal(i int) int { return i*31 + 7 }

// seed builds the ordered key/value pairs that seed New — the same insertion
// order (f0..f39) MRI preserves from the source hash literal.
func seed() []ostruct.Pair {
	p := make([]ostruct.Pair, nFields)
	for i := 0; i < nFields; i++ {
		p[i] = ostruct.Pair{Key: fieldName(i), Value: fieldVal(i)}
	}
	return p
}

// digest renders an OpenStruct as "k=v|k=v|..." in insertion order, so the Go
// and Ruby drivers can be checked byte-identical before any timing is recorded.
func digest(o *ostruct.OpenStruct) string {
	s := ""
	for i, pr := range o.ToH() {
		if i > 0 {
			s += "|"
		}
		s += fmt.Sprintf("%s=%v", pr.Key, pr.Value)
	}
	return s
}

func main() {
	pairs := seed()
	base := ostruct.New(pairs...) // pre-built 40-field struct for read/to_h/index

	// --- Correctness digests (printed for cross-runtime verification) ---------

	// construct: the bulk hash constructor.
	fmt.Printf("VERIFY\tconstruct\t%s\n", digest(ostruct.New(pairs...)))

	// read: sum of every attribute read (the method_missing reader path in MRI).
	readSum := 0
	for i := 0; i < nFields; i++ {
		readSum += base.Get(fieldName(i)).(int)
	}
	fmt.Printf("VERIFY\tread\t%d\n", readSum)

	// write: a fresh empty struct grown one dynamic member at a time.
	w := ostruct.New()
	for i := 0; i < nFields; i++ {
		w.Set(fieldName(i), fieldVal(i))
	}
	fmt.Printf("VERIFY\twrite\t%s\n", digest(w))

	// to_h: ordered serialisation of the table.
	fmt.Printf("VERIFY\ttoh\t%s\n", digest(base))

	// index: []=  write-through then []  read-back, summed.
	idxSum := 0
	for i := 0; i < nFields; i++ {
		base.SetIndex(fieldName(i), fieldVal(i))
		idxSum += base.Index(fieldName(i)).(int)
	}
	fmt.Printf("VERIFY\tindex\t%d\n", idxSum)

	// --- Timed benchmarks -----------------------------------------------------

	bench("construct-40", 2000, func() { sink = ostruct.New(pairs...) })

	bench("read-40", 2000, func() {
		s := 0
		for i := 0; i < nFields; i++ {
			s += base.Get(fieldName(i)).(int)
		}
		sink = s
	})

	bench("write-40", 2000, func() {
		o := ostruct.New()
		for i := 0; i < nFields; i++ {
			o.Set(fieldName(i), fieldVal(i))
		}
		sink = o
	})

	bench("to_h-40", 2000, func() { sink = base.ToH() })

	bench("index-40", 2000, func() {
		s := 0
		for i := 0; i < nFields; i++ {
			base.SetIndex(fieldName(i), fieldVal(i))
			s += base.Index(fieldName(i)).(int)
		}
		sink = s
	})
}
