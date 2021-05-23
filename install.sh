#!/bin/sh
echo "Remove old binaries..."
rm main
rm tledger-server

echo "Creating new binaries..."
go build main.go
mv main tledger-server

echo "moving tledger-server to $GOPATH/bin..."
mv tledger-server $GOPATH/bin

return 0