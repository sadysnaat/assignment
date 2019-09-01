package indexer

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
)

type Indexer struct {
	c *ethclient.Client
	latestBlock *big.Int

}

func NewIndexer() (Indexer, error)  {
	i := Indexer{}

	c, err := ethclient.Dial("wss://kovan.infura.io/ws/v3/6c6f87a10e12438f8fbb7fc7c762b37c")
	if err != nil {
		return i, err
	}

	i.c = c

	cid, err := c.ChainID(context.Background())
	fmt.Println(cid)

	return i, nil
}

func (in *Indexer)Start() {
	cH := make(chan *types.Header)

	_, err := in.c.SubscribeNewHead(context.Background(), cH)
	if err != nil {
		fmt.Println(err)
	}

	for value := range cH {
		fmt.Println("got block", value.Number)
	}
}