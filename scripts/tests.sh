#!/bin/bash

export LC_ALL=C
time (
curl -s "${LG_SERVICE:-http://localhost:8000}/graph" | jq -r '.[].graph' | \
	xargs -r printf "${LG_SERVICE:-http://localhost:8000}/graph/%s\n" | \
	xargs -r -n50 curl -XDELETE
sync
)

go test -v ./...