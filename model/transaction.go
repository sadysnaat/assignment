package model

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"math/big"
	"time"
)

type Transaction struct {
	From  common.Address `json:"from"`
	To    common.Address `json:"to"`
	Hash  common.Hash    `json:"hash"`
	Block *big.Int       `json:"block"`
	Value *big.Int       `json:"value"`
	Fee   uint64         `json:"fee"`
	Time  time.Time      `json:"time"`
	db    *sql.DB
}

func (tx *Transaction) TransactionsForAccount(address common.Address, limit, offset int, order, sortBy string) ([]*Transaction, error) {
	var txs []*Transaction
	query := `select
       t.to_addr,
       t.from_addr,
		t.hash,
       b.number,
       t.amount,
       t.fee,
	b. time
from transactions t join blocks b
on
 b.number = t.block
where t.to_addr = X'%s'
or t.from_addr = X'%s'
order by %s %s
limit %d offset %d`
	rows, err := tx.db.Query(fmt.Sprintf(query,
		tx.toBytes(address.Bytes()),
		tx.toBytes(address.Bytes()),
		sortBy,
		order,
		limit,
		offset))
	if err != nil {
		return txs, err
	}
	defer rows.Close()

	for rows.Next() {
		t := new(Transaction)
		var block int64
		var value float64
		var fee float64
		var ts mysql.NullTime
		rows.Scan(&t.To, &t.From, &t.Hash, &block, &value, &fee, &ts)

		t.Block = big.NewInt(block)
		t.Value, _ = big.NewFloat(value).Int(t.Value)
		t.Fee, _ = big.NewFloat(fee).Uint64()
		if ts.Valid {
			fmt.Println(ts.Time)
			t.Time = ts.Time
		} else {
			fmt.Println("wrong ts")
		}
		fmt.Println(t)
		txs = append(txs, t)
	}

	return txs, nil
}

func (tx *Transaction) WithDB(db *sql.DB) *Transaction {
	tx.db = db
	return tx
}

func (tx *Transaction) SaveToDB() {
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

func NewTransaction(tx *types.Transaction, txr *types.Receipt, db *sql.DB, b *big.Int) (*Transaction, error) {
	msg, err := tx.AsMessage(types.NewEIP155Signer(tx.ChainId()))
	if err != nil {
		fmt.Println(err)
	}

	var t *Transaction
	// if transaction is a contract creation then tx.To will be nil
	if tx.To() != nil {
		t = &Transaction{
			From:  *tx.To(),
			To:    msg.From(),
			Hash:  tx.Hash(),
			db:    db,
			Block: b,
			Value: tx.Value(),
			Fee:   txr.GasUsed,
		}
	} else {
		t = &Transaction{
			From:  common.Address{},
			To:    msg.From(),
			Hash:  tx.Hash(),
			db:    db,
			Block: b,
			Value: tx.Value(),
			Fee:   txr.GasUsed,
		}
	}

	return t, nil
}

func (tx *Transaction) toBytes(b []byte) string {
	return hex.EncodeToString(b)
}
