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

spiff is a command line tool and declarative YAML templating system, specially designed for generating deployment
manifests (for example BOSH or [kubernetes](https://github.com/kubernetes) manifests).

Contents:
- [Installation](#installation)
- [Usage](#usage)
- [Libraries](#libraries)
- [dynaml Templating Language](#dynaml-templating-language)
	- [(( foo ))](#-foo-)
	- [(( foo.bar.[1].baz ))](#-foobar1baz-)
	- [(( foo.[bar].baz ))](#-foobarbaz-)
	- [(( list.[1..3] ))](#-list13-)
	- [(( 1.2e4 ))](#-12e4-)
	- [(( "foo" ))](#-foo-)
	- [(( [ 1, 2, 3 ] ))](#--1-2-3--)
	- [(( { "alice" = 25 } ))](#--alice--25--)
	- [(( ( "alice" = 25 ) alice ))](#--alice--25---alice-)
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
		- [<<: (( merge none ))](#--merge-none-)
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
		- [(( basename(path) ))](#-basenamepath-)
		- [(( dirname(path) ))](#-dirnamepath-)
		- [(( parseurl("http://github.com") ))](#-parseurlhttpgithubcom-)
		- [(( sort(list) ))](#-sortlist-)
		- [(( replace(string, "foo", "bar") ))](#-replacestring-foo-bar-)
		- [(( substr(string, 1, 3) ))](#-substrstring-1-3-)
		- [(( match("(f.*)(b.*)", "xxxfoobar") ))](#-matchfb-xxxfoobar-)
		- [(( keys(map) ))](#-keysmap-)
		- [(( length(list) ))](#-lengthlist-)
		- [(( base64(string) ))](#-base64string-)
		- [(( hash(string) ))](#-hashstring-)
		- [(( bcrypt("password", 10) ))](#-bcryptpassword-10-)
		- [(( bcrypt_check("password", hash) ))](#-bcrypt_checkpassword-hash-)
		- [(( md5crypt("password") ))](#-md5cryptpassword-)
		- [(( md5crypt_check("password", hash) ))](#-md5crypt_checkpassword-hash-)
		- [(( decrypt("secret") ))](#-decryptsecret-)
		- [(( rand("[:alnum:]", 10) ))](#-randalnum-10-)
		- [(( type(foobar) ))](#-typefoobar-)
		- [(( defined(foobar) ))](#-definedfoobar-)
		- [(( valid(foobar) ))](#-validfoobar-)
		- [(( require(foobar) ))](#-requirefoobar-)
		- [(( stub(foo.bar) ))](#-stubfoobar-)
		- [(( eval(foo "." bar ) ))](#-evalfoo--bar--)
		- [(( env( HOME" ) ))](#-envHOME--)
		- [(( static_ips(0, 1, 3) ))](#-static_ips0-1-3-)
		- [(( ipset(ranges, 3, 3,4,5,6) ))](#-ipsetranges-3-3456-)
		- [(( list_to_map(list, "key") ))](#-list_to_maplist-key-)
		- [(( makemap(fieldlist) ))](#-makemapfieldlist-)
		- [(( makemap(key, value) ))](#-makemapkey-value-)
		- [(( merge(map1, map2) ))](#-mergemap1-map2-)
		- [(( intersect(list1, list2) ))](#-intersectlist1-list2-)
		- [(( reverse(list) ))](#-reverselist-)
		- [(( parse(yamlorjson) ))](#-parseyamlorjson-)
		- [(( asjson(expr) ))](#-asjsonexpr-)
		- [(( asyaml(expr) ))](#-asjsonexpr-)
		- [(( catch(expr) ))](#-catchexpr-)
		- [(( validate(value,"dnsdomain") ))](#-validatevaluednsdomain-)
		- [(( check(value,"dnsdomain") ))](#-checkvaluednsdomain-)
		- [(( error("message") ))](#-errormessage-)
		- [Math](#math)
		- [Conversions](#conversions)
		- [Accessing External Content](#accessing-external-content)
		    - [(( read("file.yml") ))](#-readfileyml-)
		    - [(( exec("command", arg1, arg2) ))](#-execcommand-arg1-arg2-)
            - [(( pipe(data, "command", arg1, arg2) ))](#-pipedata-command-arg1-arg2-)
		    - [(( write("file.yml", data) ))](#-writefileyml-data-)
		    - [(( tempfile("file.yml", data) ))](#-tempfilefileyml-data-)
		    - [(( lookup_file("file.yml", data) ))](#-lookup_filefileyml-list-)
		    - [(( mkdir("dir", 0755) ))](#-mkdirdir-0755-)
		    - [(( list_files(".") ))](#-list_files-)
		    - [(( archive(files, "tar") ))](#-archivefiles-tar-)
		- [X509 Functions](#x509-functions)
		    - [(( x509genkey(spec) ))](#-x509genkeyspec-)
		    - [(( x509publickey(key) ))](#-x509publickeykey-)
		    - [(( x509cert(spec) ))](#-x509certspec-)
		- [Wireguard Functions](#wireguard-functions)
            - [(( wggenkey() ))](#-wggenkey-)
        	- [(( wgpublickey(key) ))](#-wgpublickey-)
	- [(( lambda |x|->x ":" port ))](#-lambda-x-x--port-)
	    - [Positional versus Named Argunments](#positional-versus-named-arguments)
	    - [Scopes and Lambda Expressions](#scopes-and-lambda-expressions)
	    - [Optional Parameters (( |x,y=2|-> x * y ))](#optional-parameters)
	    - [Variable Argument Lists (( |x,y...|-> x y ))](#variable-argument-lists)
	    - [Currying (( function*(1) ))](#currying)
	- [(( catch[expr|v,e|->v] ))](#-catchexprve-v-)
	- [(( sync[expr|v,e|->defined(v.field),v.field|10] ))](#-syncexprve-definedvfieldvfield10-)
	- [Inline List Expansion (( [a, list..., b] ))](#inline-list-expansion)
	- [Mappings](#mappings)
		- [(( map[list|elem|->dynaml-expr] ))](#-maplistelem-dynaml-expr-)
		- [(( map[list|idx,elem|->dynaml-expr] ))](#-maplistidxelem-dynaml-expr-)
		- [(( map[map|key,value|->dynaml-expr] ))](#-mapmapkeyvalue-dynaml-expr-)
		- [(( map{map|elem|->dynaml-expr} ))](#-mapmapelem-dynaml-expr-)
		- [(( map{list|elem|->dynaml-expr} ))](#-maplistelem-dynaml-expr-)
		- [(( select[expr|elem|->dynaml-expr] ))](#-selectexprelem-dynaml-expr-)
		- [(( select{map|elem|->dynaml-expr} ))](#-selectmapelem-dynaml-expr-)
	- [Aggregations](#aggregations)
		- [(( sum[list|initial|sum,elem|->dynaml-expr] ))](#-sumlistinitialsumelem-dynaml-expr-)
		- [(( sum[list|initial|sum,idx,elem|->dynaml-expr] ))](#-sumlistinitialsumidxelem-dynaml-expr-)
		- [(( sum[map|initial|sum,key,value|->dynaml-expr] ))](#-summapinitialsumkeyvalue-dynaml-expr-)
	- [Projections](#projections)
	    - [(( expr.[*].value ))](#-exprvalue-)
		- [(( list.[1..2].value ))](#-list12value-)
	- [Markers](#markers)
	    - [(( &temporary ))](#-temporary-)
	    - [(( &local ))](#-local-)
    	- [(( &inject ))](#-inject-)
    	- [(( &default ))](#-default-)
    	- [(( &state ))](#-state-)
	- [Templates](#templates)
		- [<<: (( &template ))](#--template-)
		- [(( *foo.bar ))](#-foobar-)
	- [Scope References](#scope-references)
	    - [_](#_)
	    - [__](#__)
	    - [___](#___)
	    - [__ctx.OUTER](#__ctxouter)
	- [Special Literals](#special-literals)
	- [Access to evaluation context](#access-to-evaluation-context)
	- [Operation Priorities](#operation-priorities)
	- [String Interpolation](#string-interpolation)
- [Structural Auto-Merge](#structural-auto-merge)
- [Bringing it all together](#bringing-it-all-together)
- [Useful to Know](#useful-to-know)
- [Error Reporting](#error-reporting)
- [Using _spiff_ as Go Library](#using-spiff-as-go-library)


# Installation

Official release executable binaries can be downloaded via [Github releases](https://github.com/mandelsoft/spiff/releases) for Darwin, Linux ans PowerPc machines (and virtual machines).

Some of spiff's dependencies have changed since the last official release, and spiff will not be updated to keep up with these dependencies.  Working dependencies are vendored in the `Godeps` directory (more information on the `godep` tool is available [here](https://github.com/tools/godep)).  As such, trying to `go get` spiff will likely fail; the only supported way to use spiff is to use an official binary release.

# Usage

### `spiff merge template.yml [template2.yml ...]`

Merge a bunch of template files into one manifest, printing it out.

See 'dynaml templating language' for details of the template file, or examples/ subdir for more complicated examples.

Example:

```
spiff merge cf-release/templates/cf-deployment.yml my-cloud-stub.yml
```

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

The ` merge` command offers several options:

- The option `--partial`. If this option is
  given spiff handles incomplete expression evaluation. All errors are ignored
  and the unresolvable parts of the yaml document are returned as strings.
  
- With the option `--json` the output will be in JSON format instead of YAML.

- The option `--path <path>` can be used to output a nested path, instead of the 
  the complete processed document.
  
- If the output is a list, the option `--split` outputs every list element as
  separate documen. The _yaml_ format uses as usual `---` as separator line.
  The _json_ format outputs a sequence of _json_ documents, one per line.
  
- With `--select <field path>` it is possible to select a dedicated field of the
  processed document for the output
  
- With `--evaluate <dynaml expression>` it is possible to evaluate a given dynaml
  expression on the processed document for the output. The expression is evaluated
  before the selection path is applied, which will then work on the evaluation
  result.
  
- The option `--state <path>` enables the state support of _spiff_. If the
  given file exists it is put on top of the configured stub list for the
  given file exists it is put on top of the configured stub list for the
  merge processing. Additionally to the output of the processed document
  it is filtered for nodes marked with the [`&state` marker](#-state-).
  This filtered document is then stored under the denoted file, saving the old
  state file with the `.bak` suffix. This can be used together with a manual
  merging as offered by the [state](libraries/state/README.md) utility library.
  
- With option `--bindings <path>` a yaml file can be specified, whose content
  is used to build additional bindings for the processing. The yaml document must
  consist of a map. Each key is used as additional binding. The bindings document
  is not processed, the values are used as defined.

- With option `--define <key>=<value>` (shorthand`-D`) additional binding values
  can be specified on the command line overriding binding values from the
  binding file. The option may occur multiple times.
  
- The option `--preserve-escapes` will preserve the escaping for dynaml
  expressions and list/map merge directives. This option can be used
  if further processing steps of a processing result with *spiff* is intended.

- The option `--preserve-temporary` will preserve the fields marked as temporary
  in the final document.
  
The folder [libraries](libraries/README.md) offers some useful
utility libraries. They can also be used as an example for the power
of this templating engine.


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

### `spiff convert --json manifest.yml `

The `convert` sub command can be used to convert input files to json or
just to normalize the order of the fields.
Available options are `--json`, `--path`, `--split` or `--select` according
to their meanings for the `merge` sub command.

### `spiff encrypt secret.yaml`

The `encrypt` sub command can be used to encrypt or decrypt data
according to the [`encrypt`](#-decryptsecret-) dynaml function.
The password can be given as second argument or it is taken from the
environment variable `SPIFF_ENCRYPTION_KEY`. The last argument can be used
to pass the encryption method (see [`encrypt` function](#-decryptsecret-))

The data is taken from the specified file. If `-` is given, it is read from
stdin.

If the option `-d` is given, the data is decrypted, otherwise the data is
read as yaml document and the encrypted result is printed. 

# Libraries

The [libraries](libraries/README.md) folder contains some useful _spiff_ template
libraries. These are basically just stubs that are added to the merge file list
to offer the utility functions for the merge processing.

# dynaml Templating Language

Spiff uses a declarative, logic-free templating language called 'dynaml'
(dynamic yaml).

Every dynaml node is guaranteed to resolve to a YAML node. It is *not*
string interpolation. This keeps developers from having to think about how
a value will render in the resulting template.

A dynaml node appears in the .yml file as a string denoting an expression
surrounded by two parentheses `(( <dynaml> ))`. They can be used as the
value of a map or an entry in a list. The expression might span multiple
lines. In any case the yaml string value *must not* end with a newline
(for example using `|-`)

If a parenthesized value should not be interpreted as an *dynaml* expression and
kept as it is in the output, it can be escaped by an exclamation mark directly
after the openeing brackets.

For example, `((! .field ))` maps to the string value `(( .field ))` and
`((!! .field ))` maps to the string value `((! .field ))`.

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

Look for the nearest 'foo' key, and from there follow through to `.bar.[1].baz`.

A path is a sequence of steps separated by dots. A step is either a word for
maps, or digits surrounded by brackets for list indexing. The index might be negative (a minus followed by digits). Negative indices are taken from then end
of the list (effective index = index + length(list)).

A path that cannot be resolved lead to an evaluation error. If a reference is
expected to sometimes not be provided, it should be
used in combination with '||' (see [below](#-a--b-)) to guarantee resolution.

**Note**: The dynaml grammer has been reworked to enable the usual index syntax,
now. Instead of `foo.bar.[1]` it is possible now to use `foo.bar[1]`.

**Note**: References are always within the template or stub, and order does not
matter. You can refer to another dynamic node and presume it's resolved, and the
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
by their name value as path component.

e.g.:

The age of alice in

```yaml
list:
 - name: alice
   age: 25
```

can be referenced by using the path `list.alice.age`, instead of `list[0].age`.

By default a field with name `name` is used as key field. If another field
should be used as key field, it can be marked in one list entry as key by
prefixing the field name with the keyword `key:`. This keyword is removed
from by the processing and will not be part of the final processing result.

e.g.:

```yaml
list:
 - key:person: alice
   age: 25

alice: (( list.alice ))
```

will be resolved to

```yaml
list:
 - person: alice
   age: 25

alice:
  person: alice
  age: 25
```

This new key field will also be observed during the merging of lists.

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
from the end of the list from *length+start* to *length+end*. If the end
index is lower than the start index, the result is an empty array.

e.g.:

```yaml
list:
  - a
  - b
  - c
foo: (( list.[1..length(list) - 1] ))
```

The start or end index might be omitted. It is then selected according to the 
actual size of the list. Therefore `list.[1..length(list)]` is equivalent
to `list.[1..]`.

evaluates `foo` to the list `[b,c]`.

## `(( 1.2e4 ))`

Number literatls are supported for integers and floating point values.

## `(( "foo" ))`

String literal. All [json string encodings](https://www.json.org/) are supported
(for exmple `\n`, `\"` or `\uxxxx`).

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

## `(( ( "alice" = 25 ) alice ))`

Any expression may be preluded by any number of explicit _scope literals_. A
scope literal describes a map whose values are available for relative reference 
resolution of the expression (static scope). It creates an additional local
binding for given names.

A scope literal might consist of any number of field assignments separated by a
comma `,`. The key as well as the value are given by expressions, whereas the
key expression must evaluate to a string. All expressions are evaluated in the
next outer scope, this means later settings in a scope _cannot_ use earlier
settings in the same scope literal. 

e.g.:

```yaml
scoped: (( ( "alice" = 25, "bob" = 26 ) alice + bob ))
```

yields

```yaml
scoped: 51
```

A field name might also be denoted by a symbol (_`$`name_).

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
returning the value from the last stub that provides it.

If the corresponding value is not defined, it will return nil. This then has the
same semantics as reference expressions; a nil merge is an unresolved template.
See [`||`](#-a--b-).

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

**Note**: Instead of using a `<<:` insert field to place merge expressions it is
possible now to use `<<<:`, also, which allows to use regular yaml parsers for
spiff-like yaml documents. `<<:` is kept for backward compatibility.
be used (`(( merge required ))`).

If the merge key should not be interpreted as regular key instead of a merge
directive, it can be escaped by an excalamtion mark (`!`).

For example, a map key `<<<!` will result in a string key `<<<` and `<<<!!`
will result in a string key `<<<!`

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

### `<<: (( merge none ))`

If the reference of an redirecting merge is set to the constant `none`,
no merge is done at all. This expressions always yields the nil value.

e.g.: for

**template.yml**
```yaml
map:
  <<: (( merge none ))
  value: notmerged
```

**values.yml**
```yaml
map:
  value: merged
```

`spiff merge template.yml values.yml` yields:

```yaml
map:
  value: notmerged
```

This can be used for explicit field merging using the `stub` function
to access dedicated parts of upstream stubs.

e.g.:

**template.yml**
```yaml
map:
  <<: (( merge none ))
  value: ((  "alice"  "+" stub(map.value) ))
```

**values.yml**
```yaml
map:
  value: bob
```

`spiff merge template.yml values.yml` yields:

```yaml
test:
  value: alice+bob
```

This also works for dedicated fields:

**template.yml**
```yaml
map:
  value: ((  merge none // "alice"  "+" stub() ))
```

**values.yml**
```yaml
map:
  value: bob
```

`spiff merge template.yml values.yml` yields:

```yaml
test:
  value: alice+bob
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

The operator `//` additionally checks, whether `a` can be solved to a valid 
value (not equal `~`).

## `(( 1 + 2 * foo ))`

Dynaml expressions can be used to execute arithmetic integer and floating-point calculations. Supported operations are `+`, `-`, `*`, and `/`.
The modulo operator (`%`) only supports integer operands.

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
contains: (( contains_ip(cidr, "192.168.0.2") ))
```

yields

```yaml
cidr: 192.168.0.1/24
range: 192.168.0.0-192.168.0.255
next: 192.168.1.0
num: 192.168.0.0+256=192.168.1.0
contains: true
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

An expression is considered to be `false` if it evaluates to
- the boolean value `false`
- the integer value 0
- an empty string, map or list

Otherwise it is considered to be `true`


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
Instead of a separator string an integer value might be given,
which splits the give string into list of length limited strings.
The length is counted in runes, not bytes.

e.g.:

```yaml
list: (( split("," "alice, bob") ))
limited: (( split(4, "1234567890") ))

```

yields:

```yaml
list:
  - alice
  - ' bob'
limited:
  - "1234"
  - "5678"
  - "90"

```

An optional 3rd argument might be specified. It limits the number of returned
list entries. The value -1 leads to an unlimited list length.

If a [regular expression](https://github.com/google/re2/wiki/Syntax) should
be used as separator string, the function `split_match` can be used.


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

### `(( basename(path) ))`

The function `basename` returns the name of the last element of a path.
The argument may either be a regular path name or a URL.

e.g.:

```yaml
pathbase:  (( basename("alice/bob") ))
urlbase:  (( basename("http://foobar/alice/bob?any=parameter") ))
```

yields:

```yaml
pathbase:  bob
urlbase:  bob
```

### `(( dirname(path) ))`

The function `dirname` returns the parent directory of a path.
The argument may either be a regular path name or a URL.

e.g.:

```yaml
pathbase:  (( dirname("alice/bob") ))
urlbase:  (( dirname("http://foobar/alice/bob?any=parameter") ))
```

yields:

```yaml
pathbase:  alice
urlbase:  /alice
```

### `(( parseurl("http://github.com") ))`

This function parses a URL and yield a map with all elements of an URL.
The fields `port`, `userinfo`and `password` are optional.

e.g.:

```yaml
url:  (( parseurl("https://user:pass@github.com:443/mandelsoft/spiff?branch=master&tag=v1#anchor") ))
```

yields:

```yaml
url:
  scheme: https
  host: github.com
  port: 443
  path: /mandelsoft/spiff
  fragment: anchor
  query: branch=master&tag=v1
  values:
    branch: [ master ]
    tag: [ v1 ]
  userinfo:
    username: user
    password: pass
```



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

### `(( sort(list) ))

The function `sort` can be used to sort integer or string lists. The sort
operation is stable.

e.g.:

```yaml
list:
  - alice
  - foobar
  - bob

sorted: (( sort(list) ))

```

yields for `sorted`

```yaml
- alice
- bob
- foobar

```

If other types should be sorted, especially complex types like lists or maps, or
a different comparison rule is required, a
compare function can be specified as an optional second argument. The compare
function must be a lambda expression taking two arguments. The result type
must be `integer`or `bool`  indicating whether _a_ is less then _b_. If an
integer is returned it should be
- negative, if _a<b_
- zero, if _a==b_ and
- positive if _a>b_

e.g.:

```yaml
list:
  - alice
  - foobar
  - bob

sorted: (( sort(list, |a,b|->length(a) < length(b)) ))

```

yields for `sorted`

```yaml
- bob
- alice
- foobar

```


### `(( replace(string, "foo", "bar") ))`

Replace all occurences of a sub string in a string by a replacement string. With an optional
fourth integer argument the number of substitutions can be limited (-1 mean unlimited).

e.g.:

```yaml
string: (( replace("foobar", "o", "u") ))
```

yields `fuubar`.

If a regular expression should be used as search string the function 
`replace_match` can be used. Here the search string is evaluated as [regular
expression](https://github.com/google/re2/wiki/Syntax). It may conatain sub expressions.
These matches can be used in the [replacement string](https://golang.org/pkg/regexp/#Regexp.Expand)

e.g.:

```yaml
string: (( replace_match("foobar", "(o*)b", "b${1}") ))
```

yields `fbooar`.

The replacement argument might also be a lambda function. In this case, for
every match the function is called to determine the replacement value.
The single input argument is a list of actual sub expression matches.

e.g.:

```yaml
string: (( replace_match("foobar-barfoo", "(o*)b", |m|->upper(m.[1]) "b" ) ))
```

yields `fOObar-barfoo`.

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

Returns the match of a [regular expression](https://github.com/google/re2/wiki/Syntax)
for a given string value. The match is a list of the matched values for the sub
expressions contained in the regular expression. Index 0 refers to the match of
the complete regular expression. If the string value does not match an empty
list is returned.

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

A third argument of type integer may be given to request a multi match of a
maximum of *n* repetitions. If the value is negative all repetions are reported.
The result is a list of all matches, each in the format described above.

### `(( keys(map) ))`

Determine the sorted list of keys used in a map.

e.g.:

```yaml
map:
  alice: 25
  bob: 25
keys: (( keys(map) ))
```

yields:

```yaml
map:
  alice: 25
  bob: 25
keys:
  - alice
  - bob
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

An optional second argument can be used to specify the maximum line length.
In this case the result will be multi-line string.

### `(( hash(string) ))`

The function `hash` generates several kinds of hashes for the given string.
By default as `sha256` hash is generated. An optional second argument specifies
the hash type. Possible types are `md4`, `md5`, `sha1`, `sha224`, `sha256`, 
`sha384`, `sha2512`, `sha512/224`or `sha512/256`.

`md5`hashes can still be generated by the deprecated finctio `md5(string)`.

e.g.:

```yaml
data: alice

hash:
  deprecated: (( md5(data) ))
  md4: (( hash(data,"md4") ))
  md5: (( hash(data,"md5") ))
  sha1: (( hash(data,"sha1") ))
  sha224: (( hash(data,"sha224") ))
  sha256: (( hash(data,"sha256") ))
  sha384: (( hash(data,"sha384") ))
  sha512: (( hash(data,"sha512") ))
  sha512_224: (( hash(data,"sha512/224") ))
  sha512_256: (( hash(data,"sha512/256") ))
```

evaluates to

```yaml
data: alice
hash:
  deprecated: 6384e2b2184bcbf58eccf10ca7a6563c
  md4: 616c69636531d6cfe0d16ae931b73c59d7e0c089c0
  md5: 6384e2b2184bcbf58eccf10ca7a6563c
  sha1: 522b276a356bdf39013dfabea2cd43e141ecc9e8
  sha224: 38b7e5d5651aaf85694a7a7c6d5db1275af86a6df93a36b8a4a2e771
  sha256: 2bd806c97f0e00af1a1fc3328fa763a9269723c8db8fac4f93af71db186d6e90
  sha384: 96a5353e625adc003a01bdcd9b21b21189bdd9806851829f45b81d3dfc6721ee21f6e0e98c4dd63bc559f66c7a74233a
  sha512: 408b27d3097eea5a46bf2ab6433a7234a33d5e49957b13ec7acc2ca08e1a13c75272c90c8d3385d47ede5420a7a9623aad817d9f8a70bd100a0acea7400daa59
  sha512_224: c3b8cfaa37ae15922adf3d21606e3a9836ba2a9d7838b040b7c96fd7
  sha512_256: ad0a339b08dc090fe3b16eae376f7e162836e8728da9c45466842e19508d7627
```

### `(( bcrypt("password", 10) ))`

The function `bcrypt` generates a bcrypt password hash for the given string
using the specified cost factor (defaulted to 10, if missing).

e.g.:

```yaml
hash: (( bcrypt("password", 10) ))
```

evaluates to

```yaml
hash: $2a$10$b9RKb8NLuHB.tM9haPD3N.qrCsWrZy8iaCD4/.cCFFCRmWO4h.koe
```

### `(( bcrypt_check("password", hash) ))`

The function `bcrypt_check` validates a password against a given bcrypt hash.

e.g.:

```yaml
hash: $2a$10$b9RKb8NLuHB.tM9haPD3N.qrCsWrZy8iaCD4/.cCFFCRmWO4h.koe
valid: (( bcrypt_check("password", hash) ))
```

evaluates to

```yaml
hash: $2a$10$b9RKb8NLuHB.tM9haPD3N.qrCsWrZy8iaCD4/.cCFFCRmWO4h.koe
valid: true
```

### `(( md5crypt("password") ))`

The function `md5crypt` generates an Apache MD5 encrypted password hash for the
given string.

e.g.:

```yaml
hash: (( md5crypt("password") ))
```

evaluates to

```yaml
hash: $apr1$3Qc1aanY$16Sb5h7U1QrcqwZbDJIYZ0
```

### `(( md5crypt_check("password", hash) ))`

The function `md5crypt_check` validates a password against a given Apache MD5 encrypted hash.

e.g.:

```yaml
hash: $2a$10$b9RKb8NLuHB.tM9haPD3N.qrCsWrZy8iaCD4/.cCFFCRmWO4h.koe
valid: (( bcrypt_check("password", hash) ))
```

evaluates to

```yaml
hash: $apr1$B77VuUUZ$NkNFhkvXHW8wERSRoi74O1
valid: true
```

### `(( decrypt("secret") ))`

This function can be used to store encrypted secrets in a spiff yaml file.
The processed result will then contain the decrypted value.
All node types can be encrypted and decrypted, including complete maps and lists.

The password for the decryption can either be given as second argument, or
(the preferred way) it can be specified by the environment variable
`SPIFF_ENCRYPTION_KEY`. 

An optional last argument may select the encryption method. The only method
supported so far is `3DES`. Other methods may be added for dedicated
spiff versions by using the encryption method registration offered by the spiff
library.

A value can be encrypted by using the `encrypt("secret")` function.

e.g.:

```yaml
password: this a very secret secret and may never be exposed to unauthorized people
encrypted: (( encrypt("spiff is a cool tool", password) ))
decrypted: (( decrypt(encrypted, password) ))
```

evaluated to something like

```yaml
decrypted: spiff is a cool tool
encrypted: d889f9e4cc7ae13effcbc8bb8cd0c38d1fb2197738444f753c48796d7946083e6639e5a1bf8f77648f2a1ddf37023c65ff57d52d0519d1d92cbcf87d3e263cba
password: this a very secret secret and may never be exposed to unauthorized people
```

### `(( rand("[:alnum:]", 10) ))`

The function `rand` generates random values. The first argument 
decides what kind of values are requested. With no argument it generates
a positive random number in the `int64` range.

| argument type | result |
| ------------- | ------ |
| int | integer value in the range [0,_n_) for positive _n_ and (_n_,0] for negative _n_ |
| bool | boolean value |
| string | one rune string, where the rune is in the given character range, any combination of character classes or character ranges usable for [regexp](https://github.com/google/re2/wiki/Syntax) can be used. If an additional length argument is specified the resulting string will have the given length.

e.g.:

```yaml
int:   (( rand() ))
int10: (( rand(10) ))
neg10:   (( rand(-10) ))
bool: (( rand(true) ))
string: (( rand("[:alpha:][:digit:]-", 10) ))
upper: (( rand("A-Z", 10) ))
punct: (( rand("[:punct:]", 10) ))
alnum: (( rand("[:alnum:]", 10) ))
```

evaluates to

```yaml
int: 8037669378456096839
int10: 7
neg10: -5
bool: true
string: ghhjAYPMlD
upper: LBZQFRSURL
alnum: 0p6KS7EhAj
punct: '&{;,^])"(#'
```

### `(( type(foobar) ))`

The function `type` yields a string denoting the type of the given expression.

e.g.:

```yaml
template:
  <<: (( &template ))
  
types:
  - int: (( type(1) ))
  - bool: (( type(true) ))
  - string: (( type("foobar") ))
  - list:   (( type([]) ))
  - map:    (( type({}) ))
  - lambda: (( type(|x|->x) ))
  - template: (( type(.template) ))
  - nil: (( type(~) ))
  - undef: (( type(~~) ))
```

evaluates types to

```yaml
types:
- int: int
- bool: bool
- string: string
- list: list
- map: map
- lambda: lambda
- template: template
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

The function `stub` yields the value of a dedicated field found in the first
upstream stub defining it.

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

The argument passed to this function must either be a reference literal or
an expression evaluating to either a string denoting a reference or a string
list denoting the list of path elements for the reference.
If no argument or an undefined (`~~`) is given, the actual field path is used.

Please note, that a given sole reference will not be evaluated as expression,
if its value should be used, it must be transformed to an expression, for example 
by denoting `(ref)` or `[] ref` for a list expression.
  
Alternatively the `merge` operation could be used, for example `merge foo.bar`. The difference is that `stub` does not merge, therefore the field will still be merged (with the original path in the document).

### `(( eval(foo "." bar ) ))`

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

### `(( env("HOME" ) ))`

Read the value of an environment variable whose name is given as dynaml expression. If the environment variable is not set the evaluation fails.

In a second flavor the function `env` accepts multiple arguments and/or list arguments, which are joined to a single list. Every entry in this list is used as name of an environment variable and the result of the function is a map of the given given variables as yaml element. Hereby non-existent environment variables are omitted.

### `(( parse(yamlorjson) ))`

Parse a yaml or json string and return the content as yaml value. It can therefore be used for
further dynaml evaluation.

e.g.:

```yaml

json: |
   { "alice": 25 }
result: (( parse( json ).alice ))
```

yields the value `25` for the field `result`.

The function `parse` supports an optional second argument, the _parse mode_.
Here the same modes are possible as for the [read function](#-readfileyml-).
The default parsing mode is `import`, the content is just parsed and there is
no further evaluation during this step.

### `(( asjson(expr) ))`

This function transforms a yaml value given by its argument to a _json_ string.
The corresponding function `asyaml` yields the yaml value as _yaml document_ string.

e.g.:

```yaml
data:
  alice: 25

mapped:
  json: (( asjson(.data) ))
  yaml: (( asyaml(.data) ))
```

resolves to

```yaml
data:
  alice: 25

mapped:
  json: '{"alice":25}'
  yaml: |+
    alice: 25
```

### `(( catch(expr) ))`

This function executes an expression and yields some evaluation info map.
It always succeeds, even if the expression fails. The map includes the 
following fields:

| name  | type   | meaning |
| ----- | ------ | ------- |
| `valid` | bool   | expression is valid |
| `error` | string | the error message text of the evaluation |
| `value` | any    | the value of the expression, if evaluation was successful |


e.g.:

```yaml
data:
  fail: (( catch(1 / 0) ))
  valid: (( catch( 5 * 5) ))
```

resolves to 

```yaml
data:
  fail:
    error: division by zero
    valid: false
  valid:
    error: ""
    valid: true
    value: 25
```

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

A map might also be given by a [map expression](#--alice--25--). Here it is possible to specify
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

Instead of multiple arguments a single list argument can be given. The list
must contain the maps to be merged.

Nested merges have access to all outer bindings. Relative references are first
searched in the actual document. If they are not found there all outer bindings
are used to lookup the reference, from inner to outer bindings. Additionally the
[context (`__ctx`)](#access-to-evaluation-context) offers a field `OUTER`,
which is a list of all outer documents of the nested merges, which can be used
to lookup absolute references.

e.g.:

```yaml
data:
  alice:
    age: 24

template:
  <<: (( &template ))
  bob:  25
  outer1: (( __ctx.OUTER.[0].data )) # absolute access to outer context
  outer2: (( data.alice.age ))       # relative access to outer binding
  sum: (( .bob + .outer2 ))

merged: (( merge(template) ))
```

resolves `merged` to

```yaml
merged:
  bob: 25
  outer1:
    alice:
      age: 24
  outer2: 24
  sum: 49
```

### `(( intersect(list1, list2) ))`

The function `intersect` intersects multiple lists. A list may contain entries
of any type.

e.g.:

```yaml
list1:
- - a
- - b
- a
- b
- { a: b }
- { b: c }
- 0
- 1
- "0"
- "1"
list2:
- - a
- - c
- a
- c
- { a: b }
- { b: b }
- 0
- 2
- "0"
- "2"
intersect: (( intersect(list1, list2) ))
```

resolves `intersect` to

```yaml
intersect:
- - a
- a
- { a: b }
- 0
- "0"
```

### `(( reverse(list) ))`

The function `reverse` reverses the order of a list. The list may contain entries
of any type.

e.g.:

```yaml
list:
- - a
- b
- { a: b }
- { b: c }
- 0
- 1
reverse: (( reverse(list) ))
```

resolves `reverse` to

```yaml
reverse:
- 1
- 0
- { b: c }
- { a: b }
- b
- - a
```

### `(( validate(value,"dnsdomain") ))`

The function `validate` validates an expression using a set of validators.
The first argument is the value to validate and all other arguments are
validators that must succeed to accept the value. If at least one validator
fails an appropriate error message is generated that explains the fail reason.

A validator is denoted by a string or a list containing the validator type
as string and its arguments. A validator can be negated with a preceeding
`!` in its name.

The following validators are available:

| Type | Arguments | Meaning |
| ---- | --------- | ------- |
| `empty` | none | empty list, map or string |
| `dnsdomain` | none | dns domain name |
| `wildcarddnsdomain` | none | wildcard dns domain name |
| `dnslabel` | none | dns label |
| `dnsname` | none | dns domain or wildcard domain |
| `ip` | none | ip address |
| `cidr` | none | cidr | 
| `publickey` | none | public key in pem format |
| `privatekey` | none | private key in pem format |
| `certificate` | none | certificate in pem format |
| `ca`|  none | certificate for CA |
| `type`| list of accepted type keys | at least one [type key](#-typefoobar-) must match |
| `valueset` | list argument with values | possible values |
| `value` or `=` | value | check dedicated value |
| `gt` or `>` | value | greater than (number/string) |
| `lt` or `<` | value | less than (number/string) |
| `ge` or `>=` | value | greater or equal to (number/string) |
| `le` or `<=` | value | less or equal to (number/string) |
| `match` or `~=` | regular expression | string value matching regular expression |
| `list` | optional list of entry validators | is list and entries match given validators |
| `map` | [[ &lt;key validator&gt;, ] &lt;entry validator&gt; ] | is map and keys and entries match given validators |
| `mapfield` | &lt;field name&gt; [ , &lt;validator&gt;] | required entry in map |
| `optionalfield` | &lt;field name&gt; [ , &lt;validator&gt;] | optional entry in map |
| `and` | list of validators | all validators must succeed |
| `or` | list of validators | at least one validator must succeed |
| `not` or `!` | validator | negate the validator argument(s) |

If the validation succeeds the value is returned.

e.g.:

```yaml
dnstarget: (( validate("192.168.42.42", [ "or", "ip", "dnsdomain" ]) ))
```

evaluates to

```yaml
dnstarget: 192.168.42.42
```

If the validation fails an error explaining the failure reason is generated.

e.g.:

```yaml
dnstarget: (( validate("alice+bob", [ "or", "ip", "dnsdomain" ]) ))
```

yields the following error:

```
*condition 1 failed: (is no ip address: alice+bob and is no dns domain: [a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')]) 
```

A validator might also be a lambda expression taking at least one argument and returning
a boolean value. This way it is possible to provide own validators as part
of the yaml document.

e.g.:

```yaml
val: (( validate( 0, |x|-> x > 1 ) ))
```

If more than one parameter is declared the additional arguments
must be specified as validator arguments. The first argument is always
the value to check.

e.g.:

```yaml
val: (( validate( 0, [|x,m|-> x > m, 5] ) ))
```

The lambda function may return a list with 1, 2 or 3 elements, also.
This can be used to provide appropriate messages.

| Index | Meaning |
| ----- | ------- | 
| 0 | the first index always is the match result, it must be evaluatable as boolean |
| 1 | if two elements are given, the second index is the message describing the actual result |
| 2 | here index 1 decribes the success message and 2 the failure message |

e.g.:

```yaml
val: (( validate( 6, [|x,m|-> [x > m, "is larger than " m, "is less than or equal to " m], 5] ) ))
```

Just to mention, the validator specification might be given inline as shown
in the examples above, but as reference expressions, also. The `not`, `and` 
and `or` validators accept deeply nested validator specifications.

e.g.:

```yaml
dnsrecords:
   domain: 1.2.3.4
validator:
  - map
  - - or                              # key validator
    - dnsdomain
    - wildcarddnsdomain
  - ip                                # entry validator

val: (( validate( map, validator)  ))
```

### `(( check(value,"dnsdomain") ))`

The function `check` can be used to match a yaml structure against a yaml
based value checker. Hereby the same check description already described for 
[validate](#-validatevaluednsdomain-) can be used. The result of the call is
a boolean value indicating the match result. It does not fail if the check
fails.
 
### `(( error("message") ))`

The function `error` can be used to cause explicit evaluation failures with
a dedicated message.

This can be used, for example, to reduce a complex
processing error to a meaningful message by appending the error function
as default for the potentially failing comples expression.

e.g.:

```yaml
value: (( <some complex potentially failing expression> || error("this was an error my friend") ))
```

Another scenario could be omitting a descriptive message for missing required
fields by using an error expression as (default) value for a field intended to
be defined in an upstream stub.

### Math

*dynaml* support various math functions:

returning integers: `ceil`, `floor`, `round` and `roundtoeven`

returning floats or integers: `abs`

returning floats: `sin`,`cos`, `sinh`, `cosh`, `asin`, `acos`, `asinh`,`acosh`,
           `sqrt`, `exp`, `log`, `log10`,

### Conversions

*dynaml* supports various type conversions between `integer`, `float`, `bool`
and `string` values by appropriate functions.

e.g.:


```yaml
value: (( integer("5") ))
```

converts a string to an integer value.

Converting an integer to a string accepts an optional additional integer
argument for specifying the base for conversion, for example `string(55,2)`
will result in `"110111"`. The default base is 10. The base must be between
2 and 36.


### Accessing External Content

_Spiff_ supports access to content outside of the template and sub files. It is
possible to read files, execute commands and pipelines. All those functions exist
in two flavors.
- A cached flavor executes the operation ones and caches the result
  for subsequent identical operations. This speeds up the processing, especially
  for command executions.
- If the result evolves over time, it might be useful to always get the latest 
  content. This is the case if the [`sync`](#-syncexpr-condition.value-10-) 
  function is used, which is intended to synchronize the template processing
  with a dedicate state (provided by external content). Here the caching 
  operations would not be useful, therefore there is a second uncached flavor.
  Every function is available with the suffix `_uncached` (for example 
  `read_uncached()`)

#### `(( read("file.yml") ))`

Read a file and return its content. There is support for three content types:
`yaml` files,`text` files and `binary` files. Reading in binary mode will
result in a base64 encoded multi-line string.

If the file suffix is `.yml`, `.yaml` or `.json`,
by default the yaml type is used. If the file should be read as `text`, this
type must be explicitly specified.
In all other cases the default is `text`, therefore reading a binary file
(for example an archive) urgently requires specifying the `binary` mode.
  
An optional second parameter can be used to explicitly specifiy the desired
return type: `yaml` or `text`. For _yaml_ documents some addtional 
types are supported: `multiyaml`, `template`, `templates`, `import` and
`importmulti`.

##### yaml documents

A yaml document will be parsed and the tree is returned. The  elements of the
tree can be accessed by regular dynaml expressions.

Additionally the yaml file may again contain dynaml expressions. All included
dynaml expressions will be evaluated in the context of the reading expression.
This means that the same file included at different places in a yaml document
may result in different sub trees, depending on the used dynaml expressions.

If is poassible to read a multi-document yaml, also. If the type `multiyaml`
is given, a list node with the yaml document root nodes is returned.

The yaml or json document can also read as _template_ by specifying the type
`template`. Here the result will be a template value, that can be used like
regular inline templates. If `templates` is specified, a multi-document is
mapped to a list of templates.

If the read type is set to `import`, the file content is read as yaml document
and the root node is used to substitute the expression. Potential dynaml
expressions contained in the document will not be evaluated with the actual
binding of the expression together with the read call,
but as it would have been part of the original file.
Therefore this mode can only be used, if there is no further processing
of the read result or the delivered values are unprocessed.

This can be used together with a chained reference
 (for examle `(( read(...).selection ))`) to delect a dedicated fragment of
the imported document. Then, the evaluatio will be done for the selected
portion, only. Expressions and references in the other parts are not
evalauted and at all and cannot lead to error.

e.g.: 

**template.yaml**

```yaml
ages:
  alice: 25

data: (( read("import.yaml", "import").first ))
``` 

**import.yaml**

```yaml
first:
  age: (( ages.alice ))

second:
  age: (( ages.bob ))
```

will not fail, because the `second` section is never evaluated.

This mode should be taken with caution, because it often leads to unexpected
results.

The read type `importmulti` can be used to import multi document yaml files as a 
list of nodes.

##### text documents

A text document will be returned as single string.

##### binary documents

It is possible to read binary documents, also. The content cannot be used
as a string (or yaml document), directly. Therefore the read mode `binary` has
to be specified. The content is returned as a base64 encoded multi-line string
value.

#### `(( exec("command", arg1, arg2) ))`

Execute a command. Arguments can be any dynaml expressions including reference expressions evaluated to lists or maps. Lists or maps are passed as single arguments containing a yaml document with the given fragment.

The result is determined by parsing the standard output of the command. It might be a yaml document or a single multi-line string or integer value. A yaml document should start with the document prefix `---`. If the command fails the expression is handled as undefined.

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

#### `(( pipe(data, "command", arg1, arg2) ))`

Execute a command and feed its standard input with dedicated data. 
The command argument must be a string. Arguments
for the command can be any dynaml expressions including reference expressions
evaluated to lists or maps. Lists or maps are passed as single arguments
containing a yaml document with the given fragment.

The input stream is generated from the given data. If this is a simple type its
string representation is used. Otherwise a yaml document is generated from the
input data. The result is determined by parsing the standard output of the
command. It might be a yaml document or a single multi-line string or integer
value. A yaml document should start with the document prefix `---`. If
the command fails the expression is handled as undefined.

e.g.

```yaml
data:
  - a
  - b
list: (( pipe( data, "tr", "a", "z") ))
```

yields

```yaml
arg:
- a
- b
list:
- z
- b
```

Alternatively `pipe` can be called with data and a list argument completely describing the command line.

The same command will be executed once, only, even if it is used in multiple expressions.

#### `(( write("file.yml", data) ))`

Write a file and return its content. If the result can be parsed as yaml document,
the document is returned. An optional 3rd argument can be used to pass the
write options.
The option arguments might be an integer denoting file permissions (default is `0644`)
or a comma separated string with options. Supported options are
- `binary`: data is base64 decoded before writing
- _integer_ string: file permissions, a leading `0` is indicating an octal value.

#### `(( tempfile("file.yml", data) ))`

Write a a temporary file and return its path name. An optional 3rd argument can
be used to pass write options. It basically behavies
like [`write`](#-writefileyml-data-) 

_Attention_: A temporary file only exists during the merge processing. It will
be deleted afterwards. 

It can be used, for example, to provide a temporary file argument for the
[`exec`](#-execcommand-arg1-arg2-) function.

#### `(( lookup_file("file.yml", list) ))`

Lookup a file is a list of directories. The result is a list of existing
files. With `lookup_dir` it is possible to lookup a directory, instead.

If no existing files can be found the empty list is returned.

It is possible to pass multiple list or string arguments to compose the
search path.

#### `(( mkdir("dir", 0755) ))`

Create a directory and all its intermediate directories if they do not
exist yet.

The permission part is optional (default 0755). The path of the directory
might be given by atring like value or as a list of path components.

#### `(( list_files(".") ))`

List files in a directory. The result is a list of existing
files. With `list_dirs` it is possible to list directories, instead.

#### `(( archive(files, "tar") ))`

Create an archive of the given type (default is `tar`) containing the listed
files. The result is the base64 encoded archive.

Supported archive types are `tar` and `targz`.

`files` might be a list or map of file entries. In case of a map, the map key
is used as default for the file path. A file entry is a map with the 
following fields:

| field | type | meaning |
|-------|------|---------|
| `path`| string | optional for maps, the file path in the archive, defaulted by the map key |
| `mode` | int or int string | file mode or write options. It basically behavies like the option argument for [`write`](#-writefileyml-data-). |
| `data` | any | file content, yaml will be marshalled as yaml document. If `mode` indicates binary mode, a string value will be base64 decoded. |
| `base64` | string | base64 encoded binary data |

e.g.:

```yaml
yaml:
  alice: 26
  bob: 27

files:
  "data/a/test.yaml":
    data: (( yaml ))
  "data/b/README.md":
    data: |+
      ### Test Docu

      **Note**: This is a test

archive: (( archive(files,"targz") ))

content: (( split("\n", exec_uncached("tar", "-tvf", tempfile(archive,"binary"))) ))
```

yields:

```yaml
archive: |-
  H4sIAAAAAAAA/+zVsQqDMBAG4Mx5igO3gHqJSQS3go7tUHyBqIEKitDEoW9f
  dLRDh6KlJd/yb8ll+HOd8SY1qbfOJw8zDmQHiIhayjURcZuIQhOeSZlphVwL
  glwsAXvM8mJ23twJ4qfnbB/3I+I4pmboW1uA0LSZmgJETr89VXCUtf9Neq1O
  5blKxm6PO972X/FN/7nKVej/EaIogto6D+XUzpQydpm8ZayA+tY76B0YWHYD
  DV9CEATBf3kGAAD//5NlAmIADAAA
content:
- -rw-r--r-- 0/0              22 2019-03-18 09:01 data/a/test.yaml
- -rw-r--r-- 0/0              41 2019-03-18 09:01 data/b/README.md

files:
  data/a/test.yaml:
    data:
      alice: 26
      bob: 27
  data/b/README.md:
    data: |+
      ### Test Docu

      **Note**: This is a test

yaml:
  alice: 26
  bob: 27

```

### X509 Functions

spiff supports some useful functions to work with _X509_ certificates and keys.
Please refer also to the [Useful to Know](#useful-to-know) section to find some
tips for providing state.

#### `(( x509genkey(spec) ))`

This function can be used generate private RSA or ECDSA keys. The result will
be a PEM encoded key as multi line string value. If a key size (integer or string)
is given as argument, an RSA key will be generated with the given key size
(for example 2048). Given one of the string values

- "P224"
- "P256"
- "P384"
- "P521"

the function will generate an appropriate ECDSA key.

e.g.:

```yaml
keys:
  key: (( x509genkey(2048) ))
```

resolves to something like

```yaml
key: |+
    -----BEGIN RSA PRIVATE KEY-----
    MIIEpAIBAAKCAQEAwxdZDfzxqz4hlRwTL060pm1J12mkJlXF0VnqpQjpnRTq0rns
    CxMxvSfb4crmWg6BRaI1cEN/zmNcT2sO+RZ4jIOZ2Vi8ujqcbzxqyoBQuMNwdb32
    ...
    oqMC9QKBgQDEVP7FDuJEnCpzqddiXTC+8NsC+1+2/fk+ypj2qXMxcNiNG1Az95YE
    gRXbnghNU7RUajILoimAHPItqeeskd69oB77gig4bWwrzkijFXv0dOjDhQlmKY6c
    pNWsImF7CNhjTP7L27LKk49a+IGutyYLnXmrlarcNYeCQBin1meydA==
    -----END RSA PRIVATE KEY-----
```

#### `(( x509publickey(key) ))`

For a given key or certificate in PEM format (for example generated with the [x509genkey](#-x509genkeyspec-)
function) this function extracts the public key and returns it again in PEM format as a
multi-line string.

e.g.:

```yaml
keys:
  key: (( x509genkey(2048) ))
  public: (( x509publickey(key)
```

resolves to something like

```yaml
key: |+
    -----BEGIN RSA PRIVATE KEY-----
    MIIEpAIBAAKCAQEAwxdZDfzxqz4hlRwTL060pm1J12mkJlXF0VnqpQjpnRTq0rns
    CxMxvSfb4crmWg6BRaI1cEN/zmNcT2sO+RZ4jIOZ2Vi8ujqcbzxqyoBQuMNwdb32
    ...
    oqMC9QKBgQDEVP7FDuJEnCpzqddiXTC+8NsC+1+2/fk+ypj2qXMxcNiNG1Az95YE
    gRXbnghNU7RUajILoimAHPItqeeskd69oB77gig4bWwrzkijFXv0dOjDhQlmKY6c
    pNWsImF7CNhjTP7L27LKk49a+IGutyYLnXmrlarcNYeCQBin1meydA==
    -----END RSA PRIVATE KEY-----
public: |+
    -----BEGIN RSA PUBLIC KEY-----
    MIIBCgKCAQEAwxdZDfzxqz4hlRwTL060pm1J12mkJlXF0VnqpQjpnRTq0rnsCxMx
    vSfb4crmWg6BRaI1cEN/zmNcT2sO+RZ4jIOZ2Vi8ujqcbzxqyoBQuMNwdb325Bf/
   ...
    VzYqyeQyvvRbNe73BXc5temCaQayzsbghkoWK+Wrc33yLsvpeVQBcB93Xhus+Lt1
    1lxsoIrQf/HBsiu/5Q3M8L6klxeAUcDbYwIDAQAB
    -----END RSA PUBLIC KEY-----
```

To generate an ssh public key an optional additional format argument can be set
to `ssh`. The result will then be a regular public key format usable for ssh.
The default format is `pem` providing the pem output format shown above.

RSA keys are by default marshalled in PKCS#1 format(`RSA PUBLIC KEY`) in pem.
If the generic *PKIX* format (`PUBLIC KEY`) is required the format
argument `pkix` must be given.

Using the format `ssh` this function can also be used to convert a pem formatted
public key into an ssh key, 

#### `(( x509cert(spec) ))`

The function `x509cert` creates locally signed certificates, either a self signed
one or a certificate signed by a given ca. It returns a PEM encoded certificate
as a multi-line string value.

The single _spec_ parameter take a map with some optional and non optional
fields used to specify the certificate information. It can be an
[inline map expression](#--alice--25--) or any map reference into the rest of
the yaml document.

The following map fields are observed:

| Field Name  | Type | Required | Meaning |
| ------------| ---- | -------- | ------- |
| `commonName` | string | optional |  Common Name field of the subject |
| `organization` | string or string list | optional |  Organization field of the subject |
| `country` | string or string list | optional |  Country field of the subject |
| `isCA` | bool | optional |  CA option of certificate |
| `usage` | string or string list | required |  usage keys for the certificate (see below) |
| `validity` | integer | optional |  validity interval in hours |
| `validFrom` | string | optional |  start time in the format "Jan 1 01:22:31 2019" |
| `hosts` | string or string list | optional |  List of DNS names or IP addresses |
| `privateKey` | string | required or publicKey |  private key to generate the certificate for |
| `publicKey` | string | required or privateKey|  public key to generate the certificate for |
| `caCert` | string | optional|  certificate to sign with |
| `caPrivateKey` | string | optional|  priavte key for `caCert` |

For self-signed certificates, the `privateKey`field must be set. `publicKey`
and the `ca` fields should be omitted. If the `caCert`field is given, the `caKey`
field is required, also. If the `privateKey`field is given together with the
`caCert`, the public key for the certificate is extracted from the private key.

Additional fields are silently ignored.

The following usage keys are supported (case is ignored):

| Key |  Meaning |
| ------------| ---- |
| `Signature` | x509.KeyUsageDigitalSignature |
| `Commitment` | x509.KeyUsageContentCommitment |
| `KeyEncipherment` | x509.KeyUsageKeyEncipherment |
| `DataEncipherment` | x509.KeyUsageDataEncipherment |
| `KeyAgreement` | x509.KeyUsageKeyAgreement |
| `CertSign` | x509.KeyUsageCertSign |
| `CRLSign` | x509.KeyUsageCRLSign |
| `EncipherOnly` | x509.KeyUsageEncipherOnly |
| `DecipherOnly` | x509.KeyUsageDecipherOnly |
| `Any` | x509.ExtKeyUsageAny |
| `ServerAuth` | x509.ExtKeyUsageServerAuth |
| `ClientAuth` | x509.ExtKeyUsageClientAuth |
| `codesigning` | x509.ExtKeyUsageCodeSigning |
| `EmailProtection` | x509.ExtKeyUsageEmailProtection |
| `IPSecEndSystem` | x509.ExtKeyUsageIPSECEndSystem |
| `IPSecTunnel` | x509.ExtKeyUsageIPSECTunnel |
| `IPSecUser` | x509.ExtKeyUsageIPSECUser |
| `TimeStamping` | x509.ExtKeyUsageTimeStamping |
| `OCSPSigning` | x509.ExtKeyUsageOCSPSigning |
| `MicrosoftServerGatedCrypto` | x509.ExtKeyUsageMicrosoftServerGatedCrypto |
| `NetscapeServerGatedCrypto` | x509.ExtKeyUsageNetscapeServerGatedCrypto |
| `MicrosoftCommercialCodeSigning` | x509.ExtKeyUsageMicrosoftCommercialCodeSigning |
| `MicrosoftKernelCodeSigning` | x509.ExtKeyUsageMicrosoftKernelCodeSigning |


e.g.:

```yaml
spec:
  <<: (( &local ))
  ca:
    organization: Mandelsoft
    commonName: Uwe Krueger
    privateKey: (( data.cakey ))
    isCA: true
    usage:
      - Signature
      - KeyEncipherment

data:
  cakey: (( x509genkey(2048) ))
  cacert: (( x509cert(spec.ca) ))
```

generates a self-signed root certificate and resolves to something like

```yaml
cakey: |+
    -----BEGIN RSA PRIVATE KEY-----
    MIIEpAIBAAKCAQEAwxdZDfzxqz4hlRwTL060pm1J12mkJlXF0VnqpQjpnRTq0rns
    CxMxvSfb4crmWg6BRaI1cEN/zmNcT2sO+RZ4jIOZ2Vi8ujqcbzxqyoBQuMNwdb32
    ...
    oqMC9QKBgQDEVP7FDuJEnCpzqddiXTC+8NsC+1+2/fk+ypj2qXMxcNiNG1Az95YE
    gRXbnghNU7RUajILoimAHPItqeeskd69oB77gig4bWwrzkijFXv0dOjDhQlmKY6c
    pNWsImF7CNhjTP7L27LKk49a+IGutyYLnXmrlarcNYeCQBin1meydA==
    -----END RSA PRIVATE KEY-----
cacert: |+
    -----BEGIN CERTIFICATE-----
    MIIDCjCCAfKgAwIBAgIQb5ex4iGfyCcOa1RvnKSkMDANBgkqhkiG9w0BAQsFADAk
    MQ8wDQYDVQQKEwZTQVAgU0UxETAPBgNVBAMTCGdhcmRlbmVyMB4XDTE4MTIzMTE0
    ...
    pOUBE3Tgim5rnpa9K9RJ/m8IVqlupcONlxQmP3cCXm/lBEREjODPRNhU11DJwDdJ
    5fd+t5SMEit2BvtTNFXLAwz48EKTxsDPdnHgiQKcbIV8NmgUNPHwXaqRMBLqssKl
    Cyvds9xGtAtmZRvYNI0=
    -----END CERTIFICATE-----
```
#### `(( x509parsecert(cert) ))`

This function parses a certificate given in PEM format and returns a map
of fields:

| Field Name  | Type | Required | Meaning |
| ------------| ---- | -------- | ------- |
| `commonName` | string | optional |  Common Name field of the subject |
| `organization` | string list | optional |  Organization field of the subject |
| `country` | string list | optional |  Country field of the subject |
| `isCA` | bool | always |  CA option of certificate |
| `usage` | string list | always |  usage keys for the certificate (see below) |
| `validity` | integer | always |  validity interval in hours |
| `validFrom` | string | always |  start time in the format "Jan 1 01:22:31 2019" |
| `validUntil` | string | always |  start time in the format "Jan 1 01:22:31 2019" |
| `hosts` | string list | optional |  List of DNS names or IP addresses |
| `dnsNames` | string list | optional |  List of DNS names |
| `ipAddresses` | string list | optional |  List of IP addresses |
| `publicKey` | string | always|  public key to generate the certificate for |

e.g.:

```yaml
data:
  <<: (( &temporary ))
  spec:
    commonName: test
    organization: org
    validity: 100
    isCA: true
    privateKey: (( gen.key ))
    hosts:
      - localhost
      - 127.0.0.1

    usage:
     - ServerAuth
     - ClientAuth
     - CertSign

  gen:
    key: (( x509genkey() ))
    cert: (( x509cert(spec) ))

cert: (( x509parsecert(data.gen.cert) ))
```

resolves to

```yaml
cert:
  commonName: test
  dnsNames:
  - localhost
  hosts:
  - 127.0.0.1
  - localhost
  ipAddresses:
  - 127.0.0.1
  isCA: true
  organization:
  - org
  publickey: |+
    -----BEGIN RSA PUBLIC KEY-----
    MIIBCgKCAQEA+UIZQUTa/j+WlXC394bccBTltV+Ig3+zB1l4T6Vo36pMBmU4JIkJ
    ...
    TCsrEC5ey0cCeFij2FijOJ5kmm4cK8jpkkb6fLeQhFEt1qf+QqgBw3targ3LnZQf
    uE9t5MIR2X9ycCQSDNBxcuafHSwFrVuy7wIDAQAB
    -----END RSA PUBLIC KEY-----
  usage:
  - CertSign
  - ServerAuth
  - ClientAuth
  validFrom: Mar 11 15:34:36 2019
  validUntil: Mar 15 19:34:36 2019
  validity: 99  # yepp, that's right, there has already time passed since the creation
```

### Wireguard Functions

spiff supports some useful functions to work with _wireguard_ keys.
Please refer also to the [Useful to Know](#useful-to-know) section to find some
tips for providing state.

#### `(( wggenkey() ))`

This function can be used generate private wireguard key. The result will
base64 encoded.

e.g.:

```yaml
keys:
  key: (( wggenkey() ))
```

resolves to something like

```yaml
key: WH9xNVJuSuh7sDVIyUAlmxc+woFDJg4QA6tGUVBtGns=
```

#### `(( wgpublickey(key) ))`

For a given key (for example generated with the [wggenkey](#-wggenkey-)
function) this function extracts the public key and returns it again in base64 format-

e.g.:

```yaml
keys:
  key: (( wggenkey() ))
  public: (( wgpublickey(key)
```

resolves to something like

```yaml
key: WH9xNVJuSuh7sDVIyUAlmxc+woFDJg4QA6tGUVBtGns=
public: n405KfwLpfByhU9pOu0A/ENwp0njcEmmQQJvfYHHQ2M=
```

## `(( lambda |x|->x ":" port ))`

Lambda expressions can be used to define additional anonymous functions. They
can be assigned to yaml nodes as values and referenced with path expressions
to call the function with approriate arguments in other dynaml expressions.
For the final document they are mapped to string values.

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
If a complete expression is a lambda expression the keyword `lambda` can be omitted.

Lambda expressions evaluate to lambda values, that are used as final values
in yaml documents processed by _spiff_. 

**Note**: If the final document still contains lambda values, they are transferred
to a textual representation. It is not guaranteed that this representation can 
correctly be parsed again, if the document is re-processed by _spiff_. Especially
for complex scoped and curried functions this is not possible.

Therefore function nodes should always be _temporary_ or _local_ to be available
during processing or merging, but being omitted for the final document.


### Positional versus Named Arguments

A typical function call uses positional arguments. Here the given arguments
satisfy the declared function parameters in the given order.
For lambda values it is also possible to use named arguments in the call
expression. Here an argument is assigned to a dedicated parameter as declared
by the lambda expression. The order of named arguments can be arbitrarily chosen.

e.g.:

```yaml
func: (( |a,b,c|->{$a=a, $b=b, $c=c } ))
result: (( .func(c=1, b=2, a=1) ))
```

It is also posible to combine named with positional arguments. Hereby the
positional arguments must follow the named ones.

e.g.:

```yaml
func: (( |a,b,c|->{$a=a, $b=b, $c=c } ))
result: (( .func(c=1, 1, 2) ))
```

The same argument MUST NOT be satified by both, a named and a positional 
argument.

Instead of using the parameter name it is also possible to use the parameter
index, instead.

e.g.:

```yaml
func: (( |a,b,c|->{$a=a, $b=b, $c=c } ))
result: (( .func(3=1, 1) ))
```

As such, this feature seems to be quite useless, but it shows its power if
combined with [optional parameters](#optional-parameters) or 
[currying](#currying) as shown in the next paragraphs.

### Scopes and Lambda Expressions

A lambda expression might refer to absolute or relative nodes of the actual yaml document of the call. Relative references are evaluated in the context of the function call. Therefore

```yaml
lvalue: (( lambda |x,y|->x + y + offset ))
offset: 0
values:
  offset: 3
  value: (( .lvalue(1,2) ))
```

yields `6` for `values.value`.


Besides the specified parameters, there is an implicit name (`_`), that can be
used to refer to the function itself. It can be used to define self recursive
function. Together with the logical and conditional operators a fibunacci
function can be defined:

```yaml
fibonacci: (( lambda |x|-> x <= 0 ? 0 :x == 1 ? 1 :_(x - 2) + _( x - 1 ) ))
value: (( .fibonacci(5) ))
```

yields the value `8` for the `value` property.

By default reference expressions in a lambda expression are evaluated in the
static scope of the lambda dedinition followed by the static yaml scope of the
caller. Absolute references are always evalated in the document scope of the
caller.

The name `_` can also be used as an anchor to refer to the static definition
scope of the lambda expression in the yaml document that was used to define
the lambda function. Those references are always interpreted as relative
references related to the this static yaml document scope. There is no
denotation for accessing the root element of this definition scope.

Relative names can be used to access the static 
definition scope given inside the dynaml expression (outer scope literals and
parameters of outer lambda parameters)

e.g.:

````yaml
env:
  func: (( |x|->[ x, scope, _, _.scope ] ))
  scope: definition

call:
   result: (( env.func("arg") ))
   scope: call
````

yields the `result` list:

```yaml
call:
  result:
  - arg
  - call
  - (( lambda|x|->[x, scope, _, _.scope] ))  # the lambda expression as lambda value
  - definition
```

This also works across multiple stubs. The definition context is the stub the
lambda expression is defined in, even if it is used in stubs down the chain.
Therefore it is possible to use references in the lambda expression, not visible
at the caller location, they carry the static yaml document scope of their 
definition with them.

Inner lambda expressions remember the local binding of outer lambda expressions.
This can be used to return functions based on arguments of the outer function.

e.g.:

```yaml
mult: (( lambda |x|-> lambda |y|-> x * y ))
mult2: (( .mult(2) ))
value: (( .mult2(3) ))
```

yields `6` for property `value`.

### Optional Parameters

Trailing parameters may be defaulted in the lambda expression by assigning
values in the declaration. Those parameter are then optional, it is not
required to specify arguments for those parameters in function calls. 

e.g.:

```yaml
mult: (( lambda |x,y=2|-> x * y ))
value: (( .mult(3) ))
```

yields `6` for property `value`.

It is possible to default all parameters of a lambda expression. The function
can then be called without arguments. There might be no non-defaulted parameters
after a defaulted one.

A call with positional arguments may only omit arguments for optional parameters
from right to left. If there should be an explicit argument for the right most
parameter, arguments for all parameters must be specified or
[named arguments](#positional-versus-named-arguments) must be used.
Here the desired optional parameter can explicitly be set prior to the regular
positional arguments.

e.g.:

```yaml
func:  (( |a,b=1,c=2|->{$a=a, $b=b, $c=c } ))
result: (( .func(c=3, 2) ))
```

evaluates `result` to

```yaml
result:
  a: 2
  b: 1
  c: 3
```

The expression for the default does not need to be a constant value or even
expression, it might refer to other nodes in the yaml document. The default
expression is always evaluated in the scope of the lambda expression declaration
at the time the lambda expression is evaluated.

e.g.:

**stub.yaml**
```yaml
default: 2
mult: (( lambda |x,y=default * 2|-> x * y ))
```

**template.yaml**
```yaml
mult: (( merge ))
scope:
  default: 3
  value: (( .mult(3) ))
```

evaluates `value`to 12

### Variable Argument Lists

The last parameter in the parameter list of a lambda expression may be a 
_varargs_ parameter consuming additional argument in a fnction call.
This parameter is always a list of values, one entry per additional argument.

A _varargs_ parameter is denoted by a `...` following the last parameter name.

e.g.:

```yaml
func: (( |a,b...|-> [a] b ))
result: (( .func(1,2,3) ))
```
 
yields the list `[1, 2, 3]` for property `result`.

If no argument is given for the _varargs_ parameter its value is the empty list.

The `...` operator can also be used for [inline list expansion](#inline-list-expansion).

If a vararg parameter should be set by a [named argument](#positional-versus-named-arguments)
its value must be a list.

### Currying

Using the _currying_ operator (`*(`) a lambda function may be transformed to
another function with less parameters by specifying leading argument values.

The result is a new function taking the missing arguments (currying) and using
the original function body with a static binding for the specified parameters.

e.g.:

```yaml
mult: (( lambda |x,y|-> x * y ))
mult2: (( .mult*(2) ))
value: (( .mult2(3) ))
```

Currying may be combined with [defaulted parameters](#optional-parameters).
But the resulting function does not default the leading parameters, it
is just a new function with less parameters pinning the specified ones.

If the original function uses a [variable argument list](#variable-argument-lists),
the currying may span any number of the variable argument part, but once at
least one such argument is given, the parameter for the variable part is satisfied.
It cannot be extended by a function call of the curried function.

e.g.:

```yaml
func: (( |a,b...|->join(a,b) ))
func1: (( .func*(",","a","b")))
#invalid: (( .func1("c") ))
value: (( .func1() ))
```

evaluates `value` to `"a,b"`.

It is also possible to use currying for builtin functions, like 
[`join`](#-join---list-).

e.g.:

```yaml
stringlist: (( join*(",") ))
value: (( .stringlist("a", "b")  ))
```

evaluates `value` to `"a,b"`.

There are several builtin functions acting on unevaluated or unevaluatable
arguments, like [`defined`](#-definedfoobar-). For these functions currying is
not possible.

Using positional arguments currying is only possible from right to left.
But currying can also be done for [named arguments](#positional-versus-named-arguments).
Here any parameter combination, regardless of the position in the parameter
list, can be preset. The resulting function then has the unsatisfied parameters
in their original order. Switching the parameter order is not possible.

e.g.:

```yaml
func: (( |a,b=1,c=2|->{$a=a, $b=b, $c=c } ))
curry: (( .func(c=3, 2) ))

result: (( .curry(5) ))
```

evalutes `result` to

```yaml
result:
  a: 2
  b: 5
  c: 3
```

The resulting function keeps the parameter `b`. Hereby the default value will
be kept. Therefore it can just be called without argument (`.curry()`), which 
would produce

```yaml
result:
  a: 2
  b: 1
  c: 3
```

**Attention**: 

For compatibility reasons currying is also done, if a lambda function without
defaulted parameters is called with less arguments than declared parameters.

This behaviour is **deprecated** and will be removed in the future. It is
replaced by the currying operator.

e.g.:

```yaml
mult: (( lambda |x,y|-> x * y ))
mult2: (( .mult(2) ))
value: (( .mult2(3) ))
```

evaluates `value` to 6.

## `(( catch[expr|v,e|->v] ))`

This expression evaluates an expression (`expr`) and then
executes a lambda function with the evaluation state of the expression.
It always succeeds, even if the expression fails.
The lambda function may take one or two arguments, the first
is always the evaluated value (or `nil` in case of an error).
The optional second argument gets the error message the evaluation of
the expression failed (or `nil` otherwise)

The result of the function is the result of the whole
expression. If the function fails, the complete expression fails.

e.g.:

```yaml
data:
  fail: (( catch[1 / 0|v,e|->{$value=v, $error=e}] ))
  valid: (( catch[5 * 5|v,e|->{$value=v, $error=e}] ))
```

resolves to 

```yaml
data:
  fail:
    error: division by zero
    value: null
  valid:
    error: null
    value: 25
```

## `(( sync[expr|v,e|->defined(v.field),v.field|10] ))`

If an expression `expr` may return different results for different evaluations,
it is possible to synchronize the final output with a dedicated condition
on the expression value. Such an expression could, for example, be an
uncached `read`, `exec` or `pipe` call.

The second element must evaluate to a lambda value, given by either a
regular expression or by a lambda literal as shown in the title.
It may take one or two arguments, the actual value of the value expression
and optionally an error message in case of a failing evaluation.
The result of the evaluation of the lamda expression decides whether 
the state of the evaluation of the value expression is acceptable (`true`)
or not (`false`).

If the value is accepted, an optional third expression is used to determine
the final result of the `sync[]` expression. It might be given as an expression
evaluating to a lambda value, or by a comma separated expression using the
same binding as the preceeding lambda literal.
If not given, the value of the synched expression is returned. 

If the value is not acceptable, the evaluation is repeated until a timeout
applies. The timeout in seconds is given by an optional fourth expression
(default is 5 min). Either the fourth, or the both, the
third and the fourth elements may be omitted.

The lambda values might be given as literal, or by expression, leading to the
following flavors:

- `sync[expr|v,e|->cond,value|10]`
- `sync[expr|v,e|->cond|valuelambda|10]`
- `sync[expr|v,e|->cond|v|->value|10]`
- `sync[expr|condlambda|valuelambda|10]`
- `sync[expr|condlambda|v|->value|10]`

with or without the timeout expression.

e.g.:

```yaml
data:
  alice: 25
result: (( sync[data|v|->defined(v.alice),v.alice] ))
```

resolves to 

```yaml
data:
  alice: 25
result: 25
```

This example is quite useless, because the sync expression is a constant. It
just demonstrates the usage.

## Mappings

Mappings are used to produce a new list from the entries of a _list_ or _map_,
or a new map from entries of a _map_ containing the entries processed by a
dynaml expression. The expression is
given by a [lambda function](#-lambda-x-x--port-). There are two basic forms of
the mapping function: It can be inlined as in `(( map[list|x|->x ":" port] ))`, 
or it can be determined by a regular dynaml expression evaluating to a lambda
function as in `(( map[list|mapping.expression))` (here the mapping is taken
from the property `mapping.expression`, which should hold an approriate lambda
function).

The mapping comes in two target flavors: with `[]` or `{}` in the syntax. The first
flavor always produces a _list_ from the entries of the given source. The
second one takes only a map source and produces a filtered or transformed _map_.

Additionally the mapping uses three basic mapping behaviours:
- _transforming the values using the keyword `map`_. Here the result of the lambda
  function is used as new value to replace the original one. Or
- _filtering using the keywork `select`_. Here the result of the lambda
  function is used as a boolean to decide whether the entry should be kept
  (`true`) or omitted (`false`).
- _composing_ using the keyword `sum`. Here always the list flavor is used,
  but the result type and content is completely determined by the parameterization
  of the statement by successively aggregating one entry after the other into an
  arbitrary initial value. 

**Note**: The special reference `_` is not set for inlined lambda functions as part of
the mapping syntax. Therefore the mapping statements (and all other statements using
inlined lambda functions as part of their syntax) can be used inside regular lambda
functions without hampering the meaning of this special refrence for the surrounding
explicit lambda expression.

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

### `(( map{map|elem|->dynaml-expr} ))`

Using `{}` instead of `[]` in the mapping syntax, the result is again a map
with the old keys and the new entry values. As for a list mapping additionally
a key variable can be specified in the variable list.

```yaml
persons:
  alice: 27
  bob: 26
older: (( map{persons|x|->x + 1} ) ))
```

just increments the value of all entries by one in the field `older`:

```yaml
older:
  alice: 28
  bob: 27
```

**Remark**

An alternate way to express the same is to use `sum[persons|{}|s,k,v|->s { k = v + 1 }]`.

### `(( map{list|elem|->dynaml-expr} ))`

Using `{}` instead of `[]` together with a list in the mapping syntax, the result is again a map
with the list elements as key and the mapped entry values. For this all list entries must be strings.
As for a list mapping additionally an index variable can be specified in the variable list.

```yaml
persons:
  - alice
  - bob
length: (( map{persons|x|->length(x)} ) ))
```

just creates a map mapping the list entries to their length:

```yaml
length:
  alice: 5
  bob: 3
```

### `(( select[expr|elem|->dynaml-expr] ))`

With `select` a map or list can be filtered by evaluating a boolean expression
for every entry. An entry is selected if the expression evaluates to true
equivalent value. (see [conditions](#-a--1--foo-bar-)).

Basically it offers all the mapping flavors available for `map[]`

e.g.

```yaml
list:
  - name: alice
    age: 25
  - name: bob
    age: 26


selected: (( select[list|v|->v.age > 25 ] ))
```

evaluates selected to

```yaml
selected:
- name: bob
  age: 26
```

**Remark**

An alternate way to express the same is to use `map[list|v|->v.age > 25 ? v :~]`.

### `(( select{map|elem|->dynaml-expr} ))`

Using `{}` instead of `[]` in the mapping syntax, the result is again a map
with the old keys filtered by the given expression.

```yaml
persons:
  alice: 25
  bob: 26
older: (( select{persons|x|->x > 25} ))
```

just keeps all entries with a value greater than 25 and omits all others:

```yaml
selected:
  bob: 26
```

This flavor only works on _maps_.

**Remark**

An alternate way to express the same is to use `sum[persons|{}|s,k,v|->v > 25 ? s {k = v} :s]`.


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

## Inline List Expansion

In argument lists or list literals the _list expansion operator_ (`...`) can be
used.  It is a postfix operator on any list expression. It substituted
the list expression by a sequence of the list members. It can be be used
in combination with static list argument denotation.

e.g.:

```yaml
list:
  - a
  - b
  
result: (( [ 1, list..., 2, list... ]  ))
```

evaluates `result` to

```yaml
result:
  - 1
  - a
  - b
  - 2
  - a
  - b
```

The following example demonstrates the usage in combination with the
[_varargs_ operator](#variable_argument_lists) in functions:

```yaml
func: (( |a,b...|-> [a] b ))

list:
  - a
  - b

a: (( .func(1,2,3) ))
b: (( .func("x",list..., "z") ))
c: (( [ "x", .func(list...)..., "z" ] ))
```

evaluates the following results:

```yaml
a:
- 1
- 2
- 3
b:
- x
- a
- b
- z
c:
- x
- a
- b
- z
```

Please note, that the list expansion might span multiple arguments (including the
[_varargs_ parameter](#variable-argument-lists)) in lambda function calls.

## Markers

Nodes of the yaml document can be marked to enable dedicated behaviours for this
node. Such markers are part of the _dynaml_ syntax and may be prepended to
any dynaml expression. They are denoted by the `&` character directly followed 
by a marker name. If the expression is combination of markers and regular
expressions, the expression follows the marker list enclosed in brackets
(for example `(( &temporary( a + b ) ))`).

**Note**: Instead of using a `<<:` insert field to place markers it is possible
now to use `<<<:`, also, which allows to use regular yaml parsers for spiff-like
yaml documents. `<<:` is kept for backward compatibility.

### `(( &temporary ))`

Maps, lists or simple value nodes can be marked as *temporary*. Temporary nodes
are removed from the final output document, but are available during merging and
dynaml evaluation.

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

### `(( &local ))`

The marker `&local` acts similar to `&temporary` but local nodes are always
removed from a stub directly after resolving dynaml expressions. Such nodes
are therefore not available for merging and they are not used for further
merging of stubs and finally the template.


### `(( &inject ))`

This marker requests the marked item to be injected into the next stub level,
even is the hosting element (list or map) does not requests a merge.
This only works if the next level stub already contains the hosting element.

e.g.:

**template.yaml**
```yaml
alice:
 foo: 1
```

**stub.yaml**
```yaml
alice:
  bar: (( &inject(2) ))
  nope: not injected
bob:
  <<: (( &inject ))
  foobar: yep

```

is merged to

```yaml
alice:
  foo: 1
  bar: 2
bob:
  foobar: yep
```

### `(( &default ))`

Nodes marked as *default* will be used as default values
for downstream stub levels. If no such entry is set there it will behave like
`&inject` and implicitly add this node, but existing settings will not be
overwritten.

Maps (or lists) marked as *default* will be considered as values. 
The map is used as a whole as default if no such field is defined downstream.

e.g.:

**template.yaml**
```yaml
data: { }
```

**stub.yaml**
```yaml
data:
  foobar:
    <<: (( &default ))
    foo: claude
    bar: peter
```

is merged to

```yaml
data:
  foobar:
    foo: claude
    bar: peter
```

Their entries will neither be used for overwriting existing downstream values
nor for defaulting non-existng fields of a not defaulted map field.

e.g.:

**template.yaml**
```yaml
data:
  foobar:
    bar: bob
```

**stub.yaml**
```yaml
data:
  foobar:
    <<: (( &default ))
    foo: claude
    bar: peter
```

is merged to

```yaml
data:
  foobar:
    bar: bob
```

If sub sequent defaulting is desired, the fields of a default map must again be
marked as default.

e.g.:

**template.yaml**
```yaml
data:
  foobar:
    bar: bob
```

**stub.yaml**
```yaml
data:
  foobar:
    <<: (( &default ))
    foo: (( &default ("claude") ))
    bar: peter
```

is merged to

```yaml
data:
  foobar:
    foo: claude
    bar: bob
```

**Note**: The behaviour of list entries marked as *default* is undefined.


### `(( &state ))`

Nodes marked as *state* are handled during the merge processing as if the
marker would not be present. But there will be a special handling for enabled
state processing [(option `--state <path>`)](#usage) at the end of the
template processing.
Additionally to the regular output a document consisting only of state nodes
(plus all nested nodes) will be written to a state file. This file will be used
as top-level stub for further merge processings with enabled state support.

This enables to keep state between two merge processings. For regular
merging sich nodes are only processed during the first processing. Later
processings will keep the state from the first one, because those nodes
will be overiden by the state stub added to the end of the sub list.

If those nodes additionally disable merging (for example using 
`(( &state(merge none) ))`) dynaml expressions in sub level nodes may
perform explicit merging using the function `stub()` to refer to
values provided by already processed stubs (especially the implicitly added
state stub). For an example please refer to the 
[state library](libraries/state/README.md).

### `(( &template ))`

Nodes marked as *template* will not be evaluated at the place of their
occurrence. Instead, they will result in a template value stored as value for
the node. They can later be instantiated inside a _dynaml_ expression
(see [below](#templates)).

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

**Note**: Instead of using a `<<:` insert field to place the template marker it is
possible now to use `<<<:`, also, which allows to use regular yaml parsers for
spiff-like yaml documents. `<<:` is kept for backward compatibility.

### `(( *foo.bar ))`

The dynaml expression `*<reference expression>` can be used to evaluate a template somewhere in the yaml document.
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

## Scope References

### `_`

The special reference `_` (_self_) can be used inside of _lambda functions_
and _templates_. They refer to the containing element (the lambda function or
template).

Additionally it can be used to lookup relative reference expressions
starting with the defining document scope of the element skipping intermediate
scopes.

e.g.:

```yaml
node:
  data:
    scope: data
  funcs:
    a: (( |x|->scope ))
    b: (( |x|->_.scope ))
    c: (( |x|->_.data.scope ))
    scope: funcs

call:
  scope: call

  a: (( node.funcs.a(1) ))
  b: (( node.funcs.b(1) ))
  c: (( node.funcs.c(1) ))

```

evaluates `call` to

```yaml
call:
  a: call
  b: funcs
  c: data
  scope: call
```

### `__`

The special reference `__` can be used to lookup references as relative
references starting with the document node hosting the actually evaluated
_dynaml_ expression skipping intermediate scopes.
 
This can, for example be
used to relatively access a lambda value field besides the actual field in
a map. The usage of plain function names is reserved for builtin functions
and are not used as relative references.

This special reference is also available in expressions in _templates_ and
refer to the map node in the template hosting the actually evaluated expression.

e.g.:

```yaml
templates:
  templ:
    <<: (( &template ))
    self: (( _ ))
    value: (( ($self="value") __.self ))
    result: (( scope ))
    templ: (( _.scope ))

  scope: templates


result:
  inst: (( *templates.templ ))
  scope: result
```

evaluates `result` to

```yaml
result:
  inst:
    result: result
    templ: templates
    
    self:
      <<: (( &template ))
      result: (( scope ))
      self: (( _ ))
      templ: (( _.scope ))
      value: (( ($self="value") __.self ))
    value:
      <<: (( &template ))
      result: (( scope ))
      self: (( _ ))
      templ: (( _.scope ))
      value: (( ($self="value") __.self ))
  scope: result
```

or with referencing upper nodes:

```yaml
templates:
  templ:
    <<: (( &template ))
    alice: root
    data:
      foo: (( ($bob="local") __.bob ))
      bar: (( ($alice="local") __.alice ))
      bob: static


result: (( *templates.templ ))
```

evaluates `result`  to

```yaml
result:
  alice: root
  data:
    bar: root
    foo: static
    
    bob: static
```


### `___`

The special reference `___` can be used to lookup references in the outer most
scope. It can therefore be used to access processing bindings specified for a
document processing via command line or API. If no bindings are specified
the document root is used.

Calling `spiff merge template.yaml --bindings bindings.yaml` with a binding of

**bindings.yaml**
```yaml
input1: binding1
input2: binding2
``` 

and the template

**template.yaml**
```yaml
input1: top1
map:
  input: map
  input1: map1
  
  results:
    frommap: (( input1 ))
    fromroot: (( .input1 ))
    frombinding1: (( ___.input1 ))
    frombinding2: (( input2 ))
```

evaluates `map.results`  to

```yaml
  results:
    frombinding1: binding1
    frombinding2: binding2
    frommap: map1
    fromroot: top1
```

### `__ctx.OUTER`

The context field `OUTER` is used for nested [merges](#-mergemap1-map2-). 
It is a list of documents, index 0 is the next outer document, and so on.

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
| `OUTER` | yaml doc | outer documents for nested [merges](#-mergemap1-map2-), index 0 is the next outer document  |
| `BINDINGS` | yaml doc |  the external bindings for the actual processing (see also [___](#___)) |

If external bindings are specified they are the last elements in `OUTER`.

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

1. `||`, `//`
2. White-space separated sequence as concatenation operation (`foo bar`)
3. `-or`, `-and`
4. `==`, `!=`, `<=`, `<`, `>`, `>=`
5. `+`, `-`
6. `*`, `/`, `%`
7. Grouping `( )`, `!`, constants, references (`foo.bar`), `merge`, `auto`, `lambda`, `map[]`, and [functions](#functions)

The complete grammar can be found in [dynaml.peg](dynaml/dynaml.peg).

## String Interpolation

**Attention:** This is an alpha feature. It must be enabled on the command
line with the `--interpolation` option. Also for the spiff library it must
explicitly be enabled. By adding the key `interpolation` to the feature list
stored in the environment variable `SPIFF_FEATURES` this feature will be enabled
by default.

Typically a complete value can either be a literal or a dynaml expression.
For string literals it is possible to use an interpolation syntax to embed
dynaml expressions into strings.

For example

```yaml
data: test
interpolation: this is a (( data ))
```

replaces the part between the double brackets by the result
of the described expression evaluation. Here the brackets can be escaped
by the usual escaping (`((!`) syntax.

Those string literals will implicitly be converted to complete flat dynaml
expressions. The example above will therefore be converted into

`(( "this is a " data ))`

which is the regular dynaml equivalent. The escaping is very ticky, and
may be there are still problems. Quotes inside an embedded dynaml expression
can be escaped to enable quotes in string literals.

Incomplete or partial interpolation expressions will be ignored and 
just used a s string.

Strings inside a dynaml expression are NOT directly interpolated again, thus

```yaml
data: "test"
interpolated: "this is a (( length(\"(( data ))\") data ))"
```
 
will resolve `interpolation` to `this is 10test` and not to `this is 4test`.
 
But if the final string after the expression evaluation again describes a string
interpolation it will be processed, again.

```yaml
data: test
interpolation: this is a (( "(( data ))" data ))
```

will resolve `interpolation` to `this is testtest`.

The embedded dynaml expression must be concatenatable with strings.



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

- _Scopes can be used to parameterize templates_

  Scope literals are also considered when instantiating templates. Therefore
  they can be used to set explicit values for relative reference expressions
  used in templates.

  e.g.:

  ```yaml
  alice: 1
  template:
    <<: (( &template ))
    sum: (( alice + bob ))
  scoped: (( ( $alice = 25, "bob" = 26 ) *template ))
  ```

  evaluates to

  ```yaml
  alice: 1
  template:
    <<: (( &template ))
    sum: (( alice + bob ))
  scoped:
    sum: 51
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

- Defaulting and Requiring Fields

  Traditionally defaulting in _spiff_ is done by a downstream template where the
  playload data file is used as stub.
  
  Fields with simple values can just be specified with their values.
  They will be overwritten by stubs using the regular _spiff_ document merging
  mechanisms.
  
  It is more difficult for maps or lists. If a map is specified in the
  template only its fields will be merged (see above), but it is never
  replaced as a whole by settings in the playload definition files.
  And Lists are never merged.
  
  Therefore maps and lists that should be defaulted as a whole must be specified
  as initial expressions (referential or inline) in the template file.
  
  e.g.: merging of
  
  **template.yaml**
  ```yaml
  defaults:
    <<: (( &temporary ))
    person:
      name: alice
      age: bob
  config:
    value1: defaultvalue
    value2: defaultvalue
    person: (( defaults.person ))
  ```  
  
  and
  
  **payload.yaml**
  ```yaml
   config:
     value2: configured
     othervalue: I want this but don't get it
  ```  
    
  evaluates to 
  
  ```yaml
  config:
    person:
      age: bob
      name: alice
    value1: defaultvalue
      value2: configured
   ```

  In such a scenario the structure of the resulting document is defined by the template.
  All kinds of variable fields or sub-structures must be forseen by the template
  by using `<<: (( merge ))` expressions in maps.
  
  e.g.: changing template to
  
  **template.yaml**
  ```yaml
  defaults:
    <<: (( &temporary ))
    person:
      name: alice
      age: bob
  config:
    <<: (( merge ))
    value1: defaultvalue
    value2: defaultvalue
    person: (( defaults.person ))
  ```  
  
  Known _optional_ fields can be described using the *undefined* (`~~`) expression:

  **template.yaml**
  ```yaml
  config:
    optional: (( ~~ ))
  ```    
  
  Such fields will only be part of the final document if they are defined in
  an upstream stub, otherwise they will be completely removed.
 
  _Required_ fields can be defined with the expression `(( merge ))`. If
  no stub contains a value for this field, the merge cannot be fullfilled and
  an error is reported. If a dedicated message should be shown instead, the
  merge expression can be defaulted with an error function call.
  
  e.g.:
  
  **template.yaml**
  ```yaml
  config:
    password: (( merge || error("the field password is required") ))
  ```

   will produce the following error if no stub contains a value:
   ```
   error generating manifest: unresolved nodes:
   	(( merge || error("the field password is required") ))	in c.yaml	config.password	()	*the field password is required 
   ```
   
   This can be simplified by reducing the expression to the sole `error`
   expression.

   Besides this template based defaulting it is also possible to
   provide defaults by upstream stubs using the [`&default` marker](#-default-).
   Here the payload can be a downstream file.
   
- _X509_ and providing State

  When generating keys or certificates with the [X509 Functions](#x509-functions)
  there will be new keys or certificates for every execution of _spiff_. But 
  it is also possible to use _spiff_ to maintain key state. A very simple script
  could look like this:
  
  ```bash
  #!/bin/bash
  DIR="$(dirname "$0")/state"
  if [ ! -f "$DIR/state.yaml" ]; then
    echo "state:" > "$DIR/state.yaml"
  fi
  spiff merge "$DIR/template.yaml" "$DIR/state.yaml" > "$DIR/.$$" && mv "$DIR/.$$" "$DIR/state.yaml"
  ```
  
  It uses a template file (containing the rules) and a state file with the
  actual state as stub. The first time it is executed there is an empty state
  and the rules are not overridden, therefore the keys and certificates are
  generated. Later on, only additional new fields are calculated, the state
  fields already containing values just overrule the _dynaml_ expressions
  for those fields in the template.
  
  If a re-generation is required, the state file can just be deleted.
  
  A template may look like this:
  
  **state/template.yaml**
  ```yaml
  spec:
    <<: (( &local ))
    ca:
      organization: Mandelsoft
      commonName: rootca
      privateKey: (( state.cakey ))
      isCA: true
      usage:
        - Signature
        - KeyEncipherment
    peer:
      organization: Mandelsoft
      commonName: etcd
      publicKey: (( state.pub ))
      caCert: (( state.cacert ))
      caPrivateKey: (( state.cakey ))
      validity: 100
      usage:
        - ServerAuth
        - ClientAuth
        - KeyEncipherment
      hosts:
        - etcd.mandelsoft.org
  
  state:
    cakey: (( x509genkey(2048) ))
    capub: (( x509publickey(cakey) ))
  
    cacert: (( x509cert(spec.ca) ))
  
    key: (( x509genkey(2048) ))
    pub: (( x509publickey(key) ))
    peer: (( x509cert(spec.peer) ))

  ```
  
  The merge then generates a rootca and some TLS certificate signed with
  this CA.
  
- Generating, Deploying and Accessing Status for Kubernetes Resources

  The [`sync`](#-syncexpr-condition-value-10-) function offers the possibility
  to synchronize the template processing with external content. This can also
  be the output of a command execution. Therefore the template processing
  can not only be used to generate a deployment manifest, but also for
  applying this to a target system and retrieving deployment status values
  for the further processing.
  
  A typical scenario of this kind could be a kubernetes setup including
  a service of type _LoadBalancer_. Once deployed it gets assigned
  status information about the IP address or hostname of the assigned
  load balancer. This information might be required for some other deployment
  manifest.
  
  A simple template for such a deployment could like this:
  
  ```yaml
  service:
    apiVersion: v1
    kind: Service
    metadata:
      annotations:
        dns.mandelsoft.org/dnsnames: echo.test.garden.mandelsoft.org
        dns.mandelsoft.org/ttl: "500"
      name: test-service
      namespace: default
    spec:
      ports:
      - name: http
        port: 80
        protocol: TCP
        targetPort: 8080
      sessionAffinity: None
      type: LoadBalancer
  
  deployment:
     testservice: (( sync[pipe_uncached(service, "kubectl", "apply", "-f", "-", "-o", "yaml")|value|->defined(value.status.loadBalancer.ingress)] ))
  
  
  otherconfig:
     lb: (( deployment.testservice.status.loadBalancer.ingress ))
  
  ```
- Crazy Shit: Graph Analysis with _spiff_

  It is easy to describe a simple graph with knots and edges (for example 
  for a set of components and their dependencies) just by using a map of lists.
  
  <details><summary><b>graph.yaml</b></summary>
  
  ```yaml
  graph:
    a:
    - b
    - c
    b: []
    c:
    - b
    - a
    d:
    - b
    e:
    - d
    - b
  ```
  </details>
  
  Now it would be useful to figure out whether there are dependency cycles or
  to determine ordered transitive dependencies for a component.
  
  Let's say something like this:
  
  <details><summary><b>closures.yaml</b></summary>

  ```yaml
  graph:
  utilities:

  closures: (( utilities.graph.evaluate(graph) ))
  cycles: (( utilities.graph.cycles(closures) ))
  ```
  </details>
  
  Indeed, this can be done with spiff. The only thing required is
  a _"small utilities stub"_.
  
  <details><summary><b>utilities.yaml</b></summary>
  
  ```yaml
  utilities:
    <<: (( &temporary ))
    graph:
      _dep: (( |model,comp,closure|->contains(closure,comp) ? { $deps=[], $err=closure [comp]} :($deps=_._deps(model,comp,closure [comp]))($err=sum[deps|[]|s,e|-> length(s) >= length(e.err) ? s :e.err]) { $deps=_.join(map[deps|e|->e.deps]), $err=err} ))
      _deps: (( |model,comp,closure|->map[model.[comp]|dep|->($deps=_._dep(model,dep,closure)) { $deps=[dep] deps.deps, $err=deps.err }] ))
      join: (( |lists|->sum[lists|[]|s,e|-> s e] ))
      min: (( |list|->sum[list|~|s,e|-> s ? e < s ? e :s :e] ))
  
      normcycle: (( |cycle|->($min=_.min(cycle)) min ? sum[cycle|cycle|s,e|->s.[0] == min ? s :(s.[1..] [s.[1]])] :cycle  ))
      cycle: (( |list|->list ? ($elem=list.[length(list) - 1]) _.normcycle(sum[list|[]|s,e|->s ? s [e] :e == elem ? [e] :s]) :list ))
      norm: (( |deps|->{ $deps=_.reverse(uniq(_.reverse(deps.deps))), $err=_.cycle(deps.err) } ))
      reverse: (( |list|->sum[list|[]|s,e|->[e] s] ))
  
      evaluate: (( |model|->sum[model|{}|s,k,v|->s { k=_.norm(_._dep(model,k,[]))}] ))
      cycles: (( |result|->uniq(sum[result|[]|s,k,v|-> v.err ? s [v.err] :s]) ))
  ```
  </details>
  
  And magically _spiff_ does the work just by calling
  ```bash
  spiff merge closure.yaml graph.yaml utilities.yaml
  ```
  
  <details><summary>And the result is</summary>
  
  ```yaml
     closures:
       a:
         deps:
         - c
         - b
         - a
         err:
         - a
         - c
         - a
       b:
         deps: []
         err: []
       c:
         deps:
         - a
         - b
         - c
         err:
         - a
         - c
         - a
       d:
         deps:
         - b
         err: []
       e:
         deps:
         - d
         - b
         err: []
     cycles:
     - - a
       - c
       - a
     graph:
       a:
       - b
       - c
       b: []
       c:
       - b
       - a
       d:
       - b
       e:
       - d
       - b
  ```
  </details>
  
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

<details><summary><b>Example</b></summary>

```
	(( min_ip("10") ))	in source.yml	node.a.[0]	()	*CIDR argument required
```
</details>

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

If a problem occurs in nested lamba calls the call stack together with the lamba function and is 
local binding is listed.

<details><summary><b>Example</b></summary>

```text
	(( 2 + .func(2) ))	in local/err.yaml	value	()	*evaluation of lambda expression failed: lambda|x|->x > 0 ? _(x - 1) : *(template): {x: 2}
		... evaluation of lambda expression failed: lambda|x|->x > 0 ? _(x - 1) : *(template): {x: 1}
		... evaluation of lambda expression failed: lambda|x|->x > 0 ? _(x - 1) : *(template): {x: 0}
		... resolution of template 'template' failed
			(( z ))	in local/err.yaml	val	()*'z' not found 

```
</details>

In case of parsing errors in dynaml expressions, the error location is shown now.
If it is a multi line expression the line a character/symbol number in that line
is show, otherwise the line numer is omitted.

<details><summary><b>Example</b></summary>

```text
	((
	  2 ++ .func(2)
	))	in local/err.yaml	faulty	()	*parse error near line 2 symbol 2 - line 2 symbol 3: " " 

```
</details>

# Using _spiff_ as Go Library

_Spiff_ provides a Go package (`spiffing`) that can be used to include _spiff_ templates in Go programs.

An example program could look like this:

```go
import (
	"fmt"
	"math"
	"os"

	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/spiffing"
)

func func_pow(arguments []interface{}, binding dynaml.Binding) (interface{}, dynaml.EvaluationInfo, bool) {
	info := dynaml.DefaultInfo()

	if len(arguments) != 2 {
		return info.Error("pow takes 2 arguments")
	}

	a, b, err := dynaml.NumberOperands(arguments[0], arguments[1])

	if err != nil {
		return info.Error("%s", err)
	}
	_, i := a.(int64)
	if i {
		r := math.Pow(float64(a.(int64)), float64(b.(int64)))
		if float64(int64(r)) == r {
			return int64(r), info, true
		}
		return r, info, true
	} else {
		return math.Pow(a.(float64), b.(float64)), info, true
	}
}

var state = `
state: {}
`
var stub = `
unused: (( input ))
ages:
  alice: (( pow(2,5) ))
  bob: (( alice + 1 ))
`

var template = `
state:
  <<<: (( &state ))
  random: (( rand("[:alnum:]", 10) )) 
ages: (( &temporary ))

example:
  name: (( input ))  # direct reference to additional values 
  sum: (( sum[ages|0|s,k,v|->s + v] ))
  int: (( pow(2,4) ))
  float: 2.1
  pow: (( pow(1.1e1,2.1) ))
`

func Error(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	values := map[string]interface{}{}
	values["input"] = "this is an input"

	functions := spiffing.NewFunctions()
	functions.RegisterFunction("pow", func_pow)

	spiff, err := spiffing.New().WithFunctions(functions).WithValues(values)
	Error(err)
	pstate, err := spiff.Unmarshal("state", []byte(state))
	Error(err)
	pstub, err := spiff.Unmarshal("stub", []byte(stub))
	Error(err)
	ptempl, err := spiff.Unmarshal("template", []byte(template))
	Error(err)
	result, err := spiff.Cascade(ptempl, []spiffing.Node{pstub}, pstate)
	Error(err)
	b, err := spiff.Marshal(result)
	Error(err)
	newstate, err := spiff.Marshal(spiff.DetermineState(result))
	Error(err)
	fmt.Printf("==== new state ===\n")
	fmt.Printf("%s\n", string(newstate))
	fmt.Printf("==== result ===\n")
	fmt.Printf("%s\n", string(b))
}
```

It supports
 - transforming file data to and from spiffs internal node representation
 - the processing of stubs and templates with or without state handling
 - defining an outer binding for injected path names
 - defining additional spiff functions
 - enabling/disabling command execution and/or filesystem operations
 - using a [virtual filesystem](http://github.com/mandelsoft/vfs) for
   file system operations
