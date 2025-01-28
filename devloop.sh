#!/bin/bash

while true
do
    echo -------------------------------------------------------------
    inotifywait -r -e modify *
    padsp signalgen -t 100m sin 444
    sleep 1
    if ! go test -v
    then
        padsp signalgen -t 100m sin 300
        continue
    fi
    padsp signalgen -t 100m sin 600
    git add -A
    git status
    grok commit | tee /tmp/$$
    cat /tmp/$$ | git commit -F-
    padsp signalgen -t 100m sin 900
done
