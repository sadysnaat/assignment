package model

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

type Transaction struct {
	From common.Address
	To   common.Address
	Hash common.Hash
	Block *big.Int
	Value *big.Int
	Fee uint64
	db   *sql.DB
}

func (tx *Transaction) SaveToDB()  {
	fmt.Println("debug tx", tx)
	fmt.Println("saving tx to db", tx.From.String(), tx.To.String(), tx.Hash.String())
	txn, err := tx.db.Begin()
	if err != nil {
		fmt.Println(err)
	}

	_, err = txn.Exec(fmt.Sprintf("insert into transactions values (X'%s', X'%s', X'%s', %d, %d, %d)",
		tx.toBytes(tx.To.Bytes()),
		tx.toBytes(tx.From.Bytes()),
		tx.toBytes(tx.Hash.Bytes()),
		tx.Block,
		tx.Value,
		tx.Fee))

	if err != nil {
		fmt.Println(err)
	}

	err = txn.Commit()
	if err != nil {
		fmt.Println(err)
	}
}

func NewTransaction(tx *types.Transaction, db *sql.DB, b *big.Int) (*Transaction, error) {
	msg, err := tx.AsMessage(types.NewEIP155Signer(tx.ChainId()))
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("debug tx", tx.Value(), msg.Gas())
	var t *Transaction
	if tx.To() != nil {
		t = &Transaction{
			From: *tx.To(),
			To: msg.From(),
			Hash: tx.Hash(),
			db:   db,
			Block: b,
			Value: tx.Value(),
			Fee: tx.Gas() * tx.GasPrice().Uint64(),
		}
	} else {
		t = &Transaction{
			From: common.Address{},
			To: msg.From(),
			Hash: tx.Hash(),
			db:   db,
			Block: b,
			Value: tx.Value(),
			Fee: tx.Gas() * tx.GasPrice().Uint64(),
		}
	}


	return t, nil
}

func (tx *Transaction) toBytes(b []byte) string {
	return hex.EncodeToString(b)
}