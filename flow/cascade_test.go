package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cascading YAML templates", func() {
	It("flows through multiple templates", func() {
		source := parseYAML(`
---
foo: (( merge ))
baz: 42
`)

		secondary := parseYAML(`
---
foo:
  bar: (( merge ))
  xyz: (( bar ))
`)

		stub := parseYAML(`
---
foo:
  bar: merged!
`)

		resolved := parseYAML(`
---
foo:
  bar: merged!
  xyz: merged!
baz: 42
`)

		Expect(source).To(CascadeAs(resolved, secondary, stub))
	})

	Context("with multiple mutually-exclusive templates", func() {
		It("flows through both", func() {
			source := parseYAML(`
---
foo: (( merge ))
baz: (( merge ))
`)

			secondary := parseYAML(`
---
foo:
  bar: (( merge ))
`)

			tertiary := parseYAML(`
---
baz:
  a: 1
  b: (( merge ))
`)

			stub := parseYAML(`
---
foo:
  bar: merged!
baz:
  b: 2
`)

			resolved := parseYAML(`
---
foo:
  bar: merged!
baz:
  a: 1
  b: 2
`)

			Expect(source).To(CascadeAs(resolved, secondary, tertiary, stub))
		})
	})

	Describe("node annotation propagation", func() {

		Context("referencing a merged field", func() {
			It("is not handled as merge expression", func() {
				source := parseYAML(`
---
alice: (( merge ))
bob: (( alice ))
`)
				stub := parseYAML(`
---
alice: alice
bob: bob
`)
				resolved := parseYAML(`
---
alice: alice
bob: bob
`)
				Expect(source).To(CascadeAs(resolved, stub))
			})
		})
	})

	Describe("merging lists with specified key", func() {

		Context("auto merge with key tag", func() {
			It("overrides matching key entries", func() {
				source := parseYAML(`
---
list:
  - key:address: a
    attr: b
  - address: c
    attr: d
`)
				stub := parseYAML(`
---
list:
  - address: c
    attr: stub
  - address: e
    attr: f
`)
				resolved := parseYAML(`
---
list:
  - address: a
    attr: b
  - address: c
    attr: stub
`)
				Expect(source).To(CascadeAs(resolved, stub))
			})

			It("overrides matching key entries with key tag", func() {
				source := parseYAML(`
---
list:
  - key:address: a
    attr: b
  - address: c
    attr: d
`)
				stub := parseYAML(`
---
list:
  - key:address: c
    attr: stub
  - address: e
    attr: f
`)
				resolved := parseYAML(`
---
list:
  - address: a
    attr: b
  - address: c
    attr: stub
`)
				Expect(source).To(CascadeAs(resolved, stub))
			})
		})
	})

	Describe("using lambda expressions", func() {
		template := parseYAML(`
---
values: (( merge ))
`)

		Context("locally in a single file", func() {
			It("defines an inline lambda value", func() {
				source := parseYAML(`
---
lvalue: (( lambda |x,y|->x + y ))
values: (( "" lvalue ))
`)

				resolved := parseYAML(`
---
values: lambda|x,y|->x + y
`)
				Expect(template).To(CascadeAs(resolved, source))
			})

			It("defines an evaluated lambda value", func() {
				source := parseYAML(`
---
lvalue: (( lambda "|x,y|->x + y" ))
values: (( "" lvalue ))
`)

				resolved := parseYAML(`
---
values: lambda|x,y|->x + y
`)
				Expect(template).To(CascadeAs(resolved, source))
			})

			It("calls a lambda value by reference", func() {
				source := parseYAML(`
---
lvalue: (( lambda |x,y|->x + y ))
values: (( .lvalue(1,2) ))
`)

				resolved := parseYAML(`
---
values: 3
`)
				Expect(template).To(CascadeAs(resolved, source))
			})

			It("calls a lambda value by reference expression", func() {
				source := parseYAML(`
---
lvalue: (( lambda |x,y|->x + y ))
values: (( (lambda lvalue)(1,2) ))
`)

				resolved := parseYAML(`
---
values: 3
`)
				Expect(template).To(CascadeAs(resolved, source))
			})

			It("calls a lambda value by string expression", func() {
				source := parseYAML(`
---
values: (( (lambda "|x,y|->x + y")(1,2) ))
`)

				resolved := parseYAML(`
---
values: 3
`)
				Expect(template).To(CascadeAs(resolved, source))
			})

			It("calls a lambda value by lambda expression", func() {
				source := parseYAML(`
---
values: (( (lambda |x,y|->x + y)(1,2) ))
`)

				resolved := parseYAML(`
---
values: 3
`)
				Expect(template).To(CascadeAs(resolved, source))
			})

			It("resolves references relative to caller", func() {
				source := parseYAML(`
---
lvalue: (( lambda |x,y|->x + y + offset ))
offset: 0
values:
  offset: 3
  value: (( .lvalue(1,2) ))
`)

				resolved := parseYAML(`
---
values:
  offset: 3
  value: 6
`)
				Expect(template).To(CascadeAs(resolved, source))
			})

			It("passes lambda value as argument", func() {
				source := parseYAML(`
---
lvalue: (( lambda |x,y|->x + y ))
mod: (( lambda|x,y,m|->(lambda m)(x, y) + 3 ))
values:
  value: (( .mod(1,2, lvalue) ))
`)

				resolved := parseYAML(`
---
values:
  value: 6
`)
				Expect(template).To(CascadeAs(resolved, source))
			})

			It("passes binding to nested lambda expressions", func() {
				source := parseYAML(`
---
mult: (( lambda |x|-> lambda |y|-> x * y ))
mult2: (( .mult(2) ))
values:
  value: (( .mult2(3) ))
`)

				resolved := parseYAML(`
---
values:
  value: 6
`)
				Expect(template).To(CascadeAs(resolved, source))
			})

			It("supports self recursion", func() {
				source := parseYAML(`
---
fibonacci: (( lambda |x|-> x <= 0 ? 0 :x == 1 ? 1 :_(x - 2) + _( x - 1 ) ))
values:
  value: (( .fibonacci(5) ))
`)

				resolved := parseYAML(`
---
values:
  value: 5
`)
				Expect(template).To(CascadeAs(resolved, source))
			})

			It("supports currying", func() {
				source := parseYAML(`
---
mult: (( lambda |x,y|-> x * y ))
mult2: (( .mult(2) ))
values:
  value: (( .mult2(5) ))
`)

				resolved := parseYAML(`
---
values:
  value: 10
`)
				Expect(template).To(CascadeAs(resolved, source))
			})

			It("supports call chaining", func() {
				source := parseYAML(`
---
mult: (( lambda |x,y|-> x * y ))
values:
  value: (( .mult(2)(5) ))
`)

				resolved := parseYAML(`
---
values:
  value: 10
`)
				Expect(template).To(CascadeAs(resolved, source))
			})

			It("supports chained references", func() {
				source := parseYAML(`
---
func:
  mult: (( lambda |x,y|-> x * y ))
values:
  value: (( (|x|->x)(func).mult(2,5) ))
`)

				resolved := parseYAML(`
---
values:
  value: 10
`)
				Expect(template).To(CascadeAs(resolved, source))
			})
		})

		Context("cross stub", func() {
			It("merges lambda values", func() {
				source := parseYAML(`
---
lvalues: (( merge ))
values: (( lvalues.lvalue(1,2) ))
`)
				stub := parseYAML(`
---
lvalues:
  lvalue: (( lambda |x,y|->x + y ))
`)

				resolved := parseYAML(`
---
values: 3
`)
				Expect(template).To(CascadeAs(resolved, source, stub))
			})
		})
	})

	Describe("using local nodes", func() {
		Context("simple usage", func() {
			It("omits local map nodes", func() {
				source := parseYAML(`
---
temp:
  <<: (( &local ))
  foo: alice
alice: (( temp.foo ))
bob: false
`)

				stub := parseYAML(`
---
temp:
  <<: (( &local ))
  foo: bob
bob: (( temp.foo ))
`)
				resolved := parseYAML(`
---
alice: alice
bob: bob
`)
				Expect(source).To(CascadeAs(resolved, stub))
			})
		})
	})

	Describe("using temporary nodes", func() {
		Context("simple usage", func() {
			It("omits temporary map nodes", func() {
				source := parseYAML(`
---
temp:
  <<: (( &temporary ))
  foo: bar
alice: (( temp.foo ))
`)
				resolved := parseYAML(`
---
alice: bar
`)
				Expect(source).To(CascadeAs(resolved))
			})

			It("omits temporary list entries", func() {
				source := parseYAML(`
---
temp:
  - <<: (( &temporary ))
    foo: bar
  - peter: paul
alice: (( temp.[0].foo ))
`)
				resolved := parseYAML(`
---
temp:
  - peter: paul
alice: bar
`)
				Expect(source).To(CascadeAs(resolved))
			})

			It("propagates temporary map nodes", func() {
				source := parseYAML(`
---
temp:
  #<<: (( &temporary ))
  foo: alice
alice: (( temp.foo ))
bob: false
`)

				stub := parseYAML(`
---
temp:
  <<: (( &temporary ))
  foo: bob
bob: (( temp.foo ))
`)
				resolved := parseYAML(`
---
alice: bob
bob: bob
`)
				Expect(source).To(CascadeAs(resolved, stub))
			})
		})

		Context("combined usage", func() {
			It("omits temporary template nodes", func() {
				source := parseYAML(`
---
temp:
  <<: (( &temporary &template ))
  foo: bar
alice: (( (*temp).foo ))
`)
				resolved := parseYAML(`
---
alice: bar
`)
				Expect(source).To(CascadeAs(resolved))
			})
		})

		Context("with value", func() {
			It("omits temporary list nodes but provides fields", func() {
				source := parseYAML(`
---
temp:
  - <<: (( &temporary ( default ) ))
  - foobar
default:
  - peter

alice: (( temp.[0] ))
`)
				resolved := parseYAML(`
---
default:
  - peter
alice: peter
`)
				Expect(source).To(CascadeAs(resolved))
			})

			It("omits temporary map nodes but provides fields", func() {
				source := parseYAML(`
---
temp:
  <<: (( &temporary ( default ) ))
  foo: bar
default:
  peter: paul

alice: (( temp.peter ))
`)
				resolved := parseYAML(`
---
default:
  peter: paul
alice: paul
`)
				Expect(source).To(CascadeAs(resolved))
			})

			It("omits temporary value nodes but provides value", func() {
				source := parseYAML(`
---
temp:
  peter: paul
  foo: (( &temporary ( peter ) ))

alice: (( temp.foo ))
`)
				resolved := parseYAML(`
---
temp:
  peter: paul
alice: paul
`)
				Expect(source).To(CascadeAs(resolved))
			})
		})

		Context("merging", func() {
			It("overrides", func() {
				source := parseYAML(`
---
temp: (( &temporary ))

alice: (( temp.foo ))
`)

				stub := parseYAML(`
---
temp:
  foo: bar
`)

				resolved := parseYAML(`
---
alice: bar
`)
				Expect(source).To(CascadeAs(resolved, stub))
			})

			It("inherits temporary mode", func() {
				source := parseYAML(`
---
temp: (( merge ))

alice: (( temp.foo ))
`)

				stub := parseYAML(`
---
temp:
  <<: (( &temporary ))
  foo: bar
`)

				resolved := parseYAML(`
---
alice: bar
`)
				Expect(source).To(CascadeAs(resolved, stub))
			})

			It("omits temporary value nodes but provides value", func() {
				source := parseYAML(`
---
temp:
  peter: paul
  foo: (( &temporary ( peter ) ))

alice: (( temp.foo ))
`)
				resolved := parseYAML(`
---
temp:
  peter: paul
alice: paul
`)
				Expect(source).To(CascadeAs(resolved))
			})
		})
	})

	Describe("merging undefined values", func() {
		It("omits merge down of undefined field", func() {
			source := parseYAML(`
---
alice: 24
bob: 25
`)
			stub := parseYAML(`
---
alice: (( config.alice || ~ ))
bob: (( config.bob || ~~ ))
`)
			resolved := parseYAML(`
---
alice: ~
bob: 25
`)
			Expect(source).To(CascadeAs(resolved, stub))
		})

		It("enables merge of values from upstream", func() {
			source := parseYAML(`
---
alice: 24
bob: 25
peter: 26
`)
			stub1 := parseYAML(`
---
config:
  alice: (( ~~ ))
  bob: (( ~~ ))
alice: (( config.alice || ~~ ))
bob: (( config.bob || ~~ ))
peter: (( config.peter || ~~ ))
`)

			stub2 := parseYAML(`
---
config:
  alice: 4711
  peter: 0815
`)

			resolved := parseYAML(`
---
alice: 4711
bob: 25
peter: 26
`)
			Expect(source).To(CascadeAs(resolved, stub1, stub2))
		})
	})
})
