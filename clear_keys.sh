echo "Deleting old binaries..."
rm main
rm tledger-server

echo "Building new binaries..."
go build main.go
mv main tledger-server

echo "Installing tledger-server into $GOPATH/bin..."
mv tledger-server $GOPATH/bin

echo "Clearing keys 0...1000 for exclusive use of TLedger..."
for i in {1...1000};
    do etcdctl del i

echo "Populating keys 0...1000 with transactions. Assuming a block size of 25, this will create 40 blocks"
