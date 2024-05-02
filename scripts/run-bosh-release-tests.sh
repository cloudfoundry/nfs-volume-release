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
-v $HOME/workspace/nfs-volume-release/:/nfs-volume-release \
-v $HOME/workspace/mapfs-release:/mapfs-release \
--workdir=/ \
bosh/docker-cpi \
/nfs-volume-release/scripts/run-bosh-release-tests-in-docker-env.sh
