package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/spiff/features"
)

var _ = Describe("yaml control", func() {

	Context("merge", func() {
		It("handles nil", func() {
			source := parseYAML(`
---
map:
  <<merge: (( ~ ))
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
  <<merge: (( ~~ ))
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
  <<merge: 
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
  <<merge: 
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
  <<merge: 
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
		It("handles nil", func() {
			source := parseYAML(`
---
key: ~
selected:
  <<switch: (( key ))
  test: alice
  <<default: bob
`)
			resolved := parseYAML(`
---
key: ~
selected: bob
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
		It("handles undef", func() {
			source := parseYAML(`
---
key: ~
selected:
  <<switch: (( ~~ ))
  test: alice
  <<default: bob
`)
			resolved := parseYAML(`
---
key: ~
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
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
		It("resolve to undef case", func() {
			source := parseYAML(`
---
key: alice
selected:
  <<switch: (( key ))
  alice: (( ~~ ))
  bob: 26
  <<default: unknown
`)
			resolved := parseYAML(`
---
key: alice
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})

		It("resolve as template", func() {
			source := parseYAML(`
---
selected:
  <<: (( &template &local ))
  <<switch: (( __ctx.PATH[0] ))
  alice: (( ~~ ))
  bob: 26
  <<default: unknown

alice: (( *selected ))
bob: (( *selected ))
`)
			resolved := parseYAML(`
---
bob: 26
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})

		It("resolve as template", func() {
			source := parseYAML(`
---
temp:
  <<: (( &temporary ))
  func: (( |name|->*_.selected ))
  selected:
    <<: (( &template ))
    <<switch: (( name ))
    alice: (( ~~ ))
    bob: 26
    <<default: unknown

list:
- (( temp.func("alice") ))
- (( temp.func("bob") ))
- (( temp.func("charlie") ))
`)
			resolved := parseYAML(`
---
list:
  - 26
  - unknown
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})

		//////////////////////

		It("handles cases for nil", func() {
			source := parseYAML(`
---
key: ~
selected:
  <<switch: (( key ))
  <<cases:
  - case: test
    value: alice
  <<default: bob
`)
			resolved := parseYAML(`
---
key: ~
selected: bob
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
		It("handles cases for undef", func() {
			source := parseYAML(`
---
key: ~
selected:
  <<switch: (( ~~ ))
  <<cases:
  - case: test
    value: alice
  <<default: bob
`)
			resolved := parseYAML(`
---
key: ~
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
		It("handles cases key", func() {
			source := parseYAML(`
---
key: (( x ))
x: test
selected:
  <<switch: (( key ))
  <<cases:
  - case: test
    value: alice
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
		It("handles cases match", func() {
			source := parseYAML(`
---
key: (( x ))
x: test
selected:
  <<switch: (( key ))
  <<cases:
  - match: (( |c|-> c == "test" ))"
    value: alice
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
		It("handles cases default", func() {
			source := parseYAML(`
---
key: (( x ))
x: other
selected:
  <<switch: (( key ))
  <<cases:
  - case: test
    value: alice
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
		It("resolve to cases undef case", func() {
			source := parseYAML(`
---
key: alice
selected:
  <<switch: (( key ))
  <<cases:
  - case: alice
    value: (( ~~ ))
  - case: bob
    value: 26
  <<default: unknown
`)
			resolved := parseYAML(`
---
key: alice
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
	})

	////////////////////////////////////////////////////////////////////////////////

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
		It("handles missing case", func() {
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
		It("handles undef case", func() {
			source := parseYAML(`
---
x: test
mode: undef
cond:
  <<if: (( x == "test" ))
  <<then: (( mode == "undef" ? ~~ :"yep" ))
`)
			resolved := parseYAML(`
---
x: test
mode: undef
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
	})

	Context("for", func() {
		Context("lists", func() {
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
			It("filters list by undef", func() {
				source := parseYAML(`
---
bob:
       - 1
       - 2
       - 3
filtered:
  <<for: 
     bob: (( .bob ))
  <<do: (( bob == 2 ? ~~ :bob ))
`)
				resolved := parseYAML(`
---
bob:
- 1
- 2
- 3
filtered:
- 1
- 3 
`)
				Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
			})
			It("filters list keeping nil value", func() {
				source := parseYAML(`
---
bob:
       - 1
       - 2
       - 3
filtered:
  <<for: 
     bob: (( .bob ))
  <<do: (( bob == 2 ? ~ :bob ))
`)
				resolved := parseYAML(`
---
bob:
- 1
- 2
- 3
filtered:
- 1
- ~
- 3 
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

			It("handles control variable map", func() {
				source := parseYAML(`
---
alice:
       - a
       - b
bob:
       - 1
       - 2
       - 3
map:
  <<for: 
     key,alice: (( .alice ))
     bob: (( .bob ))
  <<mapkey: (( alice bob ))
  <<do:
    value: (( alice bob "/" key "/" index-bob ))

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
    value: a1/0/0
  a2:
    value: a2/0/1
  a3:
    value: a3/0/2
  b1:
    value: b1/1/0
  b2:
    value: b2/1/1
  b3:
    value: b3/1/2
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

		Context("maps", func() {
			It("filters map by value", func() {
				source := parseYAML(`
---
bob:
  b1: 1
  b2: 2
  b3: 3
filtered:
  <<for: 
     key,bob: (( .bob ))
  <<mapkey: (( key ))
  <<do: (( bob == 2 ? ~~ :bob ))
`)
				resolved := parseYAML(`
---
bob:
  b1: 1
  b2: 2
  b3: 3
filtered:
  b1: 1
  b3: 3
`)
				Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
			})
			It("filters map by undef key", func() {
				source := parseYAML(`
---
bob:
  b1: 1
  b2: 2
  b3: 3
filtered:
  <<for: 
     key,bob: (( .bob ))
  <<mapkey: (( bob == 2 ? ~~ :key ))
  <<do: (( bob ))
`)
				resolved := parseYAML(`
---
bob:
  b1: 1
  b2: 2
  b3: 3
filtered:
  b1: 1
  b3: 3
`)
				Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
			})
			It("filters map by nil key", func() {
				source := parseYAML(`
---
bob:
  b1: 1
  b2: 2
  b3: 3
filtered:
  <<for: 
     key,bob: (( .bob ))
  <<mapkey: (( bob == 2 ? ~ :key ))
  <<do: (( bob ))
`)
				resolved := parseYAML(`
---
bob:
  b1: 1
  b2: 2
  b3: 3
filtered:
  b1: 1
  b3: 3
`)
				Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
			})
			It("handles map", func() {
				source := parseYAML(`
---
x: suffix
alice:
  a1: a
  a2: b
bob:
  b1: 1
  b2: 2
  b3: 3
map:
  <<for: 
     alice: (( .alice ))
     bob: (( .bob ))
  <<mapkey: (( index-alice index-bob ))
  <<do:
    value: (( alice bob x ))

`)
				resolved := parseYAML(`
---
alice:
  a1: a
  a2: b
bob:
  b1: 1
  b2: 2
  b3: 3
map:
  a1b1:
    value: a1suffix
  a1b2:
    value: a2suffix
  a1b3:
    value: a3suffix
  a2b1:
    value: b1suffix
  a2b2:
    value: b2suffix
  a2b3:
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
  a1: a
  a2: b
bob:
  b1: 1
  b2: 2
  b3: 3
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
  a1: a
  a2: b
bob:
  b1: 1
  b2: 2
  b3: 3
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
			It("handles control variable list with index name", func() {
				source := parseYAML(`
---
alice:
  a1: a
  a2: b
bob:
  b1: 1
  b2: 2
  b3: 3
list:
  <<for: 
     - name: bob
       values: (( .bob ))
       index: key
     - name: alice
       values: (( .alice ))
  <<do:
    value: (( alice key ":" bob  ))

`)
				resolved := parseYAML(`
---
alice:
  a1: a
  a2: b
bob:
  b1: 1
  b2: 2
  b3: 3
list:
- value: ab1:1
- value: bb1:1
- value: ab2:2
- value: bb2:2
- value: ab3:3
- value: bb3:3
`)
				Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
			})
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

		It("merge handles if", func() {
			source := parseYAML(`
---
x: charlie
map:
  <<merge: 
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

	////////////////////////////////////////////////////////////////////////////////

	Context("cascading", func() {
		It("overrides", func() {
			source := parseYAML(`
---

selected:
  <<switch: alice
  alice: (( bob + 1 ))
  bob: 25
`)

			stub := parseYAML(`
---
selected:
   bob: 1
`)

			resolved := parseYAML(`
---
selected: 26
`)
			Expect(source).To(CascadeAs(resolved).WithFeatures(features.CONTROL))
			resolved = parseYAML(`
---
selected:
  bob: 1
`)
			Expect(source).To(CascadeAs(resolved, stub).WithFeatures(features.CONTROL))
		})
	})

	////////////////////////////////////////////////////////////////////////////////

	Context("error", func() {
		It("ignores unused error nodes", func() {
			source := parseYAML(`
---
x: test
cond:
  <<if: (( x == "test" ))
  <<then: alice
  <<else: (( 1 / 0 ))
`)
			resolved := parseYAML(`
---
x: test
cond: alice
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
		It("fails for used error nodes", func() {
			source := parseYAML(`
---
x: test1
cond:
  <<if: (( x == "test" ))
  <<then: alice
  <<else: (( 1 / 0 ))
`)
			Expect(source).To(FlowToErr("\t(( 1 / 0 ))\tin test\tcond\t(...<<else)\t*division by zero").WithFeatures(features.CONTROL))
		})
		It("fails for used nested error nodes", func() {
			source := parseYAML(`
---
x: test1
cond:
  <<if: (( x == "test" ))
  <<then: alice
  <<else:
    nested: (( 1 / 0 ))
`)
			Expect(source).To(FlowToErr("\t(( 1 / 0 ))\tin test\tcond.nested\t(cond.<<else.nested)\t*division by zero").WithFeatures(features.CONTROL))
		})
		It("fails for missing cases case", func() {
			source := parseYAML(`
---
x: test
selected:
  <<switch: (( x ))
  <<cases:
  - value: alice 
`)
			Expect(source).To(FlowToErr("\t<switch control>\tin test\tselected\t()\t*case 0 requires 'case' or `'match' field").WithFeatures(features.CONTROL))
		})
		It("fails for used nested cases error nodes", func() {
			source := parseYAML(`
---
x: test
selected:
  <<switch: (( x ))
  <<cases:
  - case: test
    value:
      nested: (( 1 / 0 ))
  - case: other
    value:
      other: (( 1 / 0 ))
`)
			Expect(source).To(FlowToErr("\t(( 1 / 0 ))\tin test\tselected.nested\t(selected.<<cases.[0].value.nested)\t*division by zero").WithFeatures(features.CONTROL))
		})

	})
})
