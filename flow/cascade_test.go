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

	It("merges root node", func() {
		source := parseYAML(`
---
<<: (( merge ))
bar: alice
`)

		stub := parseYAML(`
---
foo: bob
`)

		resolved := parseYAML(`
---
foo: bob
bar: alice
`)

		Expect(source).To(CascadeAs(resolved, stub))
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

			It("supports definition scope access", func() {
				source := parseYAML(`
---
env:
  func: (( |x|->[ x, scope, _.scope ] ))
  scope: func

values:
   value: (( env.func("arg") ))
   scope: call
`)

				resolved := parseYAML(`
---
values:
  value:
   - arg
   - call
   - func
  scope: call
`)
				Expect(template).To(CascadeAs(resolved, source))
			})

			It("supports relative names in definition scope access", func() {
				source := parseYAML(`
---
node:
  data:
    scope: data
  funcs:
    a: (( |x|->scope ))
    b: (( |x|->_.scope ))
    c: (( |x|->_.data.scope ))
    scope: funcs

values:
  scope: values

  a: (( node.funcs.a(1) ))
  b: (( node.funcs.b(1) ))
  c: (( node.funcs.c(1) ))
`)

				resolved := parseYAML(`
---
values:
  a: values
  b: funcs
  c: data
  scope: values
`)
				Expect(template).To(CascadeAs(resolved, source))
			})

			It("supports nested calls in definition scope", func() {
				source := parseYAML(`
---
util:
  func: (( |x|-> _.ident(x) ))
  ident: (( |v|->v ))


values: (( util.func(2) ))
`)

				resolved := parseYAML(`
---
values: 2
`)
				Expect(template).To(CascadeAs(resolved, source))
			})

			It("supports nested definitions (map) in definition scope", func() {
				source := parseYAML(`
---
util:
  l1: (( |x|->x <= 0 ? x :map[[x]|dep|->_.l1(dep - 1)].[0] ))

values: (( util.l1(1) ))

`)

				resolved := parseYAML(`
---
values: 0
`)
				Expect(template).To(CascadeAs(resolved, source))
			})

			It("supports nested definitions (lambda) in definition scope", func() {
				source := parseYAML(`
---
util:
  l1: (( |x|->x <= 0 ? x :(|dep|->_.l1(dep - 1))(x) ))

values: (( util.l1(1) ))

`)

				resolved := parseYAML(`
---
values: 0
`)
				Expect(template).To(CascadeAs(resolved, source))
			})

			It("supports definition scope access accross stubs", func() {
				source := parseYAML(`
---
func:

values:
   value: (( (lambda func)("arg") ))
   scope: call
`)

				stub := parseYAML(`
---
env:
  func: (( |x|->[ x, scope, _.scope ] ))
  scope: func

func: (( env.func ))
`)

				resolved := parseYAML(`
---
values:
  value:
   - arg
   - call
   - func
  scope: call
`)
				Expect(template).To(CascadeAs(resolved, source, stub))
			})

			It("supports lambda defaults", func() {
				source := parseYAML(`
---
mult: (( lambda |x,y=5|-> x * y ))
values:
  value: (( .mult(2) ))
`)

				resolved := parseYAML(`
---
values:
  value: 10
`)
				Expect(template).To(CascadeAs(resolved, source))
			})

			It("evaluates lambda defaults in definition scope", func() {
				source := parseYAML(`
---
default: 2
mult: (( lambda |x,y=default * 2|-> x * y ))
`)

				stub := parseYAML(`
---
mult: (( merge ))
values:
  default: 3
  value: (( .mult(3) ))
`)

				resolved := parseYAML(`
---
values:
  default: 3
  value: 12
`)
				Expect(template).To(CascadeAs(resolved, stub, source))
			})

			It("supports deprecated currying", func() {
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

			It("supports currying", func() {
				source := parseYAML(`
---
mult: (( lambda |x,y|-> x * y ))
mult2: (( .mult*(2) ))
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

			It("supports currying with defaults", func() {
				source := parseYAML(`
---
mult: (( lambda |x=1,y=2|-> x * y ))
mult2: (( .mult*(2) ))
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

			It("supports calls for empty currying", func() {
				source := parseYAML(`
---
mult: (( lambda |x,y|-> x * y ))
mult2: (( .mult*() ))
values:
  value: (( .mult2(2,5) ))
`)

				resolved := parseYAML(`
---
values:
  value: 10
`)
				Expect(template).To(CascadeAs(resolved, source))
			})

			It("supports currying with unused varargs", func() {
				source := parseYAML(`
---
func: (( |a,b...|->join(a,b) ))
func1: (( .func*(",")))
values:
  value: (( .func1("a","b") ))
`)

				resolved := parseYAML(`
---
values:
  value: a,b
`)
				Expect(template).To(CascadeAs(resolved, source))
			})

			It("supports currying with used varargs", func() {
				source := parseYAML(`
---
func: (( |a,b...|->join(a,b) ))
func1: (( .func*(",","a","b")))
values:
  value: (( .func1() ))
`)

				resolved := parseYAML(`
---
values:
  value: a,b
`)
				Expect(template).To(CascadeAs(resolved, source))
			})

			It("supports calls for empty currying and defaults", func() {
				source := parseYAML(`
---
mult: (( lambda |x=1,y=2|-> x * y ))
mult2: (( .mult*() ))
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

			It("supports calls for currying with defaults", func() {
				source := parseYAML(`
---
mult: (( lambda |x=1,y=2|-> x * y ))
mult2: (( .mult*(2) ))
values:
  value: (( .mult2() ))
`)

				resolved := parseYAML(`
---
values:
  value: 4
`)
				Expect(template).To(CascadeAs(resolved, source))
			})

			It("supports builtin currying", func() {
				source := parseYAML(`
---
func: (( join*(",") ))
values:
  value: (( .func("a","b") ))
`)

				resolved := parseYAML(`
---
values:
  value: "a,b"
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

		Context("lambda scopes", func() {
			Context("in lambdas", func() {
				It("are passed to nested lambdas", func() {
					source := parseYAML(`
---
func: (( |x|->($a=100) |y|->($b=101) |z|->[x,y,z,a,b] ))

values: (( .func(1)(2)(3) ))

`)
					resolved := parseYAML(`
---
values:
  - 1
  - 2
  - 3
  - 100
  - 101
`)
					Expect(template).To(CascadeAs(resolved, source))
				})
			})
			Context("in templates", func() {
				It("are passed to nested lambdas", func() {
					source := parseYAML(`
---
template:
  <<: (( &template ))
  func: (( |y|->{$x=x, $y=y} ))
func: (( |x|->*template ))
inst: (( .func("instx") ))

values: (( .inst.func("insty") ))
`)
					resolved := parseYAML(`
---
values:
  x: instx
  "y": insty
`)
					Expect(template).To(CascadeAs(resolved, source))
				})
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

	Describe("when asking for expression type", func() {
		It("handles all types", func() {
			source := parseYAML(`
---
temp:
  <<: (( &template &temporary ))

lambda: (( &temporary(|x|->x) ))

map:
  <<: (( &temporary ))
  alice: bob
list:
  - <<: (( &temporary ))
  - alice

types:
   template: (( type(.temp) ))
   lambda: (( type(.lambda) ))
   bool: (( type(true) ))
   int: (( type(1) ))
   string: (( type("s") ))
   map: (( type(.map) ))
   list: (( type(.list) ))
   nil: (( type(~) ))
   undef: (( type(~~) ))
`)
			resolved := parseYAML(`
---
types:
  bool: bool
  int: int
  lambda: lambda
  list: list
  map: map
  nil: nil
  string: string
  template: template
  undef: undef
`)
			Expect(source).To(CascadeAs(resolved))
		})
	})

	Describe("injecting fields", func() {

		Context("for maps", func() {
			It("injects top level field", func() {
				source := parseYAML(`
---
map:
`)
				stub := parseYAML(`
---
injected:
   <<: (( &inject ))
   foo: bar
`)
				resolved := parseYAML(`
---
map:
injected:
  foo: bar
`)
				Expect(source).To(CascadeAs(resolved, stub))
			})

			It("injects selected sub level field", func() {
				source := parseYAML(`
---
map:
  alice: 25
`)
				stub := parseYAML(`
---
map:
   bob: (( &inject(27) ))
   foo: bar
`)
				resolved := parseYAML(`
---
map:
  alice: 25
  bob: 27
`)
				Expect(source).To(CascadeAs(resolved, stub))
			})

			It("injects selected sub level field and overrided others", func() {
				source := parseYAML(`
---
map:
  alice: 25
  tom: 26
`)
				stub := parseYAML(`
---
map:
   bob: (( &inject(27) ))
   foo: bar
   tom: 28
`)
				resolved := parseYAML(`
---
map:
  alice: 25
  bob: 27
  tom: 28
`)
				Expect(source).To(CascadeAs(resolved, stub))
			})

			It("injects selected temporary sub level field", func() {
				source := parseYAML(`
---
map:
  alice: 25
  solution: (( alice + bob ))
`)
				stub := parseYAML(`
---
map:
   bob: (( &inject &temporary (17) ))
   foo: bar
`)
				resolved := parseYAML(`
---
map:
  alice: 25
  solution: 42
`)
				Expect(source).To(CascadeAs(resolved, stub))
			})
		})
		Context("for lists", func() {
			It("injects top level entries", func() {
				source := parseYAML(`
---
- a
- b
`)
				stub := parseYAML(`
---
- (( &inject("c") ))
- d
`)
				resolved := parseYAML(`
---
- c
- a
- b
`)
				Expect(source).To(CascadeAs(resolved, stub))
			})
			It("injects sub level entries", func() {
				source := parseYAML(`
---
list:
- a
- b
`)
				stub := parseYAML(`
---
list:
- (( &inject("c") ))
- d
`)
				resolved := parseYAML(`
---
list:
- c
- a
- b
`)
				Expect(source).To(CascadeAs(resolved, stub))
			})
			It("injects temporary sub level entries", func() {
				source := parseYAML(`
---
list:
- a
- b
`)
				stub := parseYAML(`
---
list:
- (( &inject &temporary ("c") ))
- d
`)
				resolved := parseYAML(`
---
list:
- a
- b
`)
				Expect(source).To(CascadeAs(resolved, stub))
			})
			It("merges lists ony once", func() {
				source := parseYAML(`
---
list:
- <<: (( merge ))
- a
- b
`)
				stub := parseYAML(`
---
list:
- (( &inject ("c") ))
- d
`)
				resolved := parseYAML(`
---
list:
- c
- d
- a
- b
`)
				Expect(source).To(CascadeAs(resolved, stub))
			})

		})

		Context("when there are ignorable dynaml nodes start with '!'", func() {
			It("ignores nodes", func() {
				source := parseYAML(`
---
foo: ((!template_only.foo))
`)

				resolved := parseYAML(`
---
foo: ((template_only.foo))
`)

				Expect(source).To(CascadeAs(resolved))
			})
			It("ignores merge nodes", func() {
				source := parseYAML(`
---
foo:
  <<!: test
`)

				resolved := parseYAML(`
---
foo:
  <<: test
`)

				Expect(source).To(CascadeAs(resolved))
			})

			It("ignores nodes with escape", func() {
				source := parseYAML(`
---
foo: ((!template_only.foo))
`)

				resolved := parseYAML(`
---
foo: ((template_only.foo))
`)

				Expect(source).To(CascadeAs(resolved))
			})
			It("ignores nodes with escaped escape", func() {
				source := parseYAML(`
---
foo: ((!!template_only.foo))
`)

				resolved := parseYAML(`
---
foo: ((!template_only.foo))
`)

				Expect(source).To(CascadeAs(resolved))
			})
			It("ignores nodes with escaped interpolation", func() {
				source := parseYAML(`
---
foo: x ((!template_only.foo))
`)

				resolved := parseYAML(`
---
foo: x ((template_only.foo))
`)

				Expect(source).To(CascadeAs(resolved))
			})
			It("ignores merge nodes with escaped escape", func() {
				source := parseYAML(`
---
foo:
  <<!!: test
`)

				resolved := parseYAML(`
---
foo:
  <<!: test
`)

				Expect(source).To(CascadeAs(resolved))
			})
		})
	})
})
