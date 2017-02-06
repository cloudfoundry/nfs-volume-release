#!/usr/bin/env bash
set -e

scripts_path=./$(dirname $0)

fly -t persi execute -c $scripts_path/ci/run_pats_nfsvolume.build.yml -i nfs-volume-release=/Users/pivotal/workspace/nfs-volume-release
