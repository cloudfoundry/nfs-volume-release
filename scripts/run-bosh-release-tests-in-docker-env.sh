#!/bin/bash -ex

COMMAND_TO_RUN="go get github.com/onsi/ginkgo/ginkgo && ginkgo -nodes 1 -v ."
if [[ -n "$DEV" ]]; then
    COMMAND_TO_RUN='bash'
fi

export DOCKER_STORAGE_OPTIONS='--storage-opt dm.basesize=100G'
. start-bosh

source /tmp/local-bosh/director/env
export DOCKER_TMP_DIR=$(find /tmp/ -name "tmp.*")

docker \
--tls \
--tlscacert=${DOCKER_TMP_DIR}/ca.pem \
--tlscert=${DOCKER_TMP_DIR}/cert.pem \
--tlskey=${DOCKER_TMP_DIR}/key.pem \
run \
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
--env GOPATH=/nfs-volume-release \
--env MAPFS_RELEASE_PATH=/mapfs-release \
--env GO111MODULE=off \
cfpersi/bosh-release-tests \
bash -c "$COMMAND_TO_RUN"
