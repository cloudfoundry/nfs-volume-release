#!/bin/bash

absolute_path() {
  (cd $1 && pwd)
}

scripts_path=$(absolute_path `dirname $0`)

NFS_RELEASE_DIR=${NFS_RELEASE_DIR:-$(absolute_path $scripts_path/..)}

echo NFS_RELEASE_DIR=$NFS_RELEASE_DIR
