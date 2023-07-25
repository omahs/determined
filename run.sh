#!/bin/bash

set -e

if [ $# -ne 1 ]; then
  echo "Usage: $0 N"
  exit 1
fi

N=$1

export PGPASSWORD=postgres

psql -h 127.0.0.1 -U postgres -d determined -f ddl.sql

for ((i = 1; i <= N; i++)); do
  psql -h 127.0.0.1 -U postgres -d determined -f workerscript.sql -v number_workers=$N -v worker_index=$i &
done

wait

psql -h 127.0.0.1 -U postgres -d determined -f end.sql

echo "Migration sucessfully completed"
