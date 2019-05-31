#!/bin/bash -eux

docker run \
-t \
-i \
--privileged \
-v /Users/pivotal/workspace/nfs-volume-release/:/nfs-volume-release \
-v /Users/pivotal/workspace/mapfs-release:/mapfs-release \
--workdir=/ \
bosh/main-bosh-docker \
/nfs-volume-release/scripts/run-bosh-release-tests-in-docker-env.sh
