#!/bin/bash -eux

fly -t persi \
execute \
-c ~/workspace/nfs-volume-release/scripts/ci/run_broker_integration.build.yml \
-i nfs-volume-release=/Users/pivotal/workspace/nfs-volume-release \
-j persi/nfsbroker-tests \
-p