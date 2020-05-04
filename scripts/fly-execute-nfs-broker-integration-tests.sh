#!/bin/bash -eux

fly -t persi \
execute \
-c ~/workspace/nfs-volume-release/scripts/ci/run_broker_integration.build.yml \
-i credhub=${HOME}/workspace/credhub \
-i nfs-volume-release=${HOME}/workspace/nfs-volume-release \
-i nfs-volume-release-concourse-tasks=${HOME}/workspace/nfs-volume-release \
-p
