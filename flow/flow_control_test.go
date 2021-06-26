package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/spiff/features"
)

var _ = Describe("yaml control", func() {

	Context("join", func() {
		It("handles nil", func() {
			source := parseYAML(`
---
map:
  <<join: (( ~ ))
  alice: 25
`)
			resolved := parseYAML(`
---
map:
  alice: 25
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
		It("handles undef", func() {
			source := parseYAML(`
---
map:
  <<join: (( ~~ ))
  alice: 25
`)
			resolved := parseYAML(`
---
map:
  alice: 25
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
		It("handles map", func() {
			source := parseYAML(`
---
map:
  <<join: 
    bob: 26
  alice: 25
`)
			resolved := parseYAML(`
---
map:
  alice: 25
  bob: 26
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
		It("handles list", func() {
			source := parseYAML(`
---
map:
  <<join: 
    - bob: 26
    - charlie: 27
  alice: 25
`)
			resolved := parseYAML(`
---
map:
  alice: 25
  bob: 26
  charlie: 27
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
		It("handles override", func() {
			source := parseYAML(`
---
map:
  <<join: 
    - bob: 26
      charlie: 1
    - charlie: 27
  alice: 25
  charlie: 2
`)
			resolved := parseYAML(`
---
map:
  alice: 25
  bob: 26
  charlie: 27
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
	})

	////////////////////////////////////////////////////////////////////////////////

	Context("switch", func() {
		It("handles key", func() {
			source := parseYAML(`
---
key: (( x ))
x: test
selected:
  <<switch: (( key ))
  test: alice
  <<default: bob
`)
			resolved := parseYAML(`
---
key: test
x: test
selected: alice
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
		It("handles default", func() {
			source := parseYAML(`
---
key: (( x ))
x: other
selected:
  <<switch: (( key ))
  test: alice
  <<default: bob
`)
			resolved := parseYAML(`
---
key: other
x: other
selected: bob
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
	})

	Context("type switch", func() {
		It("handles key", func() {
			source := parseYAML(`
---
key: (( x ))
x: test
selected:
  <<type: (( key ))
  string: stringtype
  <<default: unknown
`)
			resolved := parseYAML(`
---
key: test
x: test
selected: stringtype
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
		It("handles default", func() {
			source := parseYAML(`
---
key: (( [x] ))
x: other
selected:
  <<type: (( key ))
  test: stringtype
  <<default: unknown
`)
			resolved := parseYAML(`
---
key: 
- other
x: other
selected: unknown
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
	})

	Context("if", func() {
		It("handles then", func() {
			source := parseYAML(`
---
x: test
cond:
  <<if: (( x == "test" ))
  <<then: alice
  <<else: bob
`)
			resolved := parseYAML(`
---
x: test
cond: alice
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
		It("handles else", func() {
			source := parseYAML(`
---
x: test1
cond:
  <<if: (( x == "test" ))
  <<then: alice
  <<else: bob
`)
			resolved := parseYAML(`
---
x: test1
cond: bob
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
		It("handles only one case", func() {
			source := parseYAML(`
---
x: test1
cond:
  <<if: (( x == "test" ))
  <<then: alice
`)
			resolved := parseYAML(`
---
x: test1
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
	})

	Context("for", func() {
		It("handles map", func() {
			source := parseYAML(`
---
x: suffix
alice:
       - a
       - b
bob:
       - 1
       - 2
       - 3
map:
  <<for: 
     alice: (( .alice ))
     bob: (( .bob ))
  <<mapkey: (( alice bob ))
  <<do:
    value: (( alice bob x ))

`)
			resolved := parseYAML(`
---
alice:
- a
- b
bob:
- 1
- 2
- 3
map:
  a1:
    value: a1suffix
  a2:
    value: a2suffix
  a3:
    value: a3suffix
  b1:
    value: b1suffix
  b2:
    value: b2suffix
  b3:
    value: b3suffix
x: suffix
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
		It("handles list", func() {
			source := parseYAML(`
---
x: suffix
alice:
       - a
       - b
bob:
       - 1
       - 2
       - 3
list:
  <<for: 
     alice: (( .alice ))
     bob: (( .bob ))
  <<do:
    value: (( alice bob x ))

`)
			resolved := parseYAML(`
---
alice:
- a
- b
bob:
- 1
- 2
- 3
list:
- value: a1suffix
- value: a2suffix
- value: a3suffix
- value: b1suffix
- value: b2suffix
- value: b3suffix
x: suffix
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})

		It("handles control variable list", func() {
			source := parseYAML(`
---
x: suffix
alice:
       - a
       - b
bob:
       - 1
       - 2
       - 3
list:
  <<for: 
     - name: bob
       values: (( .bob ))
     - name: alice
       values: (( .alice ))
  <<do:
    value: (( alice bob x ))

`)
			resolved := parseYAML(`
---
alice:
- a
- b
bob:
- 1
- 2
- 3
list:
- value: a1suffix
- value: b1suffix
- value: a2suffix
- value: b2suffix
- value: a3suffix
- value: b3suffix
x: suffix
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
		It("handles iteration index", func() {
			source := parseYAML(`
---
alice:
       - a
       - b
bob:
       - 1
       - 2
       - 3
list:
  <<for: 
     alice: (( .alice ))
     bob: (( .bob ))
  <<do:
    value: (( alice "-" index-alice "-" bob "-" index-bob ))

`)
			resolved := parseYAML(`
---
alice:
- a
- b
bob:
- 1
- 2
- 3
list:
- value: a-0-1-0
- value: a-0-2-1
- value: a-0-3-2
- value: b-1-1-0
- value: b-1-2-1
- value: b-1-3-2
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})

	})

	////////////////////////////////////////////////////////////////////////////

	Context("cascade controls", func() {
		It("switch handles key", func() {
			source := parseYAML(`
---
key: (( x ))
x: test
selected:
  <<switch: (( key ))
  test: alice
  <<default: bob
`)
			stub := parseYAML(`
---
selected: peter
`)
			resolved := parseYAML(`
---
key: test
x: test
selected: peter
`)
			Expect(source).To(FlowAs(resolved, stub).WithFeatures(features.CONTROL))
		})

		It("join handles if", func() {
			source := parseYAML(`
---
x: charlie
map:
  <<join: 
    - <<if: (( x == "charlie" ))
      <<then:
        charlie: 27
    - <<if: (( x == "alice" ))
      <<then:
        alice: 20
  alice: 25
  charlie: 2
`)
			resolved := parseYAML(`
---
x: charlie
map:
  alice: 25
  charlie: 27
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})

		It("propgates flags", func() {
			source := parseYAML(`
---
temp:
  <<: (( &temporary ))
  <<if: (( features("control") ))
  <<then:
    alice: 25
    bob: 26

final: (( temp ))
`)
			resolved := parseYAML(`
---
final:
  alice: 25
  bob: 26
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
	})

	Context("controls in lists", func() {
		It("handles simple values", func() {
			source := parseYAML(`
---
x: test
list:
- alice
- <<if: (( x == "test" ))
  <<then: peter
- bob
`)
			resolved := parseYAML(`
---
x: test
list:
- alice
- peter
- bob
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
		It("omits undef", func() {
			source := parseYAML(`
---
x: test
list:
- alice
- <<if: (( x != "test" ))
  <<then: peter
- bob
`)
			resolved := parseYAML(`
---
x: test
list:
- alice
- bob
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
		It("inserts lists", func() {
			source := parseYAML(`
---
list:
- alice
- <<if: (( features("control") ))
  <<then:
  - peter
  - alex
- bob
`)
			resolved := parseYAML(`
---
list:
- alice
- peter
- alex
- bob
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})

		It("example", func() {
			source := parseYAML(`
---
list:
- <<if: (( features("control") ))
  <<then: alice
- <<if: (( features("control") ))
  <<then:
  - - peter
- <<if: (( features("control") ))
  <<then:
  - bob
`)
			resolved := parseYAML(`
---
list:
- alice
- - peter
- bob
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
	})
})
