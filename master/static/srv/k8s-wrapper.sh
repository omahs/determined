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

ls /run/determined/temp/tar/src

# I'm kinda worried this will break something.
# Like this requires extra permissions right???
# Theoretically we won't be able to write to it.
# Okay we might need to undo some stuff.
# We could still have this volume to hack around permissions...
# We would still need megastuff
#tar -xvf /run/determined/temp/tar/src/mega.tar.gz -C / # How can we put these in right place?



src_dir="/run/determined/temp/tar/src/"
dst_dir="/run/determined/temp/tar/dst/"

# Loop through each .tar.gz file in the src directory
for src_file in "${src_dir}"*.tar.gz; do
    # Extract the file name without extension
    file_name=$(basename "$src_file" .tar.gz)
    # Create a destination directory
    mkdir -p "${dst_dir}${file_name}"

    echo "IN" $src_file "NOW DEST" "${dst_dir}${file_name}"
    # Extract the tar.gz file into the destination directory
    tar -xvf "$src_file" -C "${dst_dir}${file_name}"
done


#for var in $(seq 1 1 $1); do
#    IDX=$((var - 1))
#    SRC=/run/determined/temp/tar/src/$IDX.tar.gz
#    DST=/run/determined/temp/tar/dst/$IDX
#    mkdir $DST
#    tar -xvf $SRC -C $DST
#done

sleep 300

log "INFO: executing $@" >&2
exec "$@"
