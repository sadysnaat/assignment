package main

import (
	"flag"
	"fmt"
	"github.com/sadysnaat/assignment/indexer"
	"github.com/sadysnaat/assignment/server"
)

func main() {
	wssURL := flag.String("wss-url", "wss://kovan.infura.io/ws/v3/6c6f87a10e12438f8fbb7fc7c762b37c", "websocket url for the subscription")
	httpsURL := flag.String("https-url", "https://kovan.infura.io/v3/6c6f87a10e12438f8fbb7fc7c762b37c", "https url for indexer")
	dbURL := flag.String("db-url", "root:my-secret-pw@tcp(localhost:32769)/assignment", "database uri")
	flag.Parse()

	fmt.Println(*wssURL, *httpsURL, *dbURL)
	i, err := indexer.NewIndexer(*wssURL, *httpsURL, *dbURL)
	if err != nil {
		fmt.Println(err)
	}

	go i.Start()

	s := server.NewServer(*dbURL)
	err = s.Start()
	if err != nil {
		fmt.Println(err)
	}
}
