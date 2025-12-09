package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Flowing YAML with deepmerge functipn", func() {
	It("handles multiple maps", func() {
		source := parseYAML(`
---
in:
  <<: (( &temporary ))
  map1:
    a: a
    b: b
    c:
      a: ca
      b: cb
    d: d

  map2:
    a: 2a
    c:
      a: 2ca
      c: 2cc
    d:
      a: 2da
    e: 2e

merge: (( deepmerge(in.map1,in.map2) ))
`)
		resolved := parseYAML(`
---
merge:
  a: 2a
  b: b
  c:
    a: 2ca
    b: cb
    c: 2cc
  d:
    a: 2da
  e: 2e
`)
		Expect(source).To(FlowAs(resolved))
	})

	It("handles multiple map list", func() {
		source := parseYAML(`
---
in:
  <<: (( &temporary ))
  map1:
    a: a
    b: b
    c:
      a: ca
      b: cb
    d: d

  map2:
    a: 2a
    c:
      a: 2ca
      c: 2cc
    d:
      a: 2da
    e: 2e
  list:
    - (( map1 ))
    - (( map2 ))

merge: (( deepmerge(in.list) ))
`)
		resolved := parseYAML(`
---
merge:
  a: 2a
  b: b
  c:
    a: 2ca
    b: cb
    c: 2cc
  d:
    a: 2da
  e: 2e
`)
		Expect(source).To(FlowAs(resolved))
	})
})
