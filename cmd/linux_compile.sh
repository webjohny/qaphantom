#!/bin/bash

# shellcheck disable=SC2034
IP_ADDR="45.90.35.231"

cd ../
export GOOS=linux
go build .
export GOOS=windows