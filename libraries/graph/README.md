
# Directed Graph Analysis

This graph library offers some useful lambda expression to work
with directed graphs.

A graph is given by a dicated yaml map hosting the nodes represented by
their names. Each name sub-node represents the edges by holding a list
of node names 

The package name is `utilities.graph`.

## Invert a Directed Graph

```
    invert(model)
```

## Evaluate a directed Graph

``` 
    evaluate(model)
```

The result a a graph evaluation map with entries for every node
providing the evaluated node info:
- `deps` The dependency closure
- `err`: Cyclic dependencies for this node
- `missing`: dangling edges (without a node in the graph)

## List all Cycles

```
    cycles(evaluatedmodel) -> list of lists
```

List all normalized dependencies cycles in a graph using its
evaluated model.

## Provide a global Execution Order

```
    order(evaluatedmodel) -> list
```

Uses the evaluated model to determine a global execution order
for nodes included i the graph.

## Reverse a List

```
    reverse(list)  -> list
```

Reverse the order of elements in a list.