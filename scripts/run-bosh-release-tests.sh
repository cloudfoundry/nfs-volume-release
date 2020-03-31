#!/bin/bash -eux

pushd ~/workspace/nfs-volume-release
    bosh reset-release
popd

pushd ~/workspace/mapfs-release
    bosh reset-release
popd

docker run \
-t \
-i \
--privileged \
-e DEV=true \
-v ~/workspace/nfs-volume-release/:/nfs-volume-release \
-v ~/workspace/mapfs-release:/mapfs-release \
--workdir=/ \
bosh/main-bosh-docker \
/nfs-volume-release/scripts/run-bosh-release-tests-in-docker-env.sh
