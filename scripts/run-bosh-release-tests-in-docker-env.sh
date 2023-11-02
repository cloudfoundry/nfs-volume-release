#!/bin/bash -ex

COMMAND_TO_RUN='go run github.com/onsi/ginkgo/ginkgo -r -nodes 1 -v .'
if [[ -n "$DEV" ]]; then
    COMMAND_TO_RUN='bash'
fi

update-alternatives --set iptables /usr/sbin/iptables-legacy
update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy

INTERNAL_NAMESERVER="$( cat /etc/resolv.conf | grep -im 1 '^nameserver' | cut -d ' ' -f2 )"

cp /usr/local/bosh-deployment/docker/cloud-config.yml{,.original}

bosh int \
  /usr/local/bosh-deployment/docker/cloud-config.yml.original \
  -o /usr/local/bosh-deployment/misc/dns.yml \
  -v internal_dns="[ $INTERNAL_NAMESERVER ]" \
  > /usr/local/bosh-deployment/docker/cloud-config.yml

cp /usr/local/bosh-deployment/bosh.yml{,.original}

bosh int \
  /usr/local/bosh-deployment/bosh.yml.original \
  -o /usr/local/bosh-deployment/misc/dns.yml \
  -v internal_dns="[ $INTERNAL_NAMESERVER ]" \
  > /usr/local/bosh-deployment/bosh.yml

export DOCKER_STORAGE_OPTIONS='--storage-opt dm.basesize=100G'
. start-bosh

source /tmp/local-bosh/director/env
export DOCKER_TMP_DIR=$(find /tmp/ -name "tmp.*")

export BOSH_ENVIRONMENT=10.245.0.3
export BOSH_CLIENT=${BOSH_CLIENT}
export BOSH_CLIENT_SECRET=${BOSH_CLIENT_SECRET}
export BOSH_CA_CERT=${BOSH_CA_CERT}
export NFS_VOLUME_RELEASE_PATH=$PWD/nfs-volume-release
export MAPFS_RELEASE_PATH=$PWD/mapfs-release

pushd $PWD/nfs-volume-release/src/bosh-release
  echo '**** from the bash shell, run ginkgo -nodes 1 -r -v .'
  $COMMAND_TO_RUN
popd
