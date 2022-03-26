#!/bin/bash

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o htm-node
tar czf htm-node.tgz htm-node
cp htm-node.tgz ansible/files/downloads/
rm -rf htm-node.tgz htm-node
