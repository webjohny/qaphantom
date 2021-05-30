#!/bin/bash

# shellcheck disable=SC2034
IP_ADDR="IP"

cd ../
export GOOS=linux
go build .
export GOOS=windows