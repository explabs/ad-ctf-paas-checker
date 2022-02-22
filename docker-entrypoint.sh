#!/bin/sh
set -e

FILE=/checker/scripts/requirements.txt
if test -f "$FILE"; then
    pip3 install -r $FILE
fi

./app