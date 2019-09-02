package indexer

import (
	"assignment/model"
	"context"
	"database/sql"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/go-sql-driver/mysql"
	"math/big"
	"time"
)

type Indexer struct {
	c           *ethclient.Client
	chs         *ethclient.Client
	latestBlock *big.Int
	// headers channel serves as queue for
	headers     chan *types.Header
	blocks      chan *types.Block
	db          *sql.DB
}

func NewIndexer() (Indexer, error) {
	ch := make(chan *types.Header, 10)
	cb := make(chan *types.Block, 10)

	i := Indexer{}

	c, err := ethclient.Dial("wss://kovan.infura.io/ws/v3/6c6f87a10e12438f8fbb7fc7c762b37c")
	if err != nil {
		return i, err
	}

	i.c = c
	i.headers = ch
	i.blocks = cb

	chS, err := ethclient.Dial("https://kovan.infura.io/v3/6c6f87a10e12438f8fbb7fc7c762b37c")

	i.chs = chS

	cid, err := c.ChainID(context.Background())
	fmt.Println(cid)

	l, err := c.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return i, err
	}

	i.latestBlock = l.Number

	db, err := sql.Open("mysql", "root:my-secret-pw@tcp(localhost:32768)/assignment")
	i.db = db
	return i, nil
}

func (in *Indexer) Start() {
	cH := make(chan *types.Header)

	_, err := in.c.SubscribeNewHead(context.Background(), cH)
	if err != nil {
		fmt.Println(err)
	}

	for value := range cH {

		select {
		case in.headers <- value:
			fmt.Println("got block", value.Number)
		}

	}
}

func (in *Indexer) StartHistory() {
	i := in.latestBlock
	one := big.NewInt(1)
	zero := big.NewInt(0)
	for {
		fmt.Println("history loop", i)
		if i.Cmp(zero) < 0 {
			break
		}

		h, err := in.c.HeaderByNumber(context.Background(), i)
		if err != nil {
			fmt.Println(err)
			continue
		}

		in.headers <- h

		i = big.NewInt(0).Sub(i, one)
	}
}

func (in *Indexer) StartDownloading() {
	fmt.Println("starting downloader")
	for header := range in.headers {
		fmt.Println("downloading block", header.Number)
		b, err := in.chs.BlockByNumber(context.Background(), header.Number)
		if err != nil {
			fmt.Println(err, header.Number, len(in.headers), len(in.blocks))
			select {
			case in.headers <- header:
				fmt.Println("rescheduled block", header.Number)
			}

			continue
		}

		fmt.Println("block found", b.Number())
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
				fmt.Println(err)
			}

			if b.FoundInDB() {
				if b.Hash == block.Hash() {
					continue
				} else {

				}
			} else {
				fmt.Println(b)
				b.Hash = block.Hash()
				b.Time = time.Unix(int64(block.Time()), 0)
				b.SaveToDB()
				txs := block.Transactions()
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
