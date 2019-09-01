package model

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

var (
	ErrBlockNotFound = errors.New("block not found in db")
	ErrBlockWithDifferentHash = errors.New("block at same height with different hash")
)

type Block struct {
	Hash   common.Hash
	Number *big.Int
	DB *sql.DB
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
		return b, nil
	}

	return b, ErrBlockNotFound
}

func (b *Block) Exists() bool {
	b.DB.Query("select * from blocks where number=X'%s'")
	return true
}

func (b *Block) HashBytes() string {
	return  hex.EncodeToString(b.Hash.Bytes())
}