#!/bin/bash
#generate_manifest.sh

set -e -x

usage () {
    echo "Usage: generate_manifest.sh cf-manifest director-stub iaas-stub nfs-props-stub"
    echo " * default"
    exit 1
}

home="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && cd .. q&& pwd )"
templates=${home}/templates

if [[  "$1" == "bosh-lite" || "$1" == "aws" || -z $1 ]]
  then
    usage
fi

MANIFEST_NAME=nfsvolume-aws-manifest

spiff merge ${templates}/nfsvolume-manifest-aws.yml \
$1 \
$2 \
$3 \
$4 \
${templates}/toplevel-manifest-overrides.yml \
> $PWD/$MANIFEST_NAME.yml

echo manifest written to $PWD/$MANIFEST_NAME.yml
