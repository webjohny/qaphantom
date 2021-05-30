#!/bin/bash

# shellcheck disable=SC2034
#IP_ADDR="IP"

cd ../
export GOOS=linux
go build .
export GOOS=windows

ssh -i "~/.ssh/id_rsa" "root@$IP_ADDR" "service qaphantom stop"
scp -r -i "~/.ssh/id_rsa" qaphantom "root@$IP_ADDR:/var/www/html"
ssh -i "~/.ssh/id_rsa" "root@$IP_ADDR" "service qaphantom restart"