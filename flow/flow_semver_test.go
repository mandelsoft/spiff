package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("semver functions", func() {
	It("normalize", func() {
		source := parseYAML(`
---
normalized: (( semver("v1.2-beta.1+demo") ))
`)
		resolved := parseYAML(`
---
normalized: 1.2.0-beta.1+demo
`)
		Expect(source).To(FlowAs(resolved))
	})
	Context("attributes", func() {
		It("handles release", func() {
			source := parseYAML(`
---
attr: (( semverrelease("1.2.3-beta.1") ))
`)
			resolved := parseYAML(`
---
attr: 1.2.3
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("handles major", func() {
			source := parseYAML(`
---
attr: (( semvermajor("1.2.3-beta.1") ))
`)
			resolved := parseYAML(`
---
attr: 1
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("handles minor", func() {
			source := parseYAML(`
---
attr: (( semverminor("1.2.3-beta.1") ))
`)
			resolved := parseYAML(`
---
attr: 2
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("handles patch", func() {
			source := parseYAML(`
---
attr: (( semverpatch("1.2.3-beta.1") ))
`)
			resolved := parseYAML(`
---
attr: 3
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("handles prerelease", func() {
			source := parseYAML(`
---
attr: (( semverprerelease("1.2.3-beta.1") ))
`)
			resolved := parseYAML(`
---
attr: beta.1
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("handles metadata", func() {
			source := parseYAML(`
---
attr: (( semvermetadata("1.2.3-beta.1+demo") ))
`)
			resolved := parseYAML(`
---
attr: demo
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("handles release", func() {
			source := parseYAML(`
---
attr: (( semverrelease("1.2.3-beta.1+demo") ))
`)
			resolved := parseYAML(`
---
attr: 1.2.3
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("sort", func() {
		It("sorts list", func() {
			source := parseYAML(`
---
versions:
  - "1.2.3"
  - "1.3.4"
  - "2.0"
  - "1.2.5"
  - "1.3.4-beta.1"
sorted: (( semversort(versions) ))
`)
			resolved := parseYAML(`
---
versions:
  - "1.2.3"
  - "1.3.4"
  - "2.0"
  - "1.2.5"
  - "1.3.4-beta.1"
sorted:
  - "1.2.3"
  - "1.2.5"
  - "1.3.4-beta.1"
  - "1.3.4"
  - "2.0"
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("sorts args", func() {
			source := parseYAML(`
---
versions:
  - "1.2.3"
  - "1.3.4"
  - "2.0"
  - "1.2.5"
  - "1.3.4-beta.1"
sorted: (( semversort(versions...) ))
`)
			resolved := parseYAML(`
---
versions:
  - "1.2.3"
  - "1.3.4"
  - "2.0"
  - "1.2.5"
  - "1.3.4-beta.1"
sorted:
  - "1.2.3"
  - "1.2.5"
  - "1.3.4-beta.1"
  - "1.3.4"
  - "2.0"
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("constraints", func() {
		It("match semver", func() {
			source := parseYAML(`
---
match: (( semvermatch("1.2.3") ))
`)
			resolved := parseYAML(`
---
match: true
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("match wrong semver", func() {
			source := parseYAML(`
---
match: (( catch(semvermatch("1.2.x")) ))
`)
			resolved := parseYAML(`
---
match:
  error: "semvermatch: \"1.2.x\": Invalid Semantic Version"
  valid: false
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("match single constraint", func() {
			source := parseYAML(`
---
match: (( semvermatch("1.2.3", "~1.2") ))
`)
			resolved := parseYAML(`
---
match: true
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("match two constraints", func() {
			source := parseYAML(`
---
match: (( semvermatch("1.2.3", "~1.2", "<1.3") ))
`)
			resolved := parseYAML(`
---
match: true
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("fail second constraint", func() {
			source := parseYAML(`
---
match: (( semvermatch("1.2.3", "~1.2", "<1.2.3") ))
`)
			resolved := parseYAML(`
---
match: false
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("validate", func() {
		It("validate semver", func() {
			source := parseYAML(`
---
match: (( validate("1.2.3", "semver", [ "semver", "~1.2"]) ))
`)
			resolved := parseYAML(`
---
match: 1.2.3
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("validate non semver", func() {
			source := parseYAML(`
---
match: (( catch(validate("1.2.x", "semver")) ))
`)
			resolved := parseYAML(`
---
match:
  error: 'condition 1 failed: semver: "1.2.x": Invalid Semantic Version'
  valid: false
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("validate invalid constraint", func() {
			source := parseYAML(`
---
match: (( catch(validate("1.2.3", ["semver", "~1.3"])) ))
`)
			resolved := parseYAML(`
---
match:
  error: 'condition 1 failed: [1.2.3 is less than 1.3]'
  valid: false
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("compare", func() {
		It("compares semver", func() {
			source := parseYAML(`
---
match: (( semvercmp("1.2.3", "2.1") ))
`)
			resolved := parseYAML(`
---
match: -1
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("compares versions", func() {
			source := parseYAML(`
---
match: (( semvercmp("1.2.3", "1.2.3-beta.1") ))
`)
			resolved := parseYAML(`
---
match: 1
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("equals semver", func() {
			source := parseYAML(`
---
match: (( semvercmp("1.2.3", "1.2.3") ))
`)
			resolved := parseYAML(`
---
match: 0
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("equals flavors", func() {
			source := parseYAML(`
---
match: (( semvercmp("1.2", "1.2.0") ))
`)
			resolved := parseYAML(`
---
match: 0
`)
			Expect(source).To(FlowAs(resolved))
		})

	})
	Context("modification", func() {
		It("increment major", func() {
			source := parseYAML(`
---
new: (( semverincmajor("1.2.3") ))
`)
			resolved := parseYAML(`
---
new: 2.0.0
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("increment minor", func() {
			source := parseYAML(`
---
new: (( semverincminor("v1.2.3") ))
`)
			resolved := parseYAML(`
---
new: v1.3.0
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("increment minor with suffix", func() {
			source := parseYAML(`
---
new: (( semverincminor("v1.2.3-beta.1+demo") ))
`)
			resolved := parseYAML(`
---
new: v1.3.0
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("increment patch", func() {
			source := parseYAML(`
---
new: (( semverincpatch("v1.2.3") ))
`)
			resolved := parseYAML(`
---
new: v1.2.4
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("increment patch with suffix", func() {
			source := parseYAML(`
---
final: (( semverincpatch("1.2.3-beta.1+demo") ))
new: (( semverincpatch(final) ))
`)
			resolved := parseYAML(`
---
final: 1.2.3
new: 1.2.4
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("set meta", func() {
			source := parseYAML(`
---
new: (( semvermetadata("1.2.3", "test") ))
`)
			resolved := parseYAML(`
---
new: 1.2.3+test
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("replace meta", func() {
			source := parseYAML(`
---
new: (( semvermetadata("v1.2.3+demo", "test") ))
`)
			resolved := parseYAML(`
---
new: v1.2.3+test
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("remove meta", func() {
			source := parseYAML(`
---
new: (( semvermetadata("v1.2.3-beta+demo", "") ))
`)
			resolved := parseYAML(`
---
new: v1.2.3-beta
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("set prerelease", func() {
			source := parseYAML(`
---
new: (( semverprerelease("1.2.3", "beta") ))
`)
			resolved := parseYAML(`
---
new: 1.2.3-beta
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("replace prerelease", func() {
			source := parseYAML(`
---
new: (( semverprerelease("v1.2.3-demo", "beta") ))
`)
			resolved := parseYAML(`
---
new: v1.2.3-beta
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("remove prerelease", func() {
			source := parseYAML(`
---
new: (( semverprerelease("v1.2.3-beta+demo", "") ))
`)
			resolved := parseYAML(`
---
new: v1.2.3+demo
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("replace release", func() {
			source := parseYAML(`
---
new: (( semverrelease("v1.2.3-demo", "1.2.1+test") ))
`)
			resolved := parseYAML(`
---
new: 1.2.1-demo
`)
			Expect(source).To(FlowAs(resolved))
		})
	})
})
