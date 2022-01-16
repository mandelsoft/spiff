package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("graph scenario", func() {

	graph := parseYAML(`
---
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
`)

	closures := parseYAML(`
graph:
utilities:

closures: (( utilities.graph.evaluate(graph) ))
cycles: (( utilities.graph.cycles(closures) ))
`)

	utilities := parseYAML(`
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
`)

	It("handles nil", func() {
		resolved := parseYAML(`
---
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
`)
		Expect(closures).To(CascadeAs(resolved, graph, utilities))
	})
})
