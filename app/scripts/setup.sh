#!/usr/bin/env sh

# sync module dependcies
echo "Getting vendor module dependencies..."
go mod init event-data-pipeline
go mod tidy
go mod verify
echo "done"
echo

# build the binary
echo "Building the event-data-pipeline service..."
go build -tags musl -mod=vendor -o ./bin/event-data-pipeline .
echo "done"
echo
echo "...complete."