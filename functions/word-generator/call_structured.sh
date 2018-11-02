#!/usr/bin/env bash

set -xe

curl -v -X POST -H 'content-type: application/cloudevent+json' `fn inspect context | grep api | awk '{print $2}'`/t/cncf/word-generator-trigger -d @payload.json
