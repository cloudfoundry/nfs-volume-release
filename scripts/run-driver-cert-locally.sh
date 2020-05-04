#!/bin/bash

docker run -it \
-v ${HOME}/workspace/nfs-volume-release/src:/go/src/ \
-v ${HOME}/workspace/nfs-volume-release:/nfs-volume-release \
-v ${HOME}/workspace/mapfs-release:/mapfs-release \
-v ${HOME}/workspace/docker_driver_integration_tests:/docker_driver_integration_tests \
-w / \
-e TRANSPORT=tcp \
-e TEST_PACKAGE=docker_driver_integration_tests \
-e BINDINGS_FILE=nfs-bindings.json \
-e ERROR_CHECK_READONLY_MOUNTS=false \
--privileged \
-u root \
cfpersi/nfs-integration-tests \
./nfs-volume-release/scripts/ci/run_docker_driver_integration_tests
