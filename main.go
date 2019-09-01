package main

import (
	"assignment/indexer"
	"fmt"
)

func main() {
	i, err := indexer.NewIndexer()
	if err != nil {
		fmt.Println(err)
	}

	i.Start()
}
