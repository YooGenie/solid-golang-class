#!/usr/bin/env sh

# sync module dependcies
echo "Getting vendor module dependencies..."
go mod init mock-data-generator
go mod tidy
go mod verify
echo "done"
echo

# build the binary
echo "Building the mock-data-generator service..."
go build -mod=vendor -o ./bin/mock-data-generator .
echo "done"
echo
echo "...complete."