package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/lyulka/trivial-ledger/server"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("To specify server endpoint: `tledger-server {localhost:xxxx}`")
		fmt.Println("No endpoint specified. Defaulting to: " + server.TLEDGER_SERVER_ENDPOINT)
	}

	if len(os.Args) == 2 {
		server.TLEDGER_SERVER_ENDPOINT = os.Args[1]
		fmt.Println("Server endpoint: " + server.TLEDGER_SERVER_ENDPOINT)
	}

	fmt.Println("Trivial Ledger: Server starting")

	s, err := server.New()
	if err != nil {
		fmt.Println(err)
	}

	// Teardown server when receive an interrupt
	// signal (for example, when user inputs Ctrl+C
	// in terminal)
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan

		s.Teardown()

		os.Exit(0)
	}()

	if err := http.ListenAndServe(server.TLEDGER_SERVER_ENDPOINT, s.Router); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
