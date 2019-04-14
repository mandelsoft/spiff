#!/bin/bash

#
# evaluate the template with a dedicated
# stub file holding the state from a previous call
# and generate a new version for this state after
# successful merging.
#

set -e

if [ -f "state.yaml" ]; then
  V="$(spiff merge template.yaml state.yaml)"
else
  V="$(spiff merge template.yaml)"
fi

spiff merge --select state - <<<"$V" >state.yaml

echo "$V"