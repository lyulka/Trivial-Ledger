package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	color "github.com/TwinProduction/go-color"
	"github.com/lyulka/trivial-ledger/server"
	structs "github.com/lyulka/trivial-ledger/structs"
	"go.etcd.io/etcd/clientv3"
)

const SERVER1_ENDPOINT = "localhost:9090"
const SERVER2_ENDPOINT = "localhost:9091"

func main() {

	defer func() {
		fmt.Println(color.Red + "Please `killall -2 tledger-server` yourself" + color.Reset)
	}()

	fmt.Println(color.Red + "MAKE SURE YOU HAVE INSTALLED tledger-server by running ../install.sh" + color.Reset)
	fmt.Println()

	fmt.Println(color.Green + "[Clearing keys TLEDGER_PREFIX/1...1000 from etcd cluster for exclusive use of TLedger]" + color.Reset)
	fmt.Println()
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   server.DEFAULT_ENDPOINTS,
		DialTimeout: server.DEFAULT_DIAL_TIMEOUT,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), server.DEFAULT_DIAL_TIMEOUT)
	_, err = client.Delete(ctx, server.PREFIX, clientv3.WithPrefix())
	cancel()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Test if keys were cleared successfully
	ctx, cancel = context.WithTimeout(context.Background(), server.DEFAULT_DIAL_TIMEOUT)
	resp, err := client.Get(ctx, server.PREFIX, clientv3.WithPrefix())
	cancel()
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(resp.Kvs) > 0 {
		fmt.Println(errors.New("failed to clear keys TLEDGER_PREFIX/0...1000 from etcd cluster"))
		return
	}

	fmt.Printf(color.Yellow+"[Start SERVER1: %s]\n"+color.Reset, SERVER1_ENDPOINT)
	fmt.Println()

	err = exec.Command("tledger-server", SERVER1_ENDPOINT).Start()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf(color.Yellow+"[Start SERVER2: %s]\n"+color.Reset, SERVER2_ENDPOINT)
	fmt.Println()

	err = exec.Command("tledger-server", SERVER2_ENDPOINT).Start()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf(color.Yellow + "[Wait for 3 seconds for servers to start up]\n" + color.Reset)
	fmt.Println()
	time.Sleep(3 * time.Second)

	fmt.Println(color.Green + "[Use both servers concurrently to fill keys 1...1000 with transactions. Assuming a block size of 25, this will create 40 blocks]" + color.Reset)
	fmt.Println()

	done := make(chan bool)

	// Make calls to SERVER1
	go func() {
		for i := 1; i < 1000; i += 1 {
			time.Sleep(5 * time.Millisecond)

			proposedTransaction := structs.ProposedTransaction{
				Content: fmt.Sprintf("Thread 1: %d", i),
			}
			marshalledProposedTx, err := json.Marshal(proposedTransaction)
			if err != nil {
				fmt.Println(err)
				return
			}

			resp, err := http.Post("http://"+SERVER1_ENDPOINT+"/proposeTransaction",
				"application/json", bytes.NewBuffer(marshalledProposedTx))
			if err != nil {
				fmt.Println(err)
				return
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusAccepted {
				fmt.Printf("Populate keys: status code: %d\n", resp.StatusCode)
			} else {
				fmt.Println(color.Blue + "Thread 1: successful transaction commit" + color.Reset)
			}
		}

		done <- true
	}()

	// Make calls to SERVER2
	go func() {
		for i := 1; i < 1000; i += 1 {
			time.Sleep(5 * time.Millisecond)

			proposedTransaction := structs.ProposedTransaction{
				Content: fmt.Sprintf("Thread 2: %d", i),
			}
			marshalledProposedTx, err := json.Marshal(proposedTransaction)
			if err != nil {
				fmt.Println(err)
				return
			}

			resp, err := http.Post("http://"+SERVER2_ENDPOINT+"/proposeTransaction",
				"application/json", bytes.NewBuffer(marshalledProposedTx))
			if err != nil {
				fmt.Println(err)
				return
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusAccepted {
				fmt.Printf("Populate keys: status code: %d\n", resp.StatusCode)
			} else {
				fmt.Println(color.Blue + "Thread 2: successful transaction commit" + color.Reset)
			}
		}

		done <- true
	}()

	// Wait until both threads are finished
	<-done
	<-done
}
