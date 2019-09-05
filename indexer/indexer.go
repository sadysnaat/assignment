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
	"sync"
	"time"
)

var (
	one  = big.NewInt(1)
	zero = big.NewInt(0)
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

	sig chan struct{}
	db  *sql.DB
}

func NewIndexer(wssURL, httpsURL, dbURL string) (Indexer, error) {
	ch := make(chan *types.Header, 100)
	cb := make(chan *types.Block, 100)
	sig := make(chan struct{}, 1)

	i := Indexer{}

	c, err := ethclient.Dial(wssURL)
	if err != nil {
		return i, err
	}

	i.c = c
	i.headers = ch
	i.blocks = cb
	i.sig = sig

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

func (in *Indexer) StartSubscription(ctx context.Context) {
	fmt.Println("starting subscription for new head")
	cH := make(chan *types.Header)

	_, err := in.c.SubscribeNewHead(context.Background(), cH)
	if err != nil {
		fmt.Println(err)
	}

	// Here we receive updates on the channel cH we iterate through the headers arrived
	// and pass them to headers queue
	for {
		select {
		case value := <-cH:
			select {
			case in.headers <- value:
				fmt.Println("received new block", value.Number)
			case <-ctx.Done():
				fmt.Println("stoppping block subscription")
				return
			}
		case <-ctx.Done():
			fmt.Println("stoppping block subscription")
			return
		}
	}
}

// This function starts queuing headers from latest block received to towards zero
// It keeps enqueuing smaller block numbers
func (in *Indexer) StartHistory(ctx context.Context) {
	i := in.latestBlock

	for {
		if i.Cmp(zero) < 0 {
			break
		}

		h, err := in.c.HeaderByNumber(ctx, i)
		if err != nil {
			fmt.Println(err)
			if ctx.Err() != nil {
				return
			} else {
				continue
			}
		}

		// write to headers is blocking not to overwhelm the
		// the headers buffers
		select {
		case in.headers <- h:
		case <-ctx.Done():
			fmt.Println("stopping history")
			return
		}

		i = big.NewInt(0).Sub(i, one)
	}
}

func (in *Indexer) StartHistoryFrom(ctx context.Context, n *big.Int) {
	latest, err := in.chs.HeaderByNumber(ctx, nil)
	if err != nil {
		if ctx.Err() != nil {
			return
		}
	} else {
		l := latest.Number
		in.latestBlock = l
		for {
			if l.Cmp(n) >= 0 {
				break
			}
			h, err := in.c.HeaderByNumber(ctx, n)
			if err != nil {
				fmt.Println(err)
				if ctx.Err() != nil {
					return
				} else {
					continue
				}
			}

			select {
			case in.headers <- h:
			case <-ctx.Done():
				fmt.Println("stopping history recovery")
				return
			}

			n = big.NewInt(0).Add(n, one)
		}
	}
}

func (in *Indexer) StartDownloading(ctx context.Context) {
	fmt.Println("starting downloader")
	for {
		select {
		case header := <-in.headers:
			b, err := in.chs.BlockByNumber(ctx, header.Number)
			if err != nil {
				if ctx.Err() != nil {
					fmt.Println(ctx.Err())
					fmt.Println("cancelling downloader")
					return
				}
				fmt.Println("couldn't find block in canonical chain", header.Number, header.Hash().String())
				select {
				case in.headers <- header:
					fmt.Println("rescheduled block", header.Number)
				case <-ctx.Done():
					fmt.Println("cancelling downloader")
					return
				case <-time.After(5 * time.Second):
					// give up after 5 seconds to avoid deadlock on  headers write
				}

				// If we encounter an error while downloading the block we reschedule
				// the block to headers done above. And continue
				continue
			}

			fmt.Println("scheduled downloaded block for indexing", b.Number())

			// If we have found the block we publish to blocks queue
			select {
			case in.blocks <- b:
			case <-ctx.Done():
				fmt.Println("cancelling downloader")
				return
			}
		case <-ctx.Done():
			fmt.Println("cancelling downloader")
			return
		}
	}
}

func (in *Indexer) SaveBlock(ctx context.Context) {
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
					go in.Stop(ctx, block.Number())
					return
				}
			} else {
				ts := time.Now()
				b.Hash = block.Hash()
				b.Time = time.Unix(int64(block.Time()), 0).UTC()
				b.SaveToDB()
				txs := block.Transactions()
				// block contains no transactions if better to skip the loop
				// no more work to do here
				if len(txs) == 0 {
					continue
				}
				// get the transaction receipt as the data of GasUsed is in the
				// TransactionReceipt Object.
				var txr []*types.Receipt
				for _, tx := range txs {
					recpt, err := in.chs.TransactionReceipt(context.Background(), tx.Hash())
					if err != nil {
						continue
					}
					txr = append(txr, recpt)
				}
				// we couldn't fetch as many receipts as my transactions rescheduling block
				// for indexing
				if len(txr) != len(txs) {
					go in.RescheduleBlock(block)
					continue
				}
				b.SaveTxsToDB(txs, txr, block.ReceivedAt)
				fmt.Println("saved block", block.Number(), "to db", time.Since(ts))
			}
		case <-ctx.Done():
			fmt.Println("stopping index to db")
			return
		}
	}
}

func (in *Indexer) Start(ctx context.Context, wg *sync.WaitGroup) {
	ctx, cancel := context.WithCancel(ctx)
	go in.StartSubscription(ctx)
	go in.StartHistory(ctx)
	go in.SaveBlock(ctx)
	go in.StartDownloading(ctx)
	go in.StartDownloading(ctx)
	go in.StartDownloading(ctx)

	go func(ctx context.Context, cancelFunc context.CancelFunc) {
		select {
		case <-ctx.Done():
			fmt.Println("cancelling start")
			wg.Done()
			return
		case <-in.sig:
			fmt.Println("cancelling via reorg")
			//cancel()
			cancelFunc()
			wg.Done()
			fmt.Println("returning back")
			return
		}
	}(ctx, cancel)
}

func (in *Indexer) RescheduleBlock(b *types.Block) {
	fmt.Println("rescheduled block", b.Number())
	in.blocks <- b
}

func (in *Indexer) Stop(ctx context.Context, n *big.Int) {
	ctxwc, cancel := context.WithCancel(ctx)
	context.Background()
	for {
		select {
		case in.sig <- struct{}{}:
			fmt.Println("stopping the indexer")
			in.RecoverFromBlock(ctxwc, n)
			return
		case <-ctx.Done():
			cancel()
			return
		}
	}
}

func (in *Indexer) RecoverFromBlock(ctx context.Context, n *big.Int) {
	fmt.Println("starting recovery")
	ni := big.NewInt(0).Set(n)
	for {
		ni = big.NewInt(0).Sub(ni, one)
		if ni.Cmp(zero) < 0 {
			break
		}

		b, err := model.GetBlockByNumber(ni, in.db)
		if err != nil {
			continue
		}
		if b.FoundInDB() {
			h, err := in.chs.HeaderByNumber(ctx, ni)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
			} else {
				if h.Hash() == b.Hash {
					break
				}
			}
		}
	}

	// Here we have got the least correct block
	// we start recovering

	// first we delete higher blocks and associated transactions
	model.DeleteHigherBlocks(ni, in.db)

	// recover blocks between latest correct block and latest block
	in.StartHistoryFrom(ctx, ni)

	select {
	case <-ctx.Done():
		fmt.Println("recovery cancelled")
		return
	}
}
