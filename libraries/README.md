
```    
                          _  __  __   _ _ _                    _           
                ___ _ __ (_)/ _|/ _| | (_) |__  _ __ __ _ _ __(_) ___  ___ 
               / __| '_ \| | |_| |_  | | | '_ \| '__/ _` | '__| |/ _ \/ __|
               \__ \ |_) | |  _|  _| | | | |_) | | | (_| | |  | |  __/\__ \
               |___/ .__/|_|_| |_|   |_|_|_.__/|_|  \__,_|_|  |_|\___||___/
                   |_|                                                                                         
```

---

# Useful Utility Stubs for _spiff++_

This folder contains some libraries to be used with _spiff++_. 
The libararies are regular yaml files containing dynaml lambda functions
and templates, that might be useful for various use cases.

They are just added as stub to a spiff processing, or they can be read
into an existing template/stub file with the _ read_ function.

All the libraries offered here follow the same basic layout that
makes them additive. Every package is stored in a separate file. They
can just be added as spiff merge stub and aggregate under
the top-level node `utilities` providing an own sub-level node representing
the package finally offering the functionality.

```
  utilities:
  
  usage: (( utilities.<package>.<function>(...) ))
```

The libraries now use the _injection_ feature, therefore the `utilities`
node is avalaible, even it is not specified in a processed yaml document.
The `utilities` node is _temporary_ by default. This assures, that the 
provided functions are not part of the final document.

## Library Overview

### Certificate Generation

This [library](certs/README.md) offers useful lambda functions to generate
certificates for self signed certificate authorities, key/certificate pairs,
ssh key pairs and simple secrets using the [state library](state/README.md).

### Manifest Generation

This [library](generate/README.md) offers manifest generation functions
processing template files with input values.

### Simple Graph Library

This [library](graph/README.md) offers simple graph analysis for directed graphs, like cycle
detection and dependeny closures, or inverting a directed graph.

### State Handling

This [library](state/README.md) offers a simple state handling support together 
with a small shell script example.

