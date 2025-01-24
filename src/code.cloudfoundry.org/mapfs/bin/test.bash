#!/bin/bash

set -eu
set -o pipefail

THIS_FILE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

configure_fstest(){
    local fstest_dir=${1:?Provide fstest dir}
    pushd ${fstest_dir} > /dev/null
    # remove all the tests that fail because of the way that mapfs works
    rm -f tests/chmod/00.t
    rm -f tests/chmod/05.t
    rm -f tests/chmod/07.t
    rm -f tests/chmod/11.t
    rm -f tests/chown/00.t
    rm -f tests/chown/02.t
    rm -f tests/chown/03.t
    rm -f tests/chown/05.t
    rm -f tests/chown/07.t
    rm -f tests/link/00.t
    rm -f tests/link/02.t
    rm -f tests/link/03.t
    rm -f tests/link/06.t
    rm -f tests/link/07.t
    rm -f tests/link/09.t
    rm -f tests/link/11.t
    rm -f tests/mkdir/00.t
    rm -f tests/mkdir/05.t
    rm -f tests/mkdir/06.t
    rm -f tests/mkfifo/00.t
    rm -f tests/mkfifo/05.t
    rm -f tests/mkfifo/06.t
    rm -f tests/open/00.t
    rm -f tests/open/02.t
    rm -f tests/open/03.t
    rm -f tests/open/05.t
    rm -f tests/open/06.t
    rm -f tests/open/07.t
    rm -f tests/open/08.t
    rm -f tests/rename/00.t
    rm -f tests/rename/04.t
    rm -f tests/rename/05.t
    rm -f tests/rename/09.t
    rm -f tests/rename/10.t
    rm -f tests/rmdir/07.t
    rm -f tests/rmdir/08.t
    rm -f tests/rmdir/11.t
    rm -f tests/symlink/05.t
    rm -f tests/symlink/06.t
    rm -f tests/truncate/00.t
    rm -f tests/truncate/05.t
    rm -f tests/truncate/06.t
    rm -f tests/unlink/00.t
    rm -f tests/unlink/05.t
    rm -f tests/unlink/06.t
    rm -f tests/unlink/11.t


    make clean
    make all
    popd > /dev/null
}
cleanup_fstest() {
    local fstest_dir=${1:?Provide fstest dir}
    pushd ${fstest_dir} > /dev/null
    make clean
    if [[ "$(git remote -v)" =~ "zfsonlinux/fstest" ]]; then
        git checkout .
        git clean -xdff
    fi
    popd > /dev/null
}

run_fstest() {
    local fstest_dir="${THIS_FILE_DIR}/../fstest"
    configure_fstest ${fstest_dir}
    mkdir -p foo1 foo2
    chown 1000:1000 foo2
    go run main.go -uid 1000 -gid 1000 foo1 foo2 &
    pushd foo1 > /dev/null
    prove -r "${fstest_dir}/"
    popd > /dev/null
    cleanup_fstest ${fstest_dir}
    umount -fl foo1 || true
    rm -rf foo1 foo2
}

run_fstest
# shellcheck disable=SC2068
# Double-quoting array expansion here causes ginkgo to fail
# Tee output to a log file but exclude component/test logs from stdout so
# concourse output is not overloaded
go run github.com/onsi/ginkgo/v2/ginkgo ${@} | tee /tmp/simulation-output.log | grep -v '{"timestamp"'
