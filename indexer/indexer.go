package indexer

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sadysnaat/assignment/model"
	"math/big"
	"time"
)

type Indexer struct {
	c           *ethclient.Client
	chs         *ethclient.Client
	latestBlock *big.Int
	// headers channel serves as queue for incoming headers
	// from history queue and newHeads
	headers chan *types.Header

	// headers channel contain downloaded blocks
	blocks chan *types.Block
	db     *sql.DB
}

func NewIndexer(wssURL, httpsURL, dbURL string) (Indexer, error) {
	ch := make(chan *types.Header, 10)
	cb := make(chan *types.Block, 10)

	i := Indexer{}

	c, err := ethclient.Dial(wssURL)
	if err != nil {
		return i, err
	}

	i.c = c
	i.headers = ch
	i.blocks = cb

	chS, err := ethclient.Dial(httpsURL)

	i.chs = chS

	cid, err := c.ChainID(context.Background())
	fmt.Println(cid)

	l, err := c.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return i, err
	}

	i.latestBlock = l.Number

	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		panic("could not connect to database")
	}
	i.db = db
	return i, nil
}

func (in *Indexer) StartSubscription() {
	fmt.Println("starting subscription for new head")
	cH := make(chan *types.Header)

	_, err := in.c.SubscribeNewHead(context.Background(), cH)
	if err != nil {
		fmt.Println(err)
	}

	// Here we receive updates on the channel cH we iterate through the headers arrived
	// and pass them to headers queue
	for value := range cH {
		select {
		case in.headers <- value:
			fmt.Println("received new block", value.Number)
		}
	}
}

// This function starts queuing headers from latest block received to towards zero
// It keeps enqueuing smaller block numbers
func (in *Indexer) StartHistory() {
	i := in.latestBlock
	one := big.NewInt(1)
	zero := big.NewInt(0)
	for {
		if i.Cmp(zero) < 0 {
			break
		}

		h, err := in.c.HeaderByNumber(context.Background(), i)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// write to headers is blocking not to overwhelm the
		// the headers buffers
		in.headers <- h

		i = big.NewInt(0).Sub(i, one)
	}
}

func (in *Indexer) StartDownloading() {
	fmt.Println("starting downloader")
	for header := range in.headers {
		b, err := in.chs.BlockByNumber(context.Background(), header.Number)
		if err != nil {
			fmt.Println("couldn't find block in canonical chain", header.Number, header.Hash().String())
			select {
			case in.headers <- header:
				fmt.Println("rescheduled block", header.Number)
			}

			// If we encounter an error while downloading the block we reschedule
			// the block to headers done above. And continue
			continue
		}

		fmt.Println("scheduled downloaded block for indexing", b.Number())

		// If we have found the block we publish to blocks queue
		in.blocks <- b
	}
}

func (in *Indexer) SaveBlock() {
	fmt.Println("starting index to db")
	for {
		select {
		case block := <-in.blocks:
			// upon receiving a block we must first check
			// if we have the block at given height in db
			// if yes we have two possible outcomes
			// 1. Hash of the block in DB matches the Hash of the block we received
			// in this case we discard the message as block is already synced.
			// 2. Hash of the block doesn't match the Hash of the block we received
			// in this case it means that reorg or fork has happened

			b, err := model.GetBlockByNumber(block.Number(), in.db)
			if err != nil {
				//fmt.Println(err)
			}

			if b.FoundInDB() {
				if b.Hash == block.Hash() {
					continue
				} else {
					// If we have reached here this means an reorg has happened or
					// our data in DB does not the data available in blockchain
					// time to resolve the reorg.
					// TODO: deepak implement reorg recovery
				}
			} else {
				b.Hash = block.Hash()
				b.Time = time.Unix(int64(block.Time()), 0).UTC()
				b.SaveToDB()
				txs := block.Transactions()
				// get the transaction receipt as the data of GasUsed is in the
				// TransactionReceipt Object.
				var txr []*types.Receipt
				block.Time()
				for _, tx := range txs {
					recpt, err := in.chs.TransactionReceipt(context.Background(), tx.Hash())
					if err != nil {
						continue
					}
					txr = append(txr, recpt)
				}
				b.SaveTxsToDB(txs, txr, block.ReceivedAt)
			}
		}
	}
}

func (in *Indexer) Start() {
	go in.StartSubscription()
	go in.StartHistory()
	go in.SaveBlock()
	go in.StartDownloading()
}
