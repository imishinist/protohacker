#!/bin/bash

target=$1

cd $target

GOOS=linux GOARCH=amd64 go build .
scp $target protohackers:

