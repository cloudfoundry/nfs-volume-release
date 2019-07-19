#!/bin/bash

docker run -it \
-v /Users/pivotal/workspace/nfs-volume-release/src:/go/src/ \
-v /Users/pivotal/workspace/nfs-volume-release:/nfs-volume-release \
-v /Users/pivotal/workspace/mapfs-release:/mapfs-release \
-w / \
-e TRANSPORT=tcp \
--privileged \
-u root \
cfpersi/nfs-integration-tests \
./nfs-volume-release/scripts/ci/run_docker_driver_integration_tests
