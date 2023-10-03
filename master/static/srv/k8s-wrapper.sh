#!/bin/bash
#!/usr/bin/env bash
# Usage:
# k8s-wrapper.sh: {realEntryPointArgs}...
#
# This wraps the realEntryPointArgs for Kubernetes to replace an init-container.
# This script just untars the additional files and calls the regular entrypoint.
# TODO(DET-xxxx) this might not be needed and can just pass files untared.
set -eE
trap 'echo >&2 "FATAL: Unexpected error terminated k8s-wrapper container initialization.  See error messages above."' ERR


# Unconditional log method
# Args: {Level} {Message}...
log() {
    echo -e "$*" >&2
}


# Script takes args: NUM_FILES SRC DST and un-tars files from SRC to DST.

if [ $1 -le 0 ]; then
    exit 0
fi

for var in $(seq 1 1 $1); do
    IDX=$((var - 1))
    SRC=$2/$IDX.tar.gz
    DST=$3/$IDX
    mkdir $DST
    tar -xvf $SRC -C $DST
done


log "INFO: executing $@" >&2
exec "$@"
