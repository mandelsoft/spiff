```
                                        ___ _ __ (_)/ _|/ _|
                                       / __| '_ \| | |_| |_
                                       \__ \ |_) | |  _|  _|
                                       |___/ .__/|_|_| |_|
                                           |_|

```

---

**NOTE**: *Active development on spiff is currently paused, including Pull Requests.  Very severe issues will be addressed, and we will still be actively responding to requests for help via Issues.*

---

*spiff* is a command line tool and declarative in-domain hybrid YAML templating system. While regular templating systems process a template file by substituting the template expressions by values taken from
external data sources, in-domain means that the templating engine knows about the syntax and structure of the processed template. It therefore can take the values for the template expressions directly
from the document processed, including those parts denoted by the template expressions itself.

For example:
```yaml
resource:
  name: bosh deployment
  version: 25
  url: (( "http://resource.location/bosh?version=" version ))
  description: (( "This document describes a " name " located at " url ))
```

spiff is a command line tool and declarative YAML templating system, specially designed for generating BOSH deployment manifests.

Contents:
- [Installation](#installation)
- [Usage](#usage)
- [dynaml Templating Language](#dynaml-templating-language)
	- [(( foo ))](#-foo-)
	- [(( foo.bar.[1].baz ))](#-foobar1baz-)
	- [(( foo.[bar].baz ))](#-foobarbaz-)
	- [(( list.[1..3] ))](#-list13-)
	- [(( "foo" ))](#-foo--1)
	- [(( [ 1, 2, 3 ] ))](#--1-2-3--)
	- [(( { "alice" = 25 } ))](#--alice--25--)
	- [(( foo bar ))](#-foo-bar-)
		- [(( "foo" bar ))](#-foo-bar--1)
		- [(( [1,2] bar ))](#-12-bar-)
		- [(( map1 map2 ))](#-map1-map2-)
	- [(( auto ))](#-auto-)
	- [(( merge ))](#-merge-)
		- [<<: (( merge ))](#--merge-)
			- [merging maps](#merging-maps)
			- [merging lists](#merging-lists)
		- [- <<: (( merge on key ))](#----merge-on-key-)
		- [<<: (( merge replace ))](#--merge-replace-)
			- [merging maps](#merging-maps-1)
			- [merging lists](#merging-lists-1)
		- [<<: (( foo )) ](#--foo-)
			- [merging maps](#merging-maps-2)
			- [merging lists](#merging-lists-2)
		- [<<: (( merge foo ))](#--merge-foo-)
			- [merging maps](#merging-maps-3)
			- [merging lists](#merging-lists-3)
	- [(( a || b ))](#-a--b-)
	- [(( 1 + 2 * foo ))](#-1--2--foo-)
	- [(( "10.10.10.10" - 11 ))](#-10101010---11-)
	- [(( a > 1 ? foo :bar ))](#-a--1--foo-bar-)
	- [(( 5 -or 6 ))](#-5--or-6-)
	- [Functions](#functions)
		- [(( format( "%s %d", alice, 25) ))](#-format-s-d-alice-25-)
		- [(( join( ", ", list) ))](#-join---list-)
		- [(( split( ",", string) ))](#-split--string-)
		- [(( trim(string) ))](#-trimstring-)
		- [(( element(list, index) ))](#-elementlist-index-)
		- [(( element(map, key) ))](#-elementmap-key-)
		- [(( compact(list) ))](#-compactlist-)
		- [(( uniq(list) ))](#-uniqlist-)
		- [(( contains(list, "foobar") ))](#-containslist-foobar-)
		- [(( index(list, "foobar") ))](#-indexlist-foobar-)
		- [(( lastindex(list, "foobar") ))](#-lastindexlist-foobar-)
		- [(( replace(string, "foo", "bar") ))](#-replacestring-foo-bar-)
		- [(( substr(string, 1, 3) ))](#-substrstring-1-3-)
		- [(( match("(f.*)(b.*)", "xxxfoobar") ))](#-matchfb-xxxfoobar-)
		- [(( length(list) ))](#-lengthlist-)
		- [(( base64(string) ))](#-base64string-)
		- [(( md5(string) ))](#-md5string-)
		- [(( defined(foobar) ))](#-definedfoobar-)
		- [(( valid(foobar) ))](#-validfoobar-)
		- [(( require(foobar) ))](#-requirefoobar-)
		- [(( stub(foo.bar) ))](#-stubfoobar-)
		- [(( exec( "command", arg1, arg2) ))](#-exec-command-arg1-arg2-)
		- [(( eval( foo "." bar ) ))](#-eval-foo--bar--)
		- [(( env( "HOME" ) ))](#-env-HOME--)
		- [(( read("file.yml") ))](#-readfileyml-)
		- [(( static_ips(0, 1, 3) ))](#-static_ips0-1-3-)
		- [(( ipset(ranges, 3, 3,4,5,6) ))](#-ipsetranges-3-3456-)
		- [(( list_to_map(list, "key") ))](#-list_to_maplist-key-)
		- [(( makemap(fieldlist) ))](#-makemapfieldlist-)
		- [(( makemap(key, value) ))](#-makemapkey-value-)
		- [(( merge(map1, map2) ))](#-mergemap1-map2-)
	- [(( lambda |x|->x ":" port ))](#-lambda-x-x--port-)
	- [(( &temporary ))](#-temporary-)
	- [Mappings](#mappings)
		- [(( map[list|elem|->dynaml-expr] ))](#-maplistelem-dynaml-expr-)
		- [(( map[list|idx,elem|->dynaml-expr] ))](#-maplistidxelem-dynaml-expr-)
		- [(( map[map|key,value|->dynaml-expr] ))](#-mapmapkeyvalue-dynaml-expr-)
	- [Aggregations](#aggregations)
		- [(( sum[list|initial|sum,elem|->dynaml-expr] ))](#-sumlistinitialsumelem-dynaml-expr-)
		- [(( sum[list|initial|sum,idx,elem|->dynaml-expr] ))](#-sumlistinitialsumidxelem-dynaml-expr-)
		- [(( sum[map|initial|sum,key,value|->dynaml-expr] ))](#-summapinitialsumkeyvalue-dynaml-expr-)
	- [Projections](#projections)
	    - [(( expr.[*].value ))](#-exprvalue-)
		- [(( list.[1..2].value ))](#-list12value-)
	- [Templates](#templates)
		- [<<: (( &template ))](#--template-)
		- [(( *foo.bar ))](#-foobar-)
	- [Special Literals](#special-literals)
	- [Access to evaluation context](#access-to-evaluation-context)
	- [Operation Priorities](#operation-priorities)
- [Structural Auto-Merge](#structural-auto-merge)
- [Bringing it all together](#bringing-it-all-together)
- [Useful to Know](#useful-to-know)
- [Error Reporting](#error-reporting)


# Installation

Official release executable binaries can be downloaded via [Github releases](https://github.com/cloudfoundry-incubator/spiff/releases) for Darwin and Linux machines (and virtual machines).

Some of spiff's dependencies have changed since the last official release, and spiff will not be updated to keep up with these dependencies.  Working dependencies are vendored in the `Godeps` directory (more information on the `godep` tool is available [here](https://github.com/tools/godep)).  As such, trying to `go get` spiff will likely fail; the only supported way to use spiff is to use an official binary release.

# Usage

### `spiff merge template.yml [template2.yml ...]`

Merge a bunch of template files into one manifest, printing it out.

See 'dynaml templating language' for details of the template file, or examples/ subdir for more complicated examples.

Example:

```
spiff merge cf-release/templates/cf-deployment.yml my-cloud-stub.yml
```

The ` merge` command offers the option `--partial`. If this option is
given spiff handles incomplete expression evaluation. All errors are ignored
and the unresolvable parts of the yaml document are returned as strings.

It is possible to read one file from standard input by using the file
name `-`. It may be used only once. This allows using spiff as part of a
pipeline to just process a single stream or to process a stream based on
several templates/stubs.

The template file (first argument) may be a multiple document stream
containing multiple YAML documents separated by a line containing only `---`.
Each YAML document will be processed independently with the given stub files.
The result is the stream of processed documents in the same order.
For example, this can be used to generate *kubernetes* manifests to be used
by `kubectl`.

### `spiff diff manifest.yml other-manifest.yml`

Show structural differences between two deployment manifests.
Here streams with multiple documents are supported, also.
To indicate no difference the number of documents in both streams must be
identical and each document in the first stream must have no difference
compared to the document with the same index in the second stream.
Found differences are shown for each document separately.

Unlike basic diffing tools and even `bosh diff`, this command has semantic
knowledge of a deployment manifest, and is not just text-based. For example,
if two manifests are the same except they have some jobs listed in different
orders, `spiff diff` will detect this, since job order matters in a manifest.
On the other hand, if two manifests differ only in the order of their
resource pools, for instance, then it will yield and empty diff since
resource pool order doesn't actually matter for a deployment.

Also unlike `bosh diff`, this command doesn't modify either file.

It's tailed for checking differences between one deployment and the next.

Typical flow:

```sh
$ spiff merge template.yml [templates...] > deployment.yml
$ bosh download manifest [deployment] current.yml
$ spiff diff deployment.yml current.yml
$ bosh deployment deployment.yml
$ bosh deploy
```


# dynaml Templating Language

Spiff uses a declarative, logic-free templating language called 'dynaml'
(dynamic yaml).

Every dynaml node is guaranteed to resolve to a YAML node. It is *not*
string interpolation. This keeps developers from having to think about how
a value will render in the resulting template.

A dynaml node appears in the .yml file as an expression surrounded by two
parentheses. They can be used as the value of a map or an entry in a list.

The following is a complete list of dynaml expressions:


## `(( foo ))`

Look for the nearest 'foo' key (i.e. lexical scoping) in the current
template and bring it in.

e.g.:

```yaml
fizz:
  buzz:
    foo: 1
    bar: (( foo ))
  bar: (( foo ))
foo: 3
bar: (( foo ))
```

This example will resolve to:

```yaml
fizz:
  buzz:
    foo: 1
    bar: 1
  bar: 3
foo: 3
bar: 3
```

The following will not resolve because the key name is the same as the value to be merged in:
```yaml
foo: 1

hi:
  foo: (( foo ))
```

## `(( foo.bar.[1].baz ))`

Look for the nearest 'foo' key, and from there follow through to .bar.baz.

A path is a sequence of steps separated by dots. A step is either a word for
maps, or digits surrounded by brackets for list indexing.

If the path cannot be resolved, this evaluates to nil. A reference node at the
top level cannot evaluate to nil; the template will be considered not fully
resolved. If a reference is expected to sometimes not be provided, it should be
used in combination with '||' (see below) to guarantee resolution.

Note that references are always within the template, and order does not matter.
You can refer to another dynamic node and presume it's resolved, and the
reference node will just eventually resolve once the dependent node resolves.

e.g.:

```yaml
properties:
  foo: (( something.from.the.stub ))
  something: (( merge ))
```

This will resolve as long as 'something' is resolveable, and as long as it
brings in something like this:

```yaml
from:
  the:
    stub: foo
```

If the path starts with a dot (`.`) the path is always evaluated from the root
of the document.

List entries consisting of a map with `name` field can directly be addressed
by their name value.

e.g.:

The age of alice in

```yaml
list:
 - name: alice
   age: 25
```

can be referenced by using the path `list.alice.age`, instead of `list[0].age`.

## `(( foo.[bar].baz ))`

Look for the nearest 'foo' key, and from there follow through to the
field(s) described by the expression `bar` and then to .baz.

The index may be an integer constant (without spaces) as described in the
last section. But it might also be an arbitrary dynaml expression (even
an integer, but with spaces). If the expression evaluates to a string,
it lookups the dedicated field. If the expression evaluates to an integer,
the array element with this index is addressed.

e.g.:

```yaml
properties:
  name: alice
  foo: (( values.[name].bar ))
  values:
    alice:
	   bar: 42
```

This will resolve `foo` to the value `42`. The dynamic index may also be at
the end of the expression (without `.bar`).

Basically this is the simplier way to express something like
[eval("values." name ".bar")](#-eval-foo--bar--)

If the expression evaluates to a list, the list elements (strings or integers)
are used as path elements to access deeper fields.

e.g.:

```yaml
properties:
  name:
   - foo
   - bar
  foo: (( values.[name] ))
  values:
    foo:
	   bar: 42
```

resolves `foo` again to the value `42`.

## `(( list.[1..3] ))`

The slice expression can be used to extract a dedicated sub list from a list
expression. The range *start* `..` *end* extracts a list of the length
*end-start+1* with the elements from
index *start* to *end*. If the start index is negative the slice is taken
from the end of the list from *length-start* to *length-end*. If the end
index is lower than the start index, the result is an empty array.

e.g.:

```yaml
list:
  - a
  - b
  - c
foo: (( list.[1..length(list) - 1] ))
```

evaluates `foo` to the list `[b,c]`.

## `(( "foo" ))`

String literal. The only escape character handled currently is '"'.

## `(( [ 1, 2, 3 ] ))`

List literal. The list elements might again be expressions. There is a special list literal `[1 .. -1]`, that can be used to resolve an increasing or descreasing number range to a list.

e.g.:

```yaml
list: (( [ 1 .. -1 ] ))
```

yields

```yaml
list:
  - 1
  - 0
  - -1
```

## `(( { "alice" = 25 } ))`

The map literal can be used to describe maps as part of a dynaml expression. Both,
the key and the value, might again be expressions, whereby the key expression must
evaluate to a string. This way it is possible to create maps with non-static keys.
The assignment operator `=` has been chosen instead of the regular colon `:`
character used in yaml, because this would result in conflicts with the yaml
syntax.

A map literal might consist of any number of field assignments separated by a
comma `,`.

e.g.:

```yaml
name: peter
age: 23
map: (( { "alice" = {}, name = age } ))
```

yields

```yaml
name: peter
age: 23
map:
  alice: {}
  peter: 23
```

Another way to compose lists based on expressions are the functions
[`makemap`](#-makemapkey-value-) and [`list_to_map`](#-list_to_maplist-key-).


## `(( foo bar ))`

Concatenation expression used to concatenate a sequence of dynaml expressions.

### `(( "foo" bar ))`

Concatenation (where bar is another dynaml expr). Any sequences of simple values (string, integer and boolean) can be concatenated, given by any dynaml expression.

e.g.:

```yaml
domain: example.com
uri: (( "https://" domain ))
```

In this example `uri` will resolve to the value `"https://example.com"`.

### `(( [1,2] bar ))`

Concatenation of lists as expression (where bar is another dynaml expr). Any sequences of lists can be concatenated, given by any dynaml expression.

e.g.:

```yaml
other_ips: [ 10.0.0.2, 10.0.0.3 ]
static_ips: (( ["10.0.1.2","10.0.1.3"] other_ips ))
```

In this example `static_ips` will resolve to the value `[ 10.0.1.2, 10.0.1.3, 10.0.0.2, 10.0.0.3 ] `.

If the second expression evaluates to a value other than a list (integer, boolean, string or map), the value is appended to the first list.

e.g.:

```yaml
foo: 3
bar: (( [1] 2 foo "alice" ))
```
yields the list `[ 1, 2, 3, "alice" ]` for `bar`.

### `(( map1 map2 ))`

Concatenation of maps as expression. Any sequences of maps can be concatenated, given by any dynaml expression. Thereby entries will be merged. Entries with the same key are overwritten from left to right.

e.g.:

```yaml
foo:
  alice: 24
  bob: 25

bar:
  bob: 26
  paul: 27

concat: (( foo bar ))
```

yields

```yaml
foo:
  alice: 24
  bob: 25

bar:
  bob: 26
  paul: 27

concat:
  alice: 24
  bob: 26
  paul: 27
```

## `(( auto ))`

Context-sensitive automatic value calculation.

In a resource pool's 'size' attribute, this means calculate based on the total
instances of all jobs that declare themselves to be in the current resource
pool.

e.g.:

```yaml
resource_pools:
  - name: mypool
    size: (( auto ))

jobs:
  - name: myjob
    resource_pool: mypool
    instances: 2
  - name: myotherjob
    resource_pool: mypool
    instances: 3
  - name: yetanotherjob
    resource_pool: otherpool
    instances: 3
```

In this case the resource pool size will resolve to '5'.

## `(( merge ))`

Bring the current path in from the stub files that are being merged in.

e.g.:

```yaml
foo:
  bar:
    baz: (( merge ))
```

Will try to bring in `foo.bar.baz` from the first stub, or the second, etc.,
returning the value from the first stub that provides it.

If the corresponding value is not defined, it will return nil. This then has the
same semantics as reference expressions; a nil merge is an unresolved template.
See `||`.

### `<<: (( merge ))`

Merging of maps or lists with the content of the same element found in some stub.

** Attention **
This form of `merge` has a compatibility propblem. In versions before 1.0.8, this expression
was never parsed, only the existence of the key `<<:` was relevant. Therefore there are often
usages of `<<: (( merge ))` where `<<: (( merge || nil ))` is meant. The first variant would
require content in at least one stub (as always for the merge operator). Now this expression
is evaluated correctly, but this would break existing manifest template sets, which use the
first variant, but mean the second. Therfore this case is explicitly handled to describe an
optional merge. If really a required merge is meant an additional explicit qualifier has to
be used (`(( merge required ))`).

#### Merging maps

**values.yml**
```yaml
foo:
  a: 1
  b: 2
```

**template.yml**
```yaml
foo:
  <<: (( merge ))
  b: 3
  c: 4
```

`spiff merge template.yml values.yml` yields:

```yaml
foo:
  a: 1
  b: 2
  c: 4
```

#### Merging lists

**values.yml**
```yaml
foo:
  - 1
  - 2
```

**template.yml**
```yaml
foo:
  - 3
  - <<: (( merge ))
  - 4
```

`spiff merge template.yml values.yml` yields:

```yaml
foo:
  - 3
  - 1
  - 2
  - 4
```

### `- <<: (( merge on key ))`

`spiff` is able to merge lists of maps with a key field. Those lists are handled like maps with the value of the key field as key. By default the key `name` is used. But with the selector `on` an arbitrary key name can be specified for a list-merge expression.

e.g.:

```yaml
list:
  - <<: (( merge on key ))
  - key: alice
    age: 25
  - key: bob
    age: 24
```

merged with

```yaml
list:
  - key: alice
    age: 20
  - key: peter
    age: 13
```

yields

```yaml
list:
  - key: peter
    age: 13
  - key: alice
    age: 20
  - key: bob
    age: 24
```

If no insertion of new entries is desired (as requested by the insertion merge expression), but only overriding of existent entries, one existing key field can be prefixed with the tag `key:` to indicate a non-standard key name, for example `- key:key: alice`.

### `<<: (( merge replace ))`

Replaces the complete content of an element by the content found in some stub instead of doing a deep merge for the existing content.

#### Merging maps

**values.yml**
```yaml
foo:
  a: 1
  b: 2
```

**template.yml**
```yaml
foo:
  <<: (( merge replace ))
  b: 3
  c: 4
```

`spiff merge template.yml values.yml` yields:

```yaml
foo:
  a: 1
  b: 2
```

#### Merging lists

**values.yml**
```yaml
foo:
  - 1
  - 2
```

**template.yml**
```yaml
foo:
  - <<: (( merge replace ))
  - 3
  - 4
```

`spiff merge template.yml values.yml` yields:

```yaml
foo:
  - 1
  - 2
```

### `<<: (( foo ))`

Merging of maps and lists found in the same template or stub.

#### Merging maps

```yaml
foo:
  a: 1
  b: 2

bar:
  <<: (( foo )) # any dynaml expression
  b: 3
```

yields:

```yaml
foo:
  a: 1
  b: 2

bar:
  a: 1
  b: 3
```

This expression just adds new entries to the actual list. It does not merge
existing entries with the content described by the merge expression.

#### Merging lists

```yaml
bar:
  - 1
  - 2

foo:
  - 3
  - <<: (( bar ))
  - 4
```

yields:

```yaml
bar:
  - 1
  - 2

foo:
  - 3
  - 1
  - 2
  - 4
```

A common use-case for this is merging lists of static ips or ranges into a list of ips. Another possibility is to use a single [concatenation expression](#-12-bar-).

### `<<: (( merge foo ))`

Merging of maps or lists with the content of an arbitrary element found in some stub (Redirecting merge). There will be no further (deep) merge with the element of the same name found in some stub. (Deep merge of lists requires maps with field `name`)

Redirecting merges can be used as direct field value, also. They can be combined with replacing merges like `(( merge replace foo ))`.

#### Merging maps

**values.yml**
```yaml
foo:
  a: 10
  b: 20

bar:
  a: 1
  b: 2
```

**template.yml**
```yaml
foo:
  <<: (( merge bar))
  b: 3
  c: 4
```

`spiff merge template.yml values.yml` yields:

```yaml
foo:
  a: 1
  b: 2
  c: 4
```

Another way doing a merge with another element in some stub could also be done the traditional way:

**values.yml**
```yaml
foo:
  a: 10
  b: 20

bar:
  a: 1
  b: 2
```

**template.yml**
```yaml
bar:
  <<: (( merge ))
  b: 3
  c: 4

foo: (( bar ))
```

But in this scenario the merge still performs the deep merge with the original element name. Therefore
`spiff merge template.yml values.yml` yields:

```yaml
bar:
  a: 1
  b: 2
  c: 4
foo:
  a: 10
  b: 20
  c: 4
```

#### Merging lists

**values.yml**
```yaml
foo:
  - 10
  - 20

bar:
  - 1
  - 2
```

**template.yml**
```yaml
foo:
  - 3
  - <<: (( merge bar ))
  - 4
```

`spiff merge template.yml values.yml` yields:

```yaml
foo:
  - 3
  - 1
  - 2
  - 4
```

## `(( a || b ))`

Uses a, or b if a cannot be resolved.

e.g.:

```yaml
foo:
  bar:
    - name: some
    - name: complicated
    - name: structure

mything:
  complicated_structure: (( merge || foo.bar ))
```

This will try to merge in `mything.complicated_structure`, or, if it cannot be
merged in, use the default specified in `foo.bar`.

## `(( 1 + 2 * foo ))`

Dynaml expressions can be used to execute arithmetic integer calculations. Supported operations are +, -, *, / and %.

e.g.:

**values.yml**
```yaml
foo: 3
bar: (( 1 + 2 * foo ))
```

`spiff merge values.yml` yields `7` for `bar`. This can be combined with [concatentions](#-foo-bar-) (calculation has higher priority than concatenation in dynaml expressions):

```yaml
foo: 3
bar: (( foo " times 2 yields " 2 * foo ))
```
The result is the string `3 times 2 yields 6`.

## `(( "10.10.10.10" - 11 ))`

Besides arithmetic on integers it is also possible to use addition and subtraction on ip addresses.

e.g.:

```yaml
ip: 10.10.10.10
range: (( ip "-" ip + 247 + 256 * 256 ))
```

yields

```yaml
ip: 10.10.10.10
range: 10.10.10.10-10.11.11.1
```

Subtraction also works on two IP addresses to calculate the number of
IP addresses between two IP addresses.

e.g.:

```yaml
diff: (( 10.0.1.0 - 10.0.0.1 + 1 ))
```

yields the value 256. IP address constants can be directly used in dynaml
expressions. They are implicitly converted to strings and back to IP
addresses if required by an operation.

Multiplication and division can be used to handle IP range shifts on CIDRs.
With division a network can be partioned. The network size is increased
to allow at least a dedicated number of subnets below the original CIDR.
Multiplication then can be used to get the n-th next subnet of the same
size.

e.g.:

```yaml
subnet: (( "10.1.2.1/24" / 12 ))  # first subnet CIDR for 16 subnets
next: (( "10.1.2.1/24" / 12 * 2)) # 2nd next (3rd) subnet CIDRS
```

yields

```yaml
subnet: 10.1.2.0/28
next: 10.1.2.32/28
```

Additionally there are functions working on IPv4 CIDRs:

```yaml
cidr: 192.168.0.1/24
range: (( min_ip(cidr) "-" max_ip(cidr) ))
next: (( max_ip(cidr) + 1 ))
num: (( min_ip(cidr) "+" num_ip(cidr) "=" min_ip(cidr) + num_ip(cidr) ))
```

yields

```yaml
cidr: 192.168.0.1/24
range: 192.168.0.0-192.168.0.255
next: 192.168.1.0
num: 192.168.0.0+256=192.168.1.0
```

## `(( a > 1 ? foo :bar ))`

Dynaml supports the comparison operators `<`, `<=`, `==`, `!=`, `>=` and `>`. The comparison operators work on
integer values. The checks for equality also work on lists and maps. The result is always a boolean value. To negate a condition the unary not opertor (`!`) can be used.

Additionally there is the ternary conditional operator `?:`, that can be used to evaluate expressions depending on a condition. The first operand is used as condition. The expression is evaluated to the second operand, if the condition is true, and to the third one, otherwise.

e.g.:

```yaml
foo: alice
bar: bob
age: 24
name: (( age > 24 ? foo :bar ))
```

yields the value `bob` for the property `name`.

**Remark**

The use of the symbol `:` may collide with the yaml syntax, if the complete expression is not a quoted string value.

The operators `-or` and `-and` can be used to combine comparison operators to compose more complex conditions.

**Remark:**

The more traditional operator symbol `||` (and `&&`) cannot be used here, because the operator `||` already exists in dynaml with a different semantic, that does not hold for logical operations. The expression `false || true` evaluates to `false`, because it yields the first operand, if it is defined, regardless of its value. To be as compatible as possible this cannot be changed and the bare symbols `or` and `and` cannot be be used, because this would invalidate the concatenation of references with such names.

## `(( 5 -or 6 ))`

If both sides of an `-or` or `-and` operator evaluate to integer values, a bit-wise operation is executed and the result is again an integer. Therefore the expression `5 -or 6` evaluates to `7`.

## Functions

Dynaml supports a set of predefined functions. A function is generally called like

```yaml
result: (( functionname(arg, arg, ...) ))
```

Additional functions may be defined as part of the yaml document using [lambda expressions](#-lambda-x-x--port-). The function name then is either a grouped expression or the path to the node hosting the lambda expression.

### `(( format( "%s %d", alice, 25) ))`

Format a string based on arguments given by dynaml expressions. There is a second flavor of this function: `error` formats an error message and sets the evaluation to failed.


### `(( join( ", ", list) ))`

Join entries of lists or direct values to a single string value using a given separator string. The arguments to join can be dynaml expressions evaluating to lists, whose values again are strings or integers, or string or integer values.

e.g.:

```yaml
alice: alice
list:
  - foo
  - bar

join: (( join(", ", "bob", list, alice, 10) ))
```

yields the string value `bob, foo, bar, alice, 10` for `join`.

### `(( split( ",", string) ))`

Split a string for a dedicated separator. The result is a list.

e.g.:

```yaml
list: (( split("," "alice, bob") ))
```

yields:

```yaml
list:
  - alice
  - ' bob'
```

### `(( trim(string) ))`

Trim a string or all elements of a list of strings. There is an optional second string argument. It can be used to specify a set of characters that will be cut. The default cut set consists of a space and a tab character.

e.g.:

```yaml
list: (( trim(split("," "alice, bob")) ))
```

yields:

```yaml
list:
  - alice
  - bob
```

### `(( element(list, index) ))`

Return a dedicated list element given by its index.

e.g.:

```yaml
list: (( trim(split("," "alice, bob")) ))
elem: (( element(list,1) ))
```

yields:

```yaml
list:
  - alice
  - bob
elem: bob
```

### `(( element(map, key) ))`

Return a dedicated map field given by its key.

```yaml
map:
  alice: 24
  bob: 25
elem: (( element(map,"bob") ))
```

yields:

```yaml
map:
  alice: 24
  bob: 25
elem: 25
```

This function is also able to handle keys containing dots (.).

### `(( compact(list) ))`

Filter a list omitting empty entries.

e.g.:

```yaml
list: (( compact(trim(split("," "alice, , bob"))) ))
```

yields:

```yaml
list:
  - alice
  - bob
```

### `(( uniq(list) ))`

Uniq provides a list without dupliates.

e.g.:

```yaml
list:
- a
- b
- a
- c
- a
- b
- 0
- "0"
uniq: (( uniq(list) ))
```

yields for field `uniq`:

```yaml
uniq:
- a
- b
- c
- 0
```

### `(( contains(list, "foobar") ))`

Checks whether a list contains a dedicated value. Values might also be lists or maps.

e.g.:

```yaml
list:
  - foo
  - bar
  - foobar
contains: (( contains(list, "foobar") ))
```

yields:

```yaml
list:
  - foo
  - bar
  - foobar
contains: true
```

The function `contains` also works on strings to look for sub strings.

e.g.:

```yaml
contains: (( contains("foobar", "bar") ))
```

yields `true`.

### `(( index(list, "foobar") ))`

Checks whether a list contains a dedicated value and returns the index of the first match.
Values might also be lists or maps. If no entry could be found `-1` is returned.

e.g.:

```yaml
list:
  - foo
  - bar
  - foobar
index: (( index(list, "foobar") ))
```

yields:

```yaml
list:
  - foo
  - bar
  - foobar
index: 2
```

The function `index` also works on strings to look for sub strings.

e.g.:

```yaml
index: (( index("foobar", "bar") ))
```

yields `3`.

### `(( lastindex(list, "foobar") ))`

The function `lastindex` works like [`index`](#-indexlist-foobar-) but the index of the last occurence is returned.

### `(( replace(string, "foo", "bar") ))`

Replace all occurences of a sub string in a string by a replacement string. With an optional
fourth integer argument the number of substitutions can be limited (-1 mean unlimited).

e.g.:

```yaml
string: (( replace("foobar", "o", "u") ))
```

yields `fuubar`.

### `(( substr(string, 1, 2) ))`

Extract a stub string from a string, starting from a given start index up to an optional end index (exclusive). If no end index is given the sub struvt up to the end of the string is extracted.
Both indices might be negative. In this case they are taken from the end of the string.

e.g.:

```yaml
string: "foobar"
end1: (( substr(string,-2) ))
end2: (( substr(string,3) ))
range: (( substr(string,1,-1) ))
```

evaluates to

```yaml
string: foobar
end1: ar
end2: bar
range: ooba
```

### `(( match("(f.*)(b.*)", "xxxfoobar") ))`

Returns the match of a regular expression for a given string value. The match is a list of the matched values for the sub expressions contained in the regular expression. Index 0 refers to the match of the complete regular expression. If the string value does not match an empty list is returned.

e.g.:

```yaml
matches: (( match("(f.*)*(b.*)", "xxxfoobar") ))
```

yields:

```yaml
matches:
- foobar
- foo
- bar
```

### `(( length(list) ))`

Determine the length of a list, a map or a string value.

e.g.:

```yaml
list:
  - alice
  - bob
length: (( length(list) ))
```

yields:

```yaml
list:
  - alice
  - bob
length: 2
```

### `(( base64(string) ))`

The function `base64` generates a base64 encoding of a given string. `base64_decode` decodes a base64 encoded string.

e.g.:

```yaml
base64: (( base64("test") ))
test: (( base64_decode(base64)))
```

evaluates to

```yaml
base54: dGVzdA==
test: test
```

### `(( md5(string) ))`

The function `md5` generates an md5 hash for the given string.

e.g.:

```yaml
hash: (( md5("test") ))
```

evaluates to

```yaml
hash: 098f6bcd4621d373cade4e832627b4f6
```

### `(( defined(foobar) ))`

The function `defined` checks whether an expression can successfully be evaluated. It yields the boolean value `true`, if the expression can be evaluated, and `false` otherwise.

e.g.:

```yaml
zero: 0
div_ok: (( defined(1 / zero ) ))
zero_def: (( defined( zero ) ))
null_def: (( defined( null ) ))
```

evaluates to

```yaml
zero: 0
div_ok: false
zero_def: true
null_def: false
```

This function can be used in combination of the [conditional operator](#-a--1--foo-bar-) to evaluate expressions depending on the resolvability of another expression.

### `(( valid(foobar) ))`

The function `valid` checks whether an expression can successfully be evaluated and evaluates to a defined value not equals to `nil`. It yields the boolean value `true`, if the expression can be evaluated, and `false` otherwise.

e.g.:

```yaml
zero: 0
empty:
map: {}
list: []
div_ok: (( valid(1 / zero ) ))
zero_def: (( valid( zero ) ))
null_def: (( valid( ~ ) ))
empty_def: (( valid( empty ) ))
map_def: (( valid( map ) ))
list_def: (( valid( list ) ))
```

evaluates to

```yaml
zero: 0
empty: null
map: {}
list: []
div_ok:   false
zero_def: true
null_def: false
empty_def: false
map_def:  true
list_def: true
```

### `(( require(foobar) ))`

The function `require` yields an error if the given argument is undefined or `nil`, otherwise it yields the given value.

e.g.:

```yaml
foo: ~
bob: (( foo || "default" ))
alice: (( require(foo) || "default" ))
```

evaluates to

```yaml
foo: ~
bob: ~
alice: default
```

### `(( stub(foo.bar) ))`

The function `stub` yields the value of a dedicated field found in the first upstream stub defining it.

e.g.:

**template.yml**
```yaml
value: (( stub(foo.bar) ))
```
merged with stub

**stub.yml**
```yaml
foo:
  bar: foobar
```

evaluates to

```yaml
value: foobar
```

The argument passed to this function must either be a reference literal or an expression evaluating to a string denoting a reference. If no argument is given, the actual field path is used.

Alternatively the `merge` operation could be used, for example `merge foo.bar`. The difference is that `stub` does not merge, therefore the field will still be merged (with the original path in the document).


### `(( exec( "command", arg1, arg2) ))`

Execute a command. Arguments can be any dynaml expressions including reference expressions evaluated to lists or maps. Lists or maps are passed as single arguments containing a yaml document with the given fragment.

The result is determined by parsing the standard output of the command. It might be a yaml document or a single multi-line string or integer value. A yaml document must start with the document prefix `---`. If the command fails the expression is handled as undefined.

e.g.

```yaml
arg:
  - a
  - b
list: (( exec( "echo", arg ) ))
string: (( exec( "echo", arg.[0] ) ))

```

yields

```yaml
arg:
- a
- b
list:
- a
- b
string: a
```

Alternatively `exec` can be called with a single list argument completely describing the command line.

The same command will be executed once, only, even if it is used in multiple expressions.

### `(( eval( foo "." bar ) ))`

Evaluate the evaluation result of a string expression again as dynaml expression. This can, for example, be used to realize indirections.

e.g.: the expression in

```yaml
alice:
  bob: married

foo: alice
bar: bob

status: (( eval( foo "." bar ) ))
```

calculates the path to a field, which is then evaluated again to yield the value of this composed field:

```yaml
alice:
  bob: married

foo: alice
bar: bob

status: married
```

### `(( env( "HOME" ) ))`

Read the value of an environment variable whose name is given as dynaml expression. If the environment variable is not set the evaluation fails.

In a second flavor the function `env` accepts multiple arguments and/or list arguments, which are joined to a single list. Every entry in this list is used as name of an environment variable and the result of the function is a map of the given given variables as yaml element. Hereby non-existent environment variables are omitted.

### `(( read("file.yml") ))`

Read a file and return its content. There is support for two content types: `yaml` files and `text` files.
If the file suffix is `.yml`, by default the yaml type is used. An optional second parameter can be used
to explicitly specifiy the desired return type: `yaml` or `text`.

#### yaml documents

A yaml document will be parsed and the tree is returned. The  elements of the tree can be accessed by regular dynaml expressions.

Additionally the yaml file may again contain dynaml expressions. All included dynaml expressions will be evaluated in the context of the reading expression. This means that the same file included at different places in a yaml document may result in different sub trees, depending on the used dynaml expressions.

If the read type is set to `import`, the file content is read as yaml document and the root node is used to substitute the expression. Potential dynaml expressions contained in the document will not be evaluated with the actual binding of the expression but as it would have been part of the original file.

#### text documents
A text document will be returned as single string.

### `(( static_ips(0, 1, 3) ))`

Generate a list of static IPs for a job.

e.g.:

```yaml
jobs:
  - name: myjob
    instances: 2
    networks:
    - name: mynetwork
      static_ips: (( static_ips(0, 3, 4) ))
```

This will create 3 IPs from `mynetwork`s subnet, and return two entries, as
there are only two instances. The two entries will be the 0th and 3rd offsets
from the static IP ranges defined by the network.

For example, given the file **bye.yml**:

```yaml
networks: (( merge ))

jobs:
  - name: myjob
    instances: 3
    networks:
    - name: cf1
      static_ips: (( static_ips(0,3,60) ))
```

and file **hi.yml**:

```yaml
networks:
- name: cf1
  subnets:
  - cloud_properties:
      security_groups:
      - cf-0-vpc-c461c7a1
      subnet: subnet-e845bab1
    dns:
    - 10.60.3.2
    gateway: 10.60.3.1
    name: default_unused
    range: 10.60.3.0/24
    reserved:
    - 10.60.3.2 - 10.60.3.9
    static:
    - 10.60.3.10 - 10.60.3.70
  type: manual
```

```
spiff merge bye.yml hi.yml
```

returns


```yaml
jobs:
- instances: 3
  name: myjob
  networks:
  - name: cf1
    static_ips:
    - 10.60.3.10
    - 10.60.3.13
    - 10.60.3.70
networks:
- name: cf1
  subnets:
  - cloud_properties:
      security_groups:
      - cf-0-vpc-c461c7a1
      subnet: subnet-e845bab1
    dns:
    - 10.60.3.2
    gateway: 10.60.3.1
    name: default_unused
    range: 10.60.3.0/24
    reserved:
    - 10.60.3.2 - 10.60.3.9
    static:
    - 10.60.3.10 - 10.60.3.70
  type: manual
```
.

If **bye.yml** was instead

```yaml
networks: (( merge ))

jobs:
  - name: myjob
    instances: 2
    networks:
    - name: cf1
      static_ips: (( static_ips(0,3,60) ))
```

```
spiff merge bye.yml hi.yml
```

instead returns

```yaml
jobs:
- instances: 2
  name: myjob
  networks:
  - name: cf1
    static_ips:
    - 10.60.3.10
    - 10.60.3.13
networks:
- name: cf1
  subnets:
  - cloud_properties:
      security_groups:
      - cf-0-vpc-c461c7a1
      subnet: subnet-e845bab1
    dns:
    - 10.60.3.2
    gateway: 10.60.3.1
    name: default_unused
    range: 10.60.3.0/24
    reserved:
    - 10.60.3.2 - 10.60.3.9
    static:
    - 10.60.3.10 - 10.60.3.70
  type: manual
```

`static_ips`also accepts list arguments, as long as all transitivly contained elements are either again lists or integer values. This allows to abbreviate the list of IPs as follows:

```
  static_ips: (( static_ips([1..5]) ))
```

### `(( ipset(ranges, 3, 3,4,5,6) ))`

While the function [static_ips](#-static_ips0-1-3-) for historical reasons
relies on the structure of a bosh manifest
and works only at dedicated locations in the manifest, the function *ipset*
offers a similar calculation purely based on its arguments. So, the available
ip ranges and the required numbers of IPs are passed as arguments.

The first (ranges) argument can be a single range as a simple string or a
list of strings. Every string might be
- a single IP address
- an explicit IP range described by two IP addresses separated by a dash (-)
- a CIDR

The second argument specifies the requested number of IP addresses in the
result set.

The additional arguments specify the indices of the IPs to choose (starting
from 0) in the given ranges. Here again lists of indices might be used.

e.g.:

```yaml
ranges:
  - 10.0.0.0 - 10.0.0.255
  - 10.0.2.0/24
ipset: (( ipset(ranges,3,[256..260]) ))
```

resolves *ipset* to `[ 10.0.2.0, 10.0.2.1, 10.0.2.2 ]`.

If no IP indices are specified (only two arguments), the IPs are chosen
starting from the beginning of the first range up to the end of the last
given range, without indirection.

### `(( list_to_map(list, "key") ))`

A list of map entries with explicit name/key fields will be mapped to a map with the dedicated keys. By default the key field `name` is used, which can changed by the optional second argument. An explicitly denoted key field in the list will also be taken into account.

e.g.:

```yaml
list:
  - key:foo: alice
    age: 24
  - foo: bob
    age: 30

map: (( list_to_map(list) ))
```

will be mapped to

```yaml
list:
  - foo: alice
    age: 24
  - foo: bob
    age: 30

map:
  alice:
    age: 24
  bob:
    age: 30
```

In combination with templates and lambda expressions this can be used to generate maps with arbitrarily named key values, although dynaml expressions are not allowed for key values.

### `(( makemap(fieldlist) ))`

In this flavor `makemap` creates a map with entries described by the given field list.
The list is expected to contain maps with the entries `key` and `value`, describing
dedicated map entries.

e.g.:

```yaml
list:
  - key: alice
    value: 24
  - key: bob
    value: 25
  - key: 5
    value: 25

map: (( makemap(list) ))
```

yields


```yaml
list:
  - key: alice
    value: 24
  - key: bob
    value: 25
  - key: 5
    value: 25

map:
  "5": 25
  alice: 24
  bob: 25
```

If the key value is a boolean or an integer it will be mapped to a string.

### `(( makemap(key, value) ))`

In this flavor `makemap` creates a map with entries described by the given argument
pairs. The arguments may be a sequence of key/values pairs (given by separate arguments).

e.g.:

```yaml
map: (( makemap("peter", 23, "paul", 22) ))
```

yields


```yaml
map:
  paul: 22
  peter: 23
```

In contrast to the previous `makemap` flavor, this one could also be handled by
[map literals](#--alice--25--).

### `(( merge(map1, map2) ))`

Beside the keyword ` merge` there is also a function called `merge` (It must always be followed by an opening bracket). It can be used to merge severals maps taken from the actual document analogous to the stub merge process. If the maps are specified by reference expressions, they cannot contain
any _dynaml_ expressions, because they are always evaluated in the context of the actual document before evaluating the arguments.

e.g.:

```yaml
map1:
  alice: 24
  bob: (( alice ))
map2:
  alice: 26
  peter: 8
result: (( merge(map1,map2) ))
```

resolves `result` to

```yaml
result:
  alice: 26
  bob: 24  # <---- expression evaluated before mergeing
```

Alternatively map [templates](#templates) can be passed (without evaluation operator!). In this case the _dynaml_ expressions from the template are evaluated while merging the given documents as for regular calls of _spiff merge_.

e.g.:

```yaml
map1:
  <<: (( &template ))
  alice: 24
  bob: (( alice ))
map2:
  alice: 26
  peter: 8
result: (( merge(map1,map2) ))
```

resolves `result` to

```yaml
result:
  alice: 26
  bob: 26
```

A map might also be given by a map expression. Here it is possible to specify
dynaml expressions using the usual syntax:

e.g.:

```yaml
map1:
  alice: 24
  bob: 25

map2:
  alice: 26
  peter: 8

result: (( merge(map1, map2, { "bob"="(( carl ))", "carl"=100 }) ))
```

resolves `result` to

```yaml
result:
  alice: 26
  bob: 100
```

## `(( lambda |x|->x ":" port ))`

Lambda expressions can be used to define additional anonymous functions. They can be assigned to yaml nodes as values and referenced with path expressions to call the function with approriate arguments in other dynaml expressions. For the final document they are mapped to string values.

There are two forms of lambda expressions. While

```yaml
lvalue: (( lambda |x|->x ":" port ))
```

yields a function taking one argument by directly taking the elements from the dynaml expression,

```yaml
string: "|x|->x \":\" port"
lvalue: (( lambda string ))
```

evaluates the result of an expression to a function. The expression must evaluate to a function or string. If the expression is evaluated to a string it parses the function from the string.

Since the evaluation result of a lambda expression is a regular value, it can also be passed as argument to function calls and merged as value along stub processing.

A complete example could look like this:

```yaml
lvalue: (( lambda |x,y|->x + y ))
mod: (( lambda|x,y,m|->(lambda m)(x, y) + 3 ))
value: (( .mod(1,2, lvalue) ))
```

yields

```yaml
lvalue: lambda |x,y|->x + y
mod: lambda|x,y,m|->(lambda m)(x, y) + 3
value: 6
```

A lambda expression might refer to absolute or relative nodes of the actual template. Relative references are evaluated in the context of the function call. Therefore

```yaml
lvalue: (( lambda |x,y|->x + y + offset ))
offset: 0
values:
  offset: 3
  value: (( .lvalue(1,2) ))
```

yields `6` for `values.value`.

Besides the specified parameters, there is an implicit name (`_`), that can be used to refer to the function itself. It can be used to define self recursive function. Together with the logical and conditional operators a fibunacci function can be defined:

```yaml
fibonacci: (( lambda |x|-> x <= 0 ? 0 :x == 1 ? 1 :_(x - 2) + _( x - 1 ) ))
value: (( .fibonacci(5) ))
```

yields the value `8` for the `value` property.

Inner lambda expressions remember the local binding of outer lambda expressions. This can be used to return functions based an arguments of the outer function.

e.g.:

```yaml
mult: (( lambda |x|-> lambda |y|-> x * y ))
mult2: (( .mult(2) ))
value: (( .mult2(3) ))
```

yields `6` for property `value`.

If a lambda function is called with less arguments than expected, the result is a new function taking the missing arguments (currying).

e.g.:

```yaml
mult: (( lambda |x,y|-> x * y ))
mult2: (( .mult(2) ))
value: (( .mult2(3) ))
```

If a complete expression is a lambda expression the keyword `lambda` can be omitted.

## `(( &temporary ))`

Maps, lists or simple value nodes can be marked as *temporary*. Temporary nodes are removed from the final output document, but are available during merging and dynaml evaluation.

e.g.:

```yaml
temp:
  <<: (( &temporary ))
  foo: bar

value: (( temp.foo ))
```

yields:

```yaml
value: bar
```
Adding `- <<: (( &temporary ))` to a list can be used to mark a list as temporary.

The temporary marker can be combined with regular dynaml expressions to tag plain fields. Hereby the
parenthesised expression is just appended to the marker

e.g.:

```yaml
data:
  alice: (( &temporary ( "bar" ) ))
  foo: (( alice ))
```

yields:

```yaml
data:
  foo: bar
```

The temporary marker can be combined with the [template marker](#templates) to omit templates from the final output.

The marker `&local` acts similar to `&temporary` but local nodes are always
removed from a stub directly after resolving dynaml expressions. Such nodes
are therefore not available for merging.

## Mappings

Mappings are used to produce a new list from the entries of a _list_ or _map_ containing the entries processed by a dynaml expression. The expression is given by a [lambda function](#-lambda-x-x--port-). There are two basic forms of the mapping function: It can be inlined as in `(( map[list|x|->x ":" port] ))`, or it can be determined by a regular dynaml expression evaluating to a lambda function as in `(( map[list|mapping.expression))` (here the mapping is taken from the property `mapping.expression`, which should hold an approriate lambda function).


### `(( map[list|elem|->dynaml-expr] ))`

Execute a mapping expression on members of a list to produce a new (mapped) list. The first expression (`list`) must resolve to a list. The last expression (`x ":" port`) defines the mapping expression used to map all members of the given list. Inside this expression an arbitrarily declared simple reference name (here `x`) can be used to access the actually processed list element.

e.g.

```yaml
port: 4711
hosts:
  - alice
  - bob
mapped: (( map[hosts|x|->x ":" port] ))
```

yields

```yaml
port: 4711
hosts:
- alice
- bob
mapped:
- alice:4711
- bob:4711
```

This expression can be combined with others, for example:

```yaml
port: 4711
list:
  - alice
  - bob
joined: (( join( ", ", map[list|x|->x ":" port] ) ))

```

which magically provides a comma separated list of ported hosts:

```yaml
port: 4711
list:
  - alice
  - bob
joined: alice:4711, bob:4711
```

### `(( map[list|idx,elem|->dynaml-expr] ))`

In this variant, the first argument `idx` is provided with the index and the
second `elem` with the value for the index.

e.g.

```yaml
list:
  - name: alice
    age: 25
  - name: bob
    age: 24

ages: (( map[list|i,p|->i + 1 ". " p.name " is " p.age ] ))
```

yields

```yaml
list:
  - name: alice
    age: 25
  - name: bob
    age: 24

ages:
- 1. alice is 25
- 2. bob is 24
```

### `(( map[map|key,value|->dynaml-expr] ))`

Mapping of a map to a list using a mapping expression. The expression may have access to the key and/or the value. If two references are declared, both values are passed to the expression, the first one is provided with the key and the second one with the value for the key. If one reference is declared, only the value is provided.

e.g.

```yaml
ages:
  alice: 25
  bob: 24

keys: (( map[ages|k,v|->k] ))

```

yields

```yaml
ages:
  alice: 25
  bob: 24

keys:
- alice
- bob
```

## Aggregations

Aggregations are used to produce a single result from the entries of a _list_ or _map_ aggregating the entries by a dynaml expression. The expression is given by a [lambda function](#-lambda-x-x--port-). There are two basic forms of the aggregation function: It can be inlined as in `(( sum[list|0|s,x|->s + x] ))`, or it can be determined by a regular dynaml expression evaluating to a lambda function as in `(( sum[list|0|aggregation.expression))` (here the aggregation function  is taken from the property `aggregation.expression`, which should hold an approriate lambda function).


### `(( sum[list|initial|sum,elem|->dynaml-expr] ))`

Execute an aggregation expression on members of a list to produce an aggregation result. The first expression (`list`) must resolve to a list. The second expression is used as initial value for the aggregation. The last expression (`s + x`) defines the aggregation expression used to aggregate all members of the given list. Inside this expression an arbitrarily declared simple reference name (here `s`) can be used to access the intermediate aggregation result and a second reference name (here `x`) can be used to access the actually processed list element.

e.g.

```yaml
list:
  - 1
  - 2
sum: (( sum[list|0|s,x|->s + x] ))
```

yields

```yaml
list:
  - 1
  - 2
sum: 3
```

### `(( sum[list|initial|sum,idx,elem|->dynaml-expr] ))`

In this variant, the second argument `idx` is provided with the index and the
third `elem` with the value for the index.

e.g.

```yaml
list:
  - 1
  - 2
  - 3

prod: (( sum[list|0|s,i,x|->s + i * x ] ))
```

yields

```yaml
list:
  - 1
  - 2
  - 3

prod: 8
```

### `(( sum[map|initial|sum,key,value|->dynaml-expr] ))`

Aggregation of the elements of a map to a single result using an aggregation expression. The expression may have access to the key and/or the value. The first argument is always the intermediate aggregation result. If three references are declared, both values are passed to the expression, the second one is provided with the key and the third one with the value for the key. If two references are declared, only the second one is provided with the value of the map entry.

e.g.

```yaml
ages:
  alice: 25
  bob: 24

sum: (( map[ages|0|s,k,v|->s + v] ))

```

yields

```yaml
ages:
  alice: 25
  bob: 24

sum: 49
```

## Projections

Projections work over the elements of a list or map yielding a result list. Hereby every element is mapped by an optional subsequent reference expression. This may contain again projections, dynamic references or lambda calls. Basically this is a simplified form of the more general [mapping](#mappings) yielding a list working with a lambda function using only a reference expression based on the elements.

### `(( expr.[*].value ))`

All elements of a map or list given by the expression `expr` are dereferenced with the subsequent reference expression (here `.expr`). If this expression works on a map the elements are ordered accoring to their key values. If the subsequent reference expression is omitted, the complete value list isreturned. For a list expression this means the identity operation.

e.g.:

```yaml
list:
  - name: alice
    age: 25
  - name: bob
    age: 26
  - name: peter
    age: 24

names: (( list.[*].name ))
```

yields for `names`:

```yaml
names:
  - alice
  - bob
  - peter
```

or for maps:

```yaml
networks:
  ext:
    cidr: 10.8.0.0/16
  zone1:
    cidr: 10.9.0.0/16

cidrs: (( .networks.[*].cidr ))
```

yields for `cidrs`:

```yaml
cidrs:
  - 10.8.0.0/16
  - 10.9.0.0/16
```

### `(( list.[1..2].value ))`

This projection flavor only works for lists. The projection is done for a dedicated slice of the initial list.

e.g.:

```yaml
list:
  - name: alice
    age: 25
  - name: bob
    age: 26
  - name: peter
    age: 24

names: (( list.[1..2].name ))
```

yields for `names`:

```yaml
names:
  - bob
  - peter
```

## Templates

A map can be tagged by a dynaml expression to be used as template. Dynaml expressions in a template are not evaluated at its definition location in the document, but can be inserted at other locations using dynaml.
At every usage location it is evaluated separately.

### `<<: (( &template ))`

The dynaml expression `&template` can be used to tag a map node as template:

e.g.:

```yaml
foo:
  bar:
    <<: (( &template ))
    alice: alice
    bob: (( verb " " alice ))
```

The template will be the value of the node `foo.bar`. As such it can be overwritten as a whole by settings in a stub during the merge process. Dynaml expressions in the template are not evaluated. A map can have only a single `<<` field. Therefore it is possible to combine the template marker with an expression just by adding the expression in parenthesis.

Adding `- <<: (( &template ))` to a list it is also possible to define list templates.
It is also possible to convert a single expression value into a simple template by adding the template
marker to the expression, for example `foo: (( &template (expression) ))`

The template marker can be combined with the [temporary marker](#-temporary-) to omit templates from the final output.

### `(( *foo.bar ))`

The dynaml expression `*<refernce expression>` can be used to evaluate a template somewhere in the yaml document.
Dynaml expressions in the template are evaluated in the context of this expression.

e.g.:

```yaml
foo:
  bar:
    <<: (( &template ))
    alice: alice
    bob: (( verb " " alice ))


use:
  subst: (( *foo.bar ))
  verb: loves

verb: hates
```

evaluates to

```yaml
foo:
  bar:
    <<: (( &template ))
    alice: alice
    bob: (( verb " " alice ))

use:
  subst:
    alice: alice
    bob: loves alice
  verb: loves

verb: hates
```

## Special Literals

### `(( {} ))`

Provides an empty map.

### `(( [] ))`

Provides an empty list. Basically this is not a dedicated literal, but just a regular list expression without a value.

### `(( ~ ))`

Provides the *null* value.

### `(( ~~ ))`

This literal evaluates to an *undefined* expression. The element (list entry or map field) carrying this value, although defined, will be removed from the document and handled as undefined for further merges and the evaluation of referential expressions.

e.g.:

```yaml
foo: (( ~~ ))
bob: (( foo || ~~ ))
alice: (( bob || "default"))
```

evaluates to

```yaml
alice: default
```

## Access to evaluation context

Inside every dynaml expression a virtual field `__ctx` is available. It allows access to information about the actual evaluation context. It can be accessed by a relative reference expression.

The following fields are supported:

| Field Name  | Type | Meaning |
| ------------| ---- | ------- |
| `FILE` | string | name of actually processed template file  |
| `DIR`  | string | name of directory of actually processed template file  |
| `RESOLVED_FILE` | string | name of actually processed template file with resolved symbolic links |
| `RESOLVED_DIR`  | string | name of directory of actually processed template file with resolved symbolic links |
| `PATHNAME` | string | path name of actually processed field |
| `PATH` | list[string] | path name as component list |

e.g.:

**template.yml**
```yaml
foo:
  bar:
    path: (( __ctx.PATH ))
    str: (( __ctx.PATHNAME ))
    file: (( __ctx.FILE ))
    dir: (( __ctx.DIR ))
```

evaluates to

e.g.:

```yaml
foo:
  bar:
    dir: .
    file: template.yml
    path:
    - foo
    - bar
    - path
    str: foo.bar.str
```

## Operation Priorities

Dynaml expressions are evaluated obeying certain priority levels. This means operations with a higher priority are evaluated first. For example the expression `1 + 2 * 3` is evaluated in the order `1 + ( 2 * 3 )`. Operations with the same priority are evaluated from left to right (in contrast to version 1.0.7). This means the expression `6 - 3 - 2` is evaluated as `( 6 - 3 ) - 2`.

The following levels are supported (from low priority to high priority)

1. `||`
2. White-space separated sequence as concatenation operation (`foo bar`)
3. `-or`, `-and`
4. `==`, `!=`, `<=`, `<`, `>`, `>=`
5. `+`, `-`
6. `*`, `/`, `%`
7. Grouping `( )`, `!`, constants, references (`foo.bar`), `merge`, `auto`, `lambda`, `map[]`, and [functions](#functions)

The complete grammar can be found in [dynaml.peg](dynaml/dynaml.peg).

# Structural Auto-Merge

By default `spiff` performs a deep structural merge of its first argument, the template file, with the given stub files. The merge is processed from right to left, providing an intermediate merged stub for every step. This means, that for every step all expressions must be locally resolvable.

Structural merge means, that besides explicit dynaml `merge` expressions, values will be overridden by values of equivalent nodes found in right-most stub files. In general, flat value lists are not merged. Only lists of maps can be merged by entries in a stub with a matching index.

There is a special support for the auto-merge of lists containing maps, if the
maps contain a `name` field. Hereby the list is handled like a map with
entries according to the value of the list entries' `name` field. If another
key field than `name` should be used, the key field of one list entry can be
tagged with the prefix `key:` to indicate the indended key name. Such tags
will be removed for the processed output.

In general the resolution of matching nodes in stubs is done using the same rules that apply for the reference expressions [(( foo.bar.[1].baz ))](#-foobar1baz-).

For example, given the file **template.yml**:

```yaml
foo:
  - name: alice
    bar: template
  - name: bob
    bar: template

plip:
  - id: 1
    plop: template
  - id: 2
    plop: template

bar:
  - foo: template

list:
  - a
  - b
```

and file **stub.yml**:

```yaml
foo:
  - name: bob
    bar: stub

plip:
  - key:id: 1
    plop: stub

bar:
  - foo: stub

list:
  - c
  - d
```

```
spiff merge template.yml stub.yml
```

returns


```yaml
foo:
- bar: template
  name: alice
- bar: stub
  name: bob

plip:
- id: 1
  plop: stub
- id: 2
  plop: template

bar:
- foo: stub

list:
- a
- b
```

Be careful that any `name:` key in the template for the first element of the
`plip` list will defeat the `key:id: 1` selector from the stub. When a `name`
field exist in a list element, then this element can only be targeted by this
name. When the selector is defeated, the resulting value is the one provided
by the template.

## Bringing it all together

Merging the following files in the given order

**deployment.yml**
```yaml
networks: (( merge ))
```

**cf.yml**
```yaml
utils: (( merge ))
network: (( merge ))
meta: (( merge ))

networks:
  - name: cf1
    <<: (( utils.defNet(network.base.z1,meta.deployment_no,30) ))
  - name: cf2
    <<: (( utils.defNet(network.base.z2,meta.deployment_no,30) ))
```

**infrastructure.yml**
```yaml
network:
  size: 16
  block_size: 256
  base:
    z1: 10.0.0.0
    z2: 10.1.0.0
```

**rules.yml**
```yaml
utils:
  defNet: (( |b,n,s|->(*.utils.network).net ))
  network:
    <<: (( &template ))
    start: (( b + n * .network.block_size ))
    first: (( start + ( n == 0 ? 2 :0 ) ))
    lower: (( n == 0 ? [] :b " - " start - 1 ))
    upper: (( start + .network.block_size " - " max_ip(net.subnets.[0].range) ))
    net:
      subnets:
      - range: (( b "/" .network.size ))
        reserved: (( [] lower upper ))
        static:
          - (( first " - " first + s - 1 ))
```

**instance.yml**
```yaml
meta:
  deployment_no: 1

```

will yield a network setting for a dedicated deployment

```yaml
networks:
- name: cf1
  subnets:
  - range: 10.0.0.0/16
    reserved:
    - 10.0.0.0 - 10.0.0.255
    - 10.0.2.0 - 10.0.255.255
    static:
    - 10.0.1.0 - 10.0.1.29
- name: cf2
  subnets:
  - range: 10.1.0.0/16
    reserved:
    - 10.1.0.0 - 10.1.0.255
    - 10.1.2.0 - 10.1.255.255
    static:
    - 10.1.1.0 - 10.1.1.29
```

Using the same config for another deployment of the same type just requires the replacement of the `instance.yml`.
Using a different `instance.yml`

```yaml
meta:
  deployment_no: 0

```

will yield a network setting for a second deployment providing the appropriate settings for a unique other IP block.

```yaml
networks:
- name: cf1
  subnets:
  - range: 10.0.0.0/16
    reserved:
    - 10.0.1.0 - 10.0.255.255
    static:
    - 10.0.0.2 - 10.0.0.31
- name: cf2
  subnets:
  - range: 10.1.0.0/16
    reserved:
    - 10.1.1.0 - 10.1.255.255
    static:
    - 10.1.0.2 - 10.1.0.31
```

If you move to another infrastructure you might want to change the basic IP layout. You can do it just by adapting the `infrastructure.yml`

```yaml
network:
  size: 17
  block_size: 128
  base:
    z1: 10.0.0.0
    z2: 10.0.128.0
```

Without any change to your other settings you'll get

```yaml
networks:
- name: cf1
  subnets:
  - range: 10.0.0.0/17
    reserved:
    - 10.0.0.128 - 10.0.127.255
    static:
    - 10.0.0.2 - 10.0.0.31
- name: cf2
  subnets:
  - range: 10.0.128.0/17
    reserved:
    - 10.0.128.128 - 10.0.255.255
    static:
    - 10.0.128.2 - 10.0.128.31
```

## Useful to Know

  There are several scenarios yielding results that do not seem to be obvious. Here are some typical pitfalls.

- _The auto merge never adds nodes to existing structures_

  For example, merging

  **template.yml**
  ```yaml
  foo:
    alice: 25
  ```
  with

  **stub.yml**
  ```yaml
  foo:
    alice: 24
    bob: 26
  ```

   yields

  ```yaml
  foo:
    alice: 24
  ```

  Use [<<: (( merge ))](#--merge-) to change this behaviour, or explicitly add desired nodes to be merged:

   **template.yml**
  ```yaml
  foo:
    alice: 25
	bob: (( merge ))
  ```


- _Simple node values are replaced by values or complete structures coming from stubs, structures are deep_ merged.

  For example, merging

  **template.yml**
  ```yaml
  foo: (( ["alice"] ))
  ```
  with

  **stub.yml**
  ```yaml
  foo:
    - peter
    - paul
  ```

  yields

  ```yaml
  foo:
    - peter
    - paul
  ```

  But the template

  ```yaml
   foo: [ (( "alice" )) ]
  ```

  is merged without any change.

- _Expressions are subject to be overridden as a whole_

  A consequence of the behaviour described above is that nodes described by an expession are basically overridden by a complete merged structure, instead of doing a deep merge with the structues resulting from the expression evaluation.

  For example, merging

  **template.yml**
  ```yaml
  men:
    - bob: 24
  women:
    - alice: 25

  people: (( women men ))
  ```
  with

  **stub.yml**
  ```yaml
  people:
    - alice: 13
  ```
   yields

  ```yaml
  men:
    - bob: 24
  women:
    - alice: 25

  people:
    - alice: 24
  ```

  To request an auto-merge of the structure resulting from the expression evaluation, the expression has to be preceeded with the modifier `prefer` (`(( prefer women men ))`). This would yield the desired result:

  ```yaml
  men:
    - bob: 24
  women:
    - alice: 25

  people:
    - alice: 24
    - bob: 24
  ```

- _Nested merge expressions use implied redirections_

  `merge` expressions implicity use a redirection implied by an outer redirecting merge. In the following
  example

  ```yaml
  meta:
    <<: (( merge deployments.cf ))
    properties:
      <<: (( merge ))
      alice: 42
  ```
  the merge expression in `meta.properties` is implicity redirected to the path `deployments.cf.properties`
  implied by the outer redirecting `merge`. Therefore merging with

  ```yaml
  deployments:
    cf:
      properties:
        alice: 24
        bob: 42
  ```

  yields

  ```yaml
  meta:
    properties:
      alice: 24
      bob: 42
  ```

- _Functions and mappings can freely be nested_

  e.g.:

  ```yaml
  pot: (( lambda |x,y|-> y == 0 ? 1 :(|m|->m * m)(_(x, y / 2)) * ( 1 + ( y % 2 ) * ( x - 1 ) ) ))
  seq: (( lambda |b,l|->map[l|x|-> .pot(b,x)] ))
  values: (( .seq(2,[ 0..4 ]) ))
  ```

  yields the list `[ 1,2,4,8,16 ]` for the property `values`.

- _Functions can be used to parameterize templates_

  The combination of functions with templates can be use to provide functions yielding complex structures.
  The parameters of a function are part of the scope used to resolve reference expressions in a template used in the function body.

  e.g.:

  ```yaml
  relation:
    template:
      <<: (( &template ))
      bob: (( x " " y ))
    relate: (( |x,y|->*relation.template ))

  banda: (( relation.relate("loves","alice") ))
  ```

  evaluates to

  ```yaml
  relation:
    relate: lambda|x,y|->*(relation.template)
    template:
      <<: (( &template ))
      bob: (( x " " y ))

	banda:
      bob: loves alice
  ```

- _Aggregations may yield complex values by using templates_

  The expression of an aggregation may return complex values by returning inline lists or instantiated templates. The binding of the function will be available (as usual) for the evaluation of the template. In the example below the aggregation provides a map with both the sum and the product of the list entries containing the integers from 1 to 4.

  e.g.:

  ```yaml
  sum: (( sum[[1..4]|init|s,e|->*temp] ))

  temp:
    <<: (( &template ))
    sum: (( s.sum + e ))
    prd: (( s.prd * e ))
  init:
    sum: 0
    prd: 1
	```

  yields for `sum` the value
  ```
  sum:
    prd: 24
    sum: 10
  ```

- _Taking advantage of the *undefined* value_

  At first glance it might look strange to introduce a value for *undefined*. But it can be really
  useful as will become apparent with the following examples.

  - Whenever a stub syntactically defines a field it overwrites the default in the template during
    merging. Therefore it would not be possible to define some expression for that field that eventually
	keeps the default value. Here the *undefined* value can help:

    e.g.: merging

    **template.yml**
    ```yaml
    alice: 24
    bob: 25
    ```

    with

    **stub.yml**
    ```yaml

    alice: (( config.alice * 2 || ~ ))
    bob: (( config.bob * 3 || ~~ ))
    ```

    yields

    ```yaml
    alice: ~
    bob: 25
    ```

  * There is a problem accessing upstream values. This is only possible if the local stub contains
    the definition of the field to use. But then there will always be a value for this field, even
	if the upstream does not overwrite it.

    Here the *undefined* value can help by providing optional access to upstream values.
	Optional means, that the field is only defined, if there is an upstream value. Otherwise it is
	undefined for the expressions in the local stub and potential downstream templates. This is
	possible because the field is formally defined, and will therefore be merged, only after evaluating
	the expression if it is not merged it will be removed again.

    e.g.: merging

    **template.yml**
    ```yaml
    alice: 24
    bob: 25
    peter: 26
    ```

    with

    **mapping.yml**
    ```yaml
    config:
      alice: (( ~~ ))
	  bob: (( ~~ ))

    alice: (( config.alice || ~~ ))
    bob: (( config.bob || ~~ ))
    peter: (( config.peter || ~~ ))
    ```

    and

    **config.yml**
    ```yaml
    config:
      alice: 4711
	  peter: 0815
    ```
    yields

    ```yaml
    alice: 4711  # transferred from config's config value
    bob: 25      # kept default value, because not set in config.yml
    peter: 26    # kept, because mapping source not available in mapping.yml
    ```

  This can be used to add an intermediate stub, that offers a dedicated
  configuration interface and contains logic to map this interface to a manifest
  structure already defining default values.

- _Templates versus map literals_

  As described earlier templates can be used inside functions and mappings to
  easily describe complex data structures based on expressions refering to
  parameters. Before the introduction of map literals this was the only way
  to achieve such behaviour. The advantage is the possibility to describe
  the complex structure as regular part of a yaml document, which allows using
  the regular yaml formatting  facilitating readability.

  e.g.:

  ```yaml
  scaling:
    runner_z1: 10
    router_z1: 4

    jobs: (( sum[scaling|[]|s,k,v|->s [ *templates.job ] ] ))

  templates:
    job:
      <<: (( &template ))
      name: (( k ))
      instances: (( v ))
  ```

  evaluates to

  ```yaml
  scaling:
    runner_z1: 10
    router_z1: 4

  jobs:
    - instances: 4
      name: router_z1
    - instances: 10
      name: runner_z1
    ...
  ```

  With map literals this construct can significantly be simplified

  ```yaml
  scaling:
    runner_z1: 10
    router_z1: 4

  jobs:  (( sum[scaling|[]|s,k,v|->s [ {"name"=k, "value"=v} ] ] ))
  ```

  Nevertheless the first, template based version might still be useful, if
  the data structures are more complex, deeper or with complex value expressions.
  For such a scenario the description of the data structure as template should be
  preferred. It provides a much better readability, because every field, list
  entry and value expression can be put into dedicated lines.

  But there is still a qualitative difference. While map literals are part of a
  single expression always evaluated as a whole before map fields are available
  for referencing, templates are evaluated as regular yaml documents that might
  contain multiple fields with separate expressions referencing each other.

  e.g.:

  ```yaml
  range: (( (|cidr,first,size|->(*templates.addr).range)("10.0.0.0/16",10,255) ))

  templates:
    addr:
      <<: (( &template ))
      base: (( min_ip(cidr) ))
      start: (( base + first ))
	  end: (( start + size - 1 ))
	  range: (( start " - " end ))
  ```

  evaluates `range` to

  ```yaml
  range: 10.0.0.10 - 10.0.1.8
  ...
  ```

# Error Reporting

The evaluation of dynaml expressions may fail because of several reasons:
- it is not parseable
- involved references cannot be satisfied
- arguments to operations are of the wrong type
- operations fail
- there are cyclic dependencies among expressions

If a dynaml expression cannot be resolved to a value, it is reported by the
`spiff merge` operation using the following layout:

```
	(( <failed expression> ))	in <file>	<path to node>	(<referred path>)	<tag><issue>
```

e.g.:

```
	(( min_ip("10") ))	in source.yml	node.a.[0]	()	*CIDR argument required
```

Cyclic dependencies are detected by iterative evaluation until the document is unchanged after a step.
Nodes involved in a cycle are therefore typically reported just as unresolved node without a specific issue.

The order of the reported unresolved nodes depends on a classification of the problem, denoted by a dedicated
tag. The following tags are used (in reporting order):

| Tag | Meaning |
| --- | ------- |
| `*` | error in local dynaml expression |
| `@` | dependent or involved in cyclic dependencies |
| `-` | subsequent error because of refering to a yaml node with an error |

Problems occuring during inline template processing are reported as nested problems. The classification is
propagated to the outer node.


