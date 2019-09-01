package model

import (
	"database/sql"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Transaction struct {
	From *common.Address
	To   common.Address
	Hash common.Hash
	db   *sql.DB
}

func (tx *Transaction) SaveToDB()  {
	fmt.Println("saving tx to db", tx.From.String(), tx.To.String(), tx.Hash.String())
}

func NewTransaction(tx *types.Transaction, db *sql.DB) (*Transaction, error) {
	msg, err := tx.AsMessage(types.NewEIP155Signer(tx.ChainId()))
	if err != nil {
		fmt.Println(err)
	}

	t := &Transaction{
		From: tx.To(),
		To: msg.From(),
		Hash: tx.Hash(),
		db:   db,
	}

	return t, nil
}
