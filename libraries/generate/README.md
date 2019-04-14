
# Manifest Generation

This library offers lambda expressions that can be used to generate
YAML manifests from templates files based on input values and value
templates.

The package name is `utilities.generate`.

## Generate a List of YAML Manifests taken from a Multi Document Template

```
    generateFile(<values>,<stubs>,<file>) -> list of manifests
```

- `values` is a map template or map defining some input values used as
           top level stub for values template files
- `stubs`  is a list of optional stub files (stubs might be ~ or []) used
           as initial value template and stubs
- `file`   is a template file containing yaml manifests processed 
           using the input value merge result using the binding
          `values` or `settings`
          
The given values are processed as top-level stub for the given stubs to produce
the value setiings for the manifest template processing.

The result is a list of processed manifests taken from the template file.

## Generate a List of YAML Manifests taken from a List of Multi Document Templates

```
    generateFiles(<values>,<stubs>,<files>) -> list of manifests
```

- `values` is a map template or map defining some input values used as
           top level stub for values template files
- `stubs`  is a list of optional stub files (stubs might be ~ or []) used
           as initial value template and stubs
- `files`  is a list of template files containing yaml manifests processed 
           using the input value merge result using the binding
          `values` or `settings`
          
The given values are processed as top-level stub for the given stubs to produce
the value setiings for the manifest template processing.

The result is a list of processed manifests taken from all the template files in
the given order.

## Generate a List of YAML Manifests taken from Multi Document Templates found in a Directory

```
    generateDir(<values>,<stubs>,<dir>) -> list of manifests
```

- `values` is a map template or map defining some input values used as
           top level stub for values template files
- `stubs`  is a list of optional stub files (stubs might be ~ or []) used
           as initial value template and stubs
- `dir`    is a directory containing template files (suffix .yaml) to be processed 
           using the input value merge result using the binding
          `values` or `settings`
          
The given values are processed as top-level stub for the given stubs to produce
the value setiings for the manifest template processing.

The result is a list of processed manifests taken from all the template files in
the given order.

## Generate a List of YAML Manifests taken from standard Chart Directory Layout

```
    generateChart(<values>,<dir>) -> list of manifests
```

- `values` is a map template or map defining some input values used as
           top level stub for values template files
- `dir`    base directoty of the chart.
          
This function uses `generateDir` to process manifest template files from the
directory `<dir>/templates`, using `<dir>/values.yaml` as value template