#!/bin/bash

# shellcheck disable=SC2034
#IP_ADDR="45.90.35.231"
#IP_ADDR="213.139.209.79"
#IP_ADDR="45.67.57.4"
#IP_ADDR="45.67.59.68"
IP_ADDR="193.200.74.67"
PASS="Ghjcnjgfhjkm1!"

cd ../
export GOOS=linux
go build .
export GOOS=windows

ssh -i "~/.ssh/id_rsa" "root@$IP_ADDR" "service qaphantom stop"
scp -r -i "~/.ssh/id_rsa" qaphantom "root@$IP_ADDR:/var/www/html"
ssh -i "~/.ssh/id_rsa" "root@$IP_ADDR" "service qaphantom restart"