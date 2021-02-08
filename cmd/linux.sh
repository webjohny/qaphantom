#!/bin/bash

# shellcheck disable=SC2034
IP_ADDR="45.90.35.231"

cd ../
export GOOS=linux
go build .
export GOOS=windows

ssh -i "~/.ssh/id_rsa" "root@45.90.35.231" "service qaphantom stop"
scp -r -i "~/.ssh/id_rsa" qaphantom "root@45.90.35.231:/var/www/html"
ssh -i "~/.ssh/id_rsa" "root@45.90.35.231" "service qaphantom restart"