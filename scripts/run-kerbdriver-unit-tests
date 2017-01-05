#!/bin/bash

set -e

scripts_path=./$(dirname $0)
eval $($scripts_path/get_paths.sh)

pushd src/code.cloudfoundry.org/kerbdriver
  ginkgo -r -keepGoing -p -trace -randomizeAllSpecs -progress "$@"
popd
