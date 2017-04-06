#!/usr/bin/env bash
set -e

usage() {
  >&2 echo "    Usage:
    $0 {ldap service username} {ldap service password}
"
  exit 1
}

if [ -z "$1" ]; then
  usage
fi
if [ -z "$2" ]; then
  usage
fi

scripts_path=./$(dirname $0)
export SOURCE=nfs://10.10.200.72/export2/certs
export LDAP_SVC_USER=$1
export LDAP_SVC_PASS=$2
export LDAP_HOST="ec2-54-159-123-136.compute-1.amazonaws.com"
export LDAP_USER_FQDN="cn=Users,dc=corp,dc=persi,dc=cf-app,dc=com"

fly -t persi execute -c $scripts_path/ci/run_driver_cert_ldap.build.yml -i nfs-volume-release=/Users/pivotal/workspace/nfs-volume-release -i lib-nfs=/Users/pivotal/workspace/libnfs -i fuse-nfs=/Users/pivotal/workspace/fuse-nfs --privileged
