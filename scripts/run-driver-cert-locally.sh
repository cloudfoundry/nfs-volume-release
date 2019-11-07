#!/bin/bash

docker run -it \
-v $HOME/workspace/nfs-volume-release/src:/go/src/ \
-v $HOME/workspace/docker_driver_integration_tests:/docker_driver_integration_tests \
-v $HOME/workspace/nfs-volume-release:/nfs-volume-release \
-v $HOME/workspace/mapfs-release:/mapfs-release \
-w / \
-e TRANSPORT=tcp \
-e TEST_PACKAGE=docker_driver_integration_tests/ \
--privileged \
-u root \
cfpersi/nfs-integration-tests \
./nfs-volume-release/scripts/ci/run_docker_driver_integration_tests
