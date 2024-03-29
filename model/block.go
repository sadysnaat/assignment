package model

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-sql-driver/mysql"
	"math/big"
	"time"
)

var (
	ErrBlockNotFound          = errors.New("block not found in db")
	ErrBlockWithDifferentHash = errors.New("block at same height with different hash")
)

type Block struct {
	Hash   common.Hash
	Number *big.Int
	found  bool
	Time   time.Time
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

func DeleteHigherBlocks(n *big.Int, db *sql.DB) {
	tx, err := db.Begin()
	if err != nil {
		fmt.Println(err)
	} else {
		_, err := tx.Exec(fmt.Sprintf("delete from blocks where number >= %d", n.Int64()))
		if err != nil {
			fmt.Println(err)
		}

		err = tx.Commit()
		if err != nil {
			fmt.Println(err)
		}
		return
	}
}

func (b *Block) SaveToDB() {
	tx, err := b.DB.Begin()
	if err != nil {
		fmt.Println(err)
	}

	_, err = tx.Exec(fmt.Sprintf("insert into blocks values (%d, X'%s', '%s')",
		b.Number.Int64(),
		b.HashBytes(),
		b.Time.Format("2006-01-02 15:04:05")))
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
	var t mysql.NullTime
	found := false

	rows, err := b.DB.Query(fmt.Sprintf("select * from blocks where number=%d", b.Number.Int64()))
	if err != nil {
		fmt.Println(err, "in block")
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&n, &h, &t)
		if err != nil {
			fmt.Println(err)
		}
		found = true
	}

	if found {
		b.Hash = common.BytesToHash(h)
		if t.Valid {
			b.Time = t.Time
		}
		b.found = true
		return b, nil
	}

	return b, ErrBlockNotFound
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

func (b *Block) SaveTxsToDB(txs types.Transactions, txr []*types.Receipt, time time.Time) {
	for i, tx := range txs {
		t, err := NewTransaction(tx, txr[i], b.DB, b.Number)
		if err != nil {
			fmt.Println(err)
		}
		t.SaveToDB()
	}
}
