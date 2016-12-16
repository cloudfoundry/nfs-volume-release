#!/bin/bash
#generate_manifest.sh

set -e -x

home="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && cd .. q&& pwd )"
templates=${home}/templates

MANIFEST_NAME=nfs-test-server-aws-manifest

spiff merge ${templates}/nfs-test-server-manifest-aws.yml $1 $2 > $PWD/$MANIFEST_NAME.yml

echo manifest written to $PWD/$MANIFEST_NAME.yml
