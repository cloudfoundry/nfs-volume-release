#!/bin/bash

absolute_path() {
  (cd $1 && pwd)
}

scripts_path=$(absolute_path `dirname $0`)

CEPHFS_RELEASE_DIR=${CEPHFS_RELEASE_DIR:-$(absolute_path $scripts_path/..)}

echo CEPHFS_RELEASE_DIR=$CEPHFS_RELEASE_DIR
