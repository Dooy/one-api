#!/bin/bash

set -e 
go build -ldflags "-s -w" -o one-api &&  ./one-api --port 3008 --log-dir ./logs