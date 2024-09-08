package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("merging lists", func() {
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
})
