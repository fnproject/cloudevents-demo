#!/usr/bin/env bash

set -ex

curl -v -X POST -H 'ce-cloudeventsversion: 0.1' \
    -H 'ce-eventtime: 2018-10-23T12:28:22.4579346Z' \
    -H 'ce-eventid: 96fb5f0b-001e-0108-6dfe-da6e2806f124' \
    -H 'ce-source: http://srcdog.com/cedemo' \
    -H 'ce-eventtype: word.found.name' \
    -H 'content-type: application/json' `fn inspect context | grep api | awk '{print $2}'`/t/cncf/word-generator-trigger
