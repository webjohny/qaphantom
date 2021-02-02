#!/bin/bash

export GOOS=linux
go build
export GOOS=windows

echo Ghjcnjgfhjkm1!
scp install.sh root@213.139.209.79:/var/www/html