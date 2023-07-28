#!/usr/bin/env bash
#
#  Launch multiple "psql" processes concurrently to achieve a level of
#  parallelization for the "add summary metrics" migration that's part
#  of the HPE MLDE 0.22.0 upgrade
#
#  The only argument, which is required, is a positive integer specifying
#  the number of concurrent "workers" to spawn.
# 
#  The user is given an opportunity to confirm database connection info
#  and number of workers prior to any "psql" processes are launched. If 
#  the info isn't what is intended, typing 'n' at the prompt aborts the
#  script.  Typing 'y' allows the script to start the migration.
#
#  General workflow description (in three steps after input validation):
# 
#  1. ddl.sql (single): adds summary_metrics et al. columns to trials table
# 
#  2. workerscript.sql (parallel): designed to run as multiple processes
#     running in parallel.  The script will display when each worker
#     process completes and will proceed to step 3 only when all workers
#     have completed
# 
#  3. end.sql (single): sets to NOW() the summary_metrics_timestamp for
#     all rows in the trials table.
#
#  This script and workerscript.sql send timestamped messages to stdout 
#  to help provide a sense of progress as well as help with diagnosis
#  should there be a problem.
#  
#  Also, each worker logs to its own file, "add_summary_metrics_PID.log"
#  where PID is the process id of the "psql" process running the worker.
#  Some of the same messages logged to stdout are also logged in these
#  individual files.  One thing to note is each worker log file will
#  contain a sorted list of the trial ids which the worker will touch.

#

#  These pg* parameters will set the analogous PG* environment variables
#  used by "psql" if these envvars don't already exist
#
pghost='127.0.0.1'
pgport='5432'
pguser='postgres'
pgpassword='no-see-um'
pgdatabase='determined'


USAGE="Usage: ${0##*/} N\n\twhere N is a positive integer specifying the number of concurrent workers"
datefmt='+%F_%T'

set -e
shopt -s extglob xpg_echo

if [[ "$1" != +([0-9]) ]]; then
  echo "${USAGE}"
  exit 1
fi
declare -i N=$1

if (( N < 1 )); then
  echo "${USAGE}"
  exit 1
fi

# set PG envvars if not already set
#
PG_ENVVARS=$(env | egrep '^PG(HOST|PORT|USER|PASSWORD|DATABASE)='; true)
[[ "${PG_ENVVARS}" != *PGHOST=* ]] && export PGHOST=${pghost}
[[ "${PG_ENVVARS}" != *PGPORT=* ]] && export PGPORT=${pgport}
[[ "${PG_ENVVARS}" != *PGUSER=* ]] && export PGUSER=${pguser}
[[ "${PG_ENVVARS}" != *PGPASSWORD=* ]] && export PGPASSWORD=${pgpassword}
[[ "${PG_ENVVARS}" != *PGDATABASE=* ]] && export PGDATABASE=${pgdatabase}
echo
env | sed -rn '/^PG(HOST|PORT|USER|DATABASE)=/p;/^PGPASSWORD=/s/=.*$/=<MASKED>/p'

while [[ "${ANS}" != @([NYny]) ]]; do
  echo
  read -n1 -p "Run ${N} parallel workers to the above database? (y|n): " ANS
done
if [[ "${ANS}" = @([Nn]) ]]; then
  echo "  ..aborting then"
  exit 1
fi
echo # newline

echo
echo $(date ${datefmt}) 'running ddl.sql..'
psql -f ddl.sql -v ON_ERROR_STOP=on
echo

for ((i = 1; i <= N; i++)); do
  psql -f workerscript.sql -v number_workers=$N -v worker_index=$i -v ON_ERROR_STOP=on &
  echo $(date ${datefmt}) "running workerscript.sql, ${i} of ${N} is PID $!.."
  pidlist="${pidlist} $!"
done
echo

while [[ -n "${pidlist}" ]]; do
    wait -p childpid -n ${pidlist}
    echo $(date ${datefmt}) 'PID' ${childpid} 'exited'
    pidlist=$(echo "${pidlist}" | sed "s/ ${childpid}//")
done
echo

echo $(date ${datefmt}) 'running end.sql..'
psql -f end.sql -v ON_ERROR_STOP=on

echo "Migration sucessfully completed"
