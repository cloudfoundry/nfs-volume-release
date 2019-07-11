#!/bin/bash -eux

export DOCKER_STORAGE_OPTIONS='--storage-opt dm.basesize=100G'
start-bosh

export DOCKER_TMP_DIR=$(find /tmp/ -name "tmp.*")
export DOCKER_HOST=$(ps aux | grep dockerd | grep -o '\-\-host tcp.*4243' | awk '{print $2}')

eval "$(cat /tmp/local-bosh/director/env)"

docker \
--tls \
--tlscacert=${DOCKER_TMP_DIR}/ca.pem \
--tlscert=${DOCKER_TMP_DIR}/cert.pem \
--tlskey=${DOCKER_TMP_DIR}/key.pem run \
--network=director_network \
-v $PWD/nfs-volume-release/:/nfs-volume-release \
-v $PWD/mapfs-release:/mapfs-release \
-v /tmp:/tmp \
-w /nfs-volume-release/src/bosh_release \
-t \
-i \
--env BOSH_ENVIRONMENT=10.245.0.3 \
--env BOSH_CLIENT=${BOSH_CLIENT} \
--env BOSH_CLIENT_SECRET=${BOSH_CLIENT_SECRET} \
--env BOSH_CA_CERT=${BOSH_CA_CERT} \
--env NFS_VOLUME_RELEASE_PATH=/nfs-volume-release \
--env MAPFS_RELEASE_PATH=/mapfs-release \
cfpersi/bosh-release-tests \
ginkgo .