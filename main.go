package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/sadysnaat/assignment/indexer"
	"github.com/sadysnaat/assignment/server"
	"os"
	"os/signal"
	"sync"
	"time"
)

func main() {
	wssURL := flag.String("wss-url", "wss://kovan.infura.io/ws/v3/6c6f87a10e12438f8fbb7fc7c762b37c", "websocket url for the subscription")
	httpsURL := flag.String("https-url", "https://kovan.infura.io/v3/6c6f87a10e12438f8fbb7fc7c762b37c", "https url for indexer")
	dbURL := flag.String("db-url", "root:my-secret-pw@tcp(localhost:32769)/assignment", "database uri")
	apiHost := flag.String("api-host", "0.0.0.0", "api host")
	apiPort := flag.String("api-port", "8081", "api port")
	flag.Parse()

	fmt.Println(*wssURL, *httpsURL, *dbURL)
	i, err := indexer.NewIndexer(*wssURL, *httpsURL, *dbURL)
	if err != nil {
		fmt.Println(err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// starts and keeps running the indexer as reorg cancels all the pipelines
	go func(ctx context.Context) {
		for {
			var wg sync.WaitGroup
			wg.Add(1)
			go i.Start(ctx, &wg)

			wg.Wait()
			fmt.Println("indexer has stopped")

			select {
			case <-ctx.Done():
				return
			}
		}
	}(ctx)

	s := server.NewServer(*dbURL)
	err = s.Start(*apiHost, *apiPort)
	if err != nil {
		fmt.Println(err)
	}

	// Wait for a SIGINT (perhaps triggered by user with CTRL-C)
	// Run cleanup when signal is received
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	select {
	case <-signalChan:
		fmt.Println("ctrl+c pressed")
		cancel()
		// wait for context cleanup
		time.Sleep(5 * time.Second)
	}
}
