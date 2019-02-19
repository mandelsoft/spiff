#!/bin/bash -e
P="$(pwd)"
O="mandelsoft/spiff"
if [ ! -d "../../$O" ]; then
  echo "preparing original path"
  cd ../..
  mkdir -p "$(dirname "$O")"
  mv "$P" "$O"
  cd "$O"
  echo "now in $(pwd)"
fi
