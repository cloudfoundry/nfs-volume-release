#!/usr/bin/env bash
set -e

scripts_path=./$(dirname $0)

fly -t persi execute -c $scripts_path/ci/run_kerbdriver_cert_tcp.build.yml -i nfs-volume-release=/Users/pivotal/workspace/nfs-volume-release --privileged
