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
   - <<<: (( merge ))
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
