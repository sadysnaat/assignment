package main

import (
	"assignment/indexer"
	"assignment/server"
	"fmt"
)

func main() {
	i, err := indexer.NewIndexer()
	if err != nil {
		fmt.Println(err)
	}

	go i.StartSubscription()
	go i.StartHistory()
	go i.SaveBlock()
	go i.StartDownloading()

	s := server.NewServer()
	s.Start()

}
