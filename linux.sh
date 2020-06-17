#!/bin/bash

export GOOS=linux
go build
export GOOS=windows

echo Ghjcnjgfhjkm1!
scp qaphantom root@45.84.225.29:/var/www/html