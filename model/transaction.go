package model

import (
	"database/sql"
	"github.com/ethereum/go-ethereum/common"
)

type Transaction struct {
	From common.Address
	To common.Address
	Hash common.Hash
	db *sql.DB
}
