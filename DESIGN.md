Package assignement 
```
.
├── API.md
├── DESIGN.md
├── Dockerfile
├── README.md
├── go.mod
├── go.sum
├── indexer 
│   └── indexer.go
├── main.go
├── model
│   ├── block.go
│   └── transaction.go
├── server
│   └── server.go
└── store
    ├── blocks.sql
    └── transactions.sql
```

main.go 

starts two services one Indexer and one Server(api server) 

Indexer 
* gets latest block and start downloading blocks from latest block to 0 and put them in header channel
* starts subscription to new head and put them in header channel
* starts downloader which reads from header channel and puts downloaded block to blocks channels
* starts persisting to database reads from blocks channel and writes to db
    * if the block is already present in db and hashes match if continues 
    * if the block is already present in db and hashes not match we stop and start reorg recovery
    * if the block is not present in db write block to db
    
in case of reorg
* stop the running subscriptions and pipelines
* delete the blocks at height higher than last correct block
* code in main.go handles restart of the

Scalability 
all the pipelines use channel to share the work we can start multiple instances of pipeline to share the work
multiple StartDownloading pipelines started 

```go
func (in *Indexer) Start(ctx context.Context, wg *sync.WaitGroup) {
	ctx, cancel := context.WithCancel(ctx)
	go in.StartSubscription(ctx)
	go in.StartHistory(ctx)
	go in.SaveBlock(ctx)
	go in.StartDownloading(ctx)
	go in.StartDownloading(ctx)
	go in.StartDownloading(ctx)
```   
