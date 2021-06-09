package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Template Auto Substitution", func() {
	Context("Local substitution", func() {
		It("locally substitutes template", func() {
			source := parseYAML(`
---
id: (( &inject &dynamic(template) ))
template: (( &temporary &template(__ctx.FILE) ))
`, "SOURCE")

			resolved := parseYAML(`
---
id: SOURCE
`)

			Expect(source).To(CascadeAs(resolved))
		})

		It("locally substitutes direct template", func() {
			source := parseYAML(`
---
id: (( &inject &dynamic &template(__ctx.FILE) ))
`, "SOURCE")

			resolved := parseYAML(`
---
id: SOURCE
`)

			Expect(source).To(CascadeAs(resolved))
		})
	})

	Context("Injection", func() {
		It("injects substituted templates", func() {
			source := parseYAML(`
---
data: 1
`, "SOURCE")

			stub := parseYAML(`
---
id: (( &inject &dynamic(template) ))
template: (( &template(__ctx.FILE) ))
`, "STUB")

			resolved := parseYAML(`
---
id: SOURCE
data: 1
`)

			Expect(source).To(CascadeAs(resolved, stub))
		})

		It("injects substituted direct templates", func() {
			source := parseYAML(`
---
data: 1
`, "SOURCE")

			stub := parseYAML(`
---
id: (( &inject &dynamic &template(__ctx.FILE) ))
`, "STUB")

			resolved := parseYAML(`
---
id: SOURCE
data: 1
`)

			Expect(source).To(CascadeAs(resolved, stub))
		})

		It("injects through multiple stubs", func() {
			source := parseYAML(`
---
data: 1
`, "SOURCE")

			secondary := parseYAML(`
---
data: 2
`, "SECOND")

			stub := parseYAML(`
---
id: (( &inject &dynamic(template) ))
template: (( &template(__ctx.FILE) ))
`, "STUB")

			resolved := parseYAML(`
---
id: SOURCE
data: 2
`)

			Expect(source).To(CascadeAs(resolved, secondary, stub))
		})

		It("injects direct substitution through multiple stubs", func() {
			source := parseYAML(`
---
data: 1
`, "SOURCE")

			secondary := parseYAML(`
---
data: 2
`, "SECOND")

			stub := parseYAML(`
---
id: (( &inject &dynamic &template(__ctx.FILE) ))
`, "STUB")

			resolved := parseYAML(`
---
id: SOURCE
data: 2
`)

			Expect(source).To(CascadeAs(resolved, secondary, stub))
		})
	})

	////////////////////////////////////////////////////////////////////////////////

	Context("Override", func() {
		It("overrides substituted templates", func() {
			source := parseYAML(`
---
data: 1
id: wrong
`, "SOURCE")

			stub := parseYAML(`
---
id: (( &dynamic(template) ))
template: (( &template(__ctx.FILE) ))
`, "STUB")

			resolved := parseYAML(`
---
id: SOURCE
data: 1
`)

			Expect(source).To(CascadeAs(resolved, stub))
		})
		It("overrides substituted direct templates", func() {
			source := parseYAML(`
---
data: 1
id: wrong
`, "SOURCE")

			stub := parseYAML(`
---
id: (( &dynamic &template(__ctx.FILE) ))
`, "STUB")

			resolved := parseYAML(`
---
id: SOURCE
data: 1
`)

			Expect(source).To(CascadeAs(resolved, stub))
		})

	})

	It("overrides substitutions through multiple stubs", func() {
		source := parseYAML(`
---
data: 1
id: wrong
`, "SOURCE")

		secondary := parseYAML(`
---
data: 2
`, "SECOND")

		stub := parseYAML(`
---
id: (( &dynamic(template) ))
template: (( &template(__ctx.FILE) ))
`, "STUB")

		resolved := parseYAML(`
---
id: SOURCE
data: 2
`)

		Expect(source).To(CascadeAs(resolved, secondary, stub))
	})

	It("overrides direct substitution through multiple stubs", func() {
		source := parseYAML(`
---
data: 1
id: wrong
`, "SOURCE")

		secondary := parseYAML(`
---
data: 2
`, "SECOND")

		stub := parseYAML(`
---
id: (( &dynamic &template(__ctx.FILE) ))
`, "STUB")

		resolved := parseYAML(`
---
id: SOURCE
data: 2
`)

		Expect(source).To(CascadeAs(resolved, secondary, stub))
	})

	////////////////////////////////////////////////////////////////////////////////

	Context("Default", func() {
		It("is ignored if already set", func() {
			source := parseYAML(`
---
data: 1
id: own
`, "SOURCE")

			stub := parseYAML(`
---
id: (( &default &dynamic(template) ))
template: (( &template(__ctx.FILE) ))
`, "STUB")

			resolved := parseYAML(`
---
id: own
data: 1
`)

			Expect(source).To(CascadeAs(resolved, stub))
		})
		It("defaults substituted templates", func() {
			source := parseYAML(`
---
data: 1
`, "SOURCE")

			stub := parseYAML(`
---
id: (( &default &dynamic(template) ))
template: (( &template(__ctx.FILE) ))
`, "STUB")

			resolved := parseYAML(`
---
id: SOURCE
data: 1
`)

			Expect(source).To(CascadeAs(resolved, stub))
		})
		It("defaults substituted direct templates", func() {
			source := parseYAML(`
---
data: 1
`, "SOURCE")

			stub := parseYAML(`
---
id: (( &default &dynamic &template(__ctx.FILE) ))
`, "STUB")

			resolved := parseYAML(`
---
id: SOURCE
data: 1
`)

			Expect(source).To(CascadeAs(resolved, stub))
		})

	})

	It("defaults substitutions through multiple stubs", func() {
		source := parseYAML(`
---
data: 1
`, "SOURCE")

		secondary := parseYAML(`
---
data: 2
`, "SECOND")

		stub := parseYAML(`
---
id: (( &default &dynamic(template) ))
template: (( &template(__ctx.FILE) ))
`, "STUB")

		resolved := parseYAML(`
---
id: SOURCE
data: 2
`)

		Expect(source).To(CascadeAs(resolved, secondary, stub))
	})

	It("defaults direct substitution through multiple stubs", func() {
		source := parseYAML(`
---
data: 1
`, "SOURCE")

		secondary := parseYAML(`
---
data: 2
`, "SECOND")

		stub := parseYAML(`
---
id: (( &default &dynamic &template(__ctx.FILE) ))
`, "STUB")

		resolved := parseYAML(`
---
id: SOURCE
data: 2
`)

		Expect(source).To(CascadeAs(resolved, secondary, stub))
	})

	Context("temporary", func() {
		It("locally substitutes template", func() {
			source := parseYAML(`
---
data: 1
`, "SOURCE")

			stub := parseYAML(`
---
id: (( &temporary &inject &dynamic(template) ))
template: (( &temporary &template(__ctx.FILE) ))
`, "STUB")

			resolved := parseYAML(`
---
data: 1
`)

			Expect(source).To(CascadeAs(resolved, stub))
		})
	})

	It("local temporary", func() {
		source := parseYAML(`
---
id: (( &temporary &inject &dynamic &template(__ctx.FILE) ))
data: (( id ))
`, "SOURCE")

		resolved := parseYAML(`
---
data: SOURCE
`)

		Expect(source).To(CascadeAs(resolved))
	})
})
