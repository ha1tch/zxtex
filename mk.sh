#!/bin/bash
mkdir -p bin

echo "Building all binaries for zxtex..."
pushd bin
GOOS=windows GOARCH=amd64 go build -x -o zxtex.exe          zxtex.go
GOOS=windows GOARCH=386   go build -x -o zxtex.win32.exe    zxtex.go
GOOS=linux   GOARCH=amd64 go build -x -o zxtex.linux        zxtex.go
GOOS=linux   GOARCH=386   go build -x -o zxtex.linux32      zxtex.go
GOOS=linux   GOARCH=arm   go build -x -o zxtex.rpi          zxtex.go
GOOS=linux   GOARCH=arm64 go build -x -o zxtex.rpi64        zxtex.go
GOOS=darwin  GOARCH=arm64 go build -x -o zxtex.mac          zxtex.go
popd

