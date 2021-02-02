#!/bin/bash

IP_ADDR="45.90.35.231"
export GOOS=linux
go build
export GOOS=windows

ssh -i "c:/Users/geryh/.ssh/id_rsa" "root@45.90.35.231" "service qaphantom stop"
scp -r -i "c:/Users/geryh/.ssh/id_rsa" qaphantom "root@45.90.35.231:/var/www/html"
ssh -i "c:/Users/geryh/.ssh/id_rsa" "root@45.90.35.231" "service qaphantom start"