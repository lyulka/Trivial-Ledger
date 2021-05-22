package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/exec"

	"github.com/lyulka/trivial-ledger/server"
	structs "github.com/lyulka/trivial-ledger/structs"
	"go.etcd.io/etcd/clientv3"
)

func main() {

	fmt.Println("Deleting old binaries...")
	out, err := exec.Command("rm", "main", "tledger-evaluate").Output()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(out)

	fmt.Println("Installing tledger-server into $GOPATH/bin...")
	out, err = exec.Command("../install.sh").Output()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(out)

	fmt.Println("Clearing keys TLEDGER_PREFIX/1...1000 from etcd cluster for exclusive use of TLedger...")
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   server.DEFAULT_ENDPOINTS,
		DialTimeout: server.DEFAULT_DIAL_TIMEOUT,
	})
	ctx, cancel := context.WithTimeout(context.Background(), server.DEFAULT_DIAL_TIMEOUT)

	_, err = client.Delete(ctx, server.TLEDGER_PREFIX, clientv3.WithPrefix())
	cancel()
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx, cancel = context.WithTimeout(context.Background(), server.DEFAULT_DIAL_TIMEOUT)

	resp, err := client.Get(ctx, server.TLEDGER_PREFIX, clientv3.WithPrefix())
	cancel()
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(resp.Kvs) > 0 {
		fmt.Println(errors.New("Failed to clear keys TLEDGER_PREFIX/0...1000 from etcd cluster"))
		return
	}

	fmt.Println("Populating keys 1...1000 with transactions. Assuming a block size of 25, this will create 40 blocks")
	for i := 1; i < 1000; i += 1 {
		proposedTransaction := structs.ProposedTransaction{
			Content: fmt.Sprintf("This is transaction number: %d", i),
		}
		marshalledProposedTx, err := json.Marshal(proposedTransaction)
		if err != nil {
			fmt.Println(err)
			return
		}

		http.Post(server.DEFAULT_TLEDGER_SERVER_ENDPOINT,
			"application/json", bytes.NewBuffer(marshalledProposedTx))
	}
}
