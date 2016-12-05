#!/bin/bash

set -x

function usage() {
    echo "Usage: startdriver.sh listen-addr drivers-path"
    echo $1
    exit 1
}

if [ -z "$1" ] || [ "$1" == "" ]; then usage 'Listen address not set'; fi
if [ -z "$2" ] || [ "$2" == "" ]; then usage 'Drivers path not set'; fi

listen_addr=$1
drivers_path=$2

cd `dirname $0`

killall -9 nfsv3driver

mkdir -p ../mountdir

 ../exec/nfsv3driver -listenAddr="${listen_addr}" -transport="$TRANSPORT" -driversPath="$drivers_path" &
