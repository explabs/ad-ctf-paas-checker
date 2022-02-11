#!/bin/sh
set -e

FILE=/scripts/requirements.txt
if test -f "$FILE"; then
    pip3 install -r $FILE
fi

./app