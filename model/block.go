package model

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

var (
	ErrBlockNotFound          = errors.New("block not found in db")
	ErrBlockWithDifferentHash = errors.New("block at same height with different hash")
)

type Block struct {
	Hash   common.Hash
	Number *big.Int
	found  bool
	DB     *sql.DB
}

func GetBlockByNumber(n *big.Int, db *sql.DB) (*Block, error) {
	b := &Block{Number: n, DB: db, Hash: common.Hash{}}
	b, err := b.ReadFromDB()
	if err != nil {
		return b, err
	}
	return b, nil
}

func (b *Block) SaveToDB() {
	tx, err := b.DB.Begin()
	if err != nil {
		fmt.Println(err)
	}

	_, err = tx.Exec(fmt.Sprintf("insert into blocks values (%d, X'%s')", b.Number.Int64(), b.HashBytes()))
	if err != nil {
		fmt.Println(err)
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println(err)
	}
}

func (b *Block) ReadFromDB() (*Block, error) {
	var n int64
	var h []byte
	found := false
	if b.Number == nil {
		fmt.Println("got nil")
	}
	fmt.Println(b.Number.Int64())
	rows, err := b.DB.Query(fmt.Sprintf("select * from blocks where number=%d", b.Number.Int64()))
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&n, &h)
		if err != nil {
			fmt.Println(err)
		}
		found = true
	}

	if found {
		b.Hash = common.BytesToHash(h)
		b.found = true
		return b, nil
	}

	return b, ErrBlockNotFound
}

func (b *Block) Exists() bool {
	b.DB.Query("select * from blocks where number=X'%s'")
	return true
}

func (b *Block) HashBytes() string {
	return hex.EncodeToString(b.Hash.Bytes())
}

func (b *Block) FoundInDB() bool {
	if b != nil {
		return b.found
	}
	return false
}

func (b *Block) SaveTxsToDB(txs types.Transactions) {
	for _, tx := range txs {
		t, err := NewTransaction(tx, b.DB, b.Number)
		if err != nil {
			fmt.Println(err)
		}
		t.SaveToDB()
	}
}