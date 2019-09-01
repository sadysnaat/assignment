package model

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
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

func (b *Block) Exists() bool {
	b.DB.Query("select * from blocks where number=X'%s'")
	return true
}

func (b *Block) HashBytes() string {
	return  hex.EncodeToString(b.Hash.Bytes())
}