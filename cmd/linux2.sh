#!/bin/bash

# shellcheck disable=SC2034
#IP_ADDR="45.90.35.231"
#IP_ADDR="213.139.209.79"
#IP_ADDR="45.67.57.4"
#IP_ADDR="45.67.59.68"
IP_ADDR="45.141.78.99"
#IP_ADDR="45.141.76.83"
#IP_ADDR="45.141.76.84"
#IP_ADDR="193.200.74.67"
PASS="Ghjcnjgfhjkm1!"

cd ../
export GOOS=linux
go build .
export GOOS=windows

scp -r -i "~/.ssh/id_rsa" qaphantom "root@$IP_ADDR:/var/www/html/qaphantom"