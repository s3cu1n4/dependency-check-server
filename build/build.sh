#!/bin/bash

project_path=$(cd `dirname $0`; pwd)
project_name="${project_path##*/}"

cd $project_path

# macos server
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -o darwin_server ../server/server.go 

# linux server
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o linux_server ../server/server.go 


# macos client
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -o darwin_client ../client/client.go

# linux client
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o linux_client ../client/client.go



echo "build successful"
