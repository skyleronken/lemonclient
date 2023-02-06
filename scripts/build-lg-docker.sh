#!/bin/bash

echo "==== Building docker image ===="
docker build https://github.com/NationalSecurityAgency/lemongraph.git#lg-lite -t lg-lite:latest
