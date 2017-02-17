#!/bin/bash -e
P="$(pwd)"
O="mandelsoft/spiff"
if [ ! -d "../../$O" ]; then
  echo "preparing original path"
  cd ../..
  mkdir -p "$(dirname "$O")"
  ln -s "$P" "$O"
  cd "$O"
  echo "now in $(pwd)"
fi
echo getting dependencies
godep get -v
echo getting test dependencies
godep get -v -t
godep go test -i ./...

