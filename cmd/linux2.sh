#!/bin/bash

# shellcheck disable=SC2034
#IP_ADDR="IP"

cd ../
export GOOS=linux
go build .
export GOOS=windows

scp -r -i "~/.ssh/id_rsa" qaphantom "root@$IP_ADDR:/var/www/html/qaphantom"