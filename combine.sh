#!/bin/bash
set -euo pipefail

if [[ $# -le 0 ]]
then
    echo 'No arguments provided.'
    exit 1
fi

BASE_FILE=$1
PROJECT_ROOT=$(head -n 1 $BASE_FILE | jq -r '.projectRoot')
NUM_IDS=$(wc -l $BASE_FILE | awk '{print $1}')
FILES=( "$BASE_FILE" )
shift

while [[ $# -gt 0 ]]
do
    NEXT_FILE=$1
    NEXT_ROOT=$(head -n 1 $BASE_FILE | jq -r '.projectRoot')
    ESCAPED_PROJECT_ROOT=$(printf '%s\n' "$PROJECT_ROOT" | sed 's:[\\/&]:\\&:g;$!s/$/\\/')
    ESCAPED_NEXT_ROOT=$(printf '%s\n' "$PROJECT_ROOT" | sed 's:[][\\/.^$*]:\\&:g')

    tail -n +3 $NEXT_FILE | jq -rc "(.id,.inV,.outV,.document,.inVs[]? | select(. != null)) |= . + $NUM_IDS" | sed "s/$ESCAPED_NEXT_ROOT/$ESCAPED_PROJECT_ROOT/" > "/tmp/$NEXT_FILE.tmp"&
    FILES+=("/tmp/$NEXT_FILE.tmp")

    NUM_NEW_IDS=$(wc -l $NEXT_FILE | awk '{print $1}')
    NUM_IDS=$((NUM_IDS + NUM_NEW_IDS - 2))
    shift
done

wait

for FILE in "${FILES[@]}"
do
    cat $FILE
done
