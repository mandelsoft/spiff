package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("merging lists", func() {
	Context("unkeyed lists", func() {
		It("merges map lists", func() {
			source := parseYAML(`
---
list:
   - field: e
     attr: f
   - field: g
     attr: h
   - <<<: (( merge ))
   - field: i
     attr: j
   - field: k
     attr: l
`)
			stub := parseYAML(`
---
list:
  - field: a
    attr: b
  - field: c
    attr: d
`)
			resolved := parseYAML(`
---
list:
  - field: e
    attr: f
  - field: g
    attr: h
  - field: a
    attr: b
  - field: c
    attr: d
  - field: i
    attr: j
  - field: k
    attr: l
`)
			Expect(source).To(FlowAs(resolved, stub))
		})

		It("merges map lists with two inserts", func() {
			source := parseYAML(`
---
temp:
- <<: (( &temporary ))
- field: t1
  attr: vt1
- field: t2
  attr: vt2

list:
   - field: a1
     attr: va1
   - <<<: (( merge ))
   - field: a2
     attr: va2
   - <<: (( temp ))
`)
			stub := parseYAML(`
---
list:
  - field: b1
    attr: vb1
  - field: b2
    attr: vb2
`)
			resolved := parseYAML(`
---
list:
  - field: a1
    attr: va1
  - field: b1
    attr: vb1
  - field: b2
    attr:  vb2
  - field: a2
    attr: va2
  - field: t1
    attr: vt1
  - field: t2
    attr: vt2
`)
			Expect(source).To(FlowAs(resolved, stub))
		})

		It("merges map lists with marker", func() {
			source := parseYAML(`
---
list : (( temp ))
temp:
   - field: e
     attr: f
   - <<<: (( &temporary(merge) ))
`)
			stub := parseYAML(`
---
temp:
  - field: a
    attr: b
  - field: c
    attr: d
`)
			resolved := parseYAML(`
---
list:
  - field: e
    attr: f
  - field: a
    attr: b
  - field: c
    attr: d
`)
			Expect(source).To(FlowAs(resolved, stub))
		})

		It("merges map lists with complex merge", func() {
			source := parseYAML(`
---
list:
   - field: e
     attr: f
   - <<<: (( ( merge || ~ ) ))
`)
			stub := parseYAML(`
---
list:
  - field: a
    attr: b
  - field: c
    attr: d
`)
			resolved := parseYAML(`
---
list:
  - field: e
    attr: f
  - field: a
    attr: b
  - field: c
    attr: d
`)
			Expect(source).To(FlowAs(resolved, stub))
		})
	})

	Context("keyed lists", func() {
		It("merges existing and adds new", func() {
			source := parseYAML(`
---
list:
   - name: a
     attr: b
   - <<<: (( merge ))
`)
			stub := parseYAML(`
---
list:
  - name: c
    attr: d
  - name: a
    attr: e
`)
			resolved := parseYAML(`
---
list:
  - name: a
    attr: e
  - name: c
    attr: d
`)
			Expect(source).To(FlowAs(resolved, stub))
		})

		It("merges existing and adds new with marker", func() {
			source := parseYAML(`
---
list: (( temp ))
temp:
   - name: a
     attr: b
   - <<<: (( &temporary(merge) ))
`)
			stub := parseYAML(`
---
temp:
  - name: c
    attr: d
  - name: a
    attr: e
`)
			resolved := parseYAML(`
---
list:
  - name: a
    attr: e
  - name: c
    attr: d
`)
			Expect(source).To(FlowAs(resolved, stub))
		})

		It("merges existing and adds new with complex merge", func() {
			source := parseYAML(`
---
list:
   - name: a
     attr: b
   - <<<: (( ( merge || ~ ) ))
`)
			stub := parseYAML(`
---
list:
  - name: c
    attr: d
  - name: a
    attr: e
`)
			resolved := parseYAML(`
---
list:
  - name: a
    attr: e
  - name: c
    attr: d
`)
			Expect(source).To(FlowAs(resolved, stub))
		})

		It("merges map lists with field merge", func() {
			source := parseYAML(`
---
list:
   - field: e
     attr: f
   - field: a
     attr: x
   - <<<: (( ( merge on field ) ))
`)
			stub := parseYAML(`
---
list:
  - field: a
    attr: b
  - field: c
    attr: d
`)
			resolved := parseYAML(`
---
list:
  - field: e
    attr: f
  - field: a
    attr: b
  - field: c
    attr: d
`)
			Expect(source).To(FlowAs(resolved, stub))
		})
	})
})
