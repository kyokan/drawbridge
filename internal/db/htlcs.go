package db

import (
	"database/sql"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type HTLCs interface {
	LastHTLCID(channelId [32]byte) (int, error)
}

type PostgresHTLCs struct {
	db *sql.DB
}

func (p *PostgresHTLCs) LastHTLCID(channelId [32]byte) (int, error) {
	cid := hexutil.Encode(channelId[:])

	var count int

	err := p.db.QueryRow("SELECT COUNT(*) as count FROM htlcs WHERE channel_id = ?", cid).Scan(&count)

	if err != nil {
		return -1, err
	}

	if count == 0 {
		return -1, nil
	}

	var id int

	err = p.db.QueryRow("SELECT MAX(id) AS id FROM htlcs WHERE channel_id = ?",
		hexutil.Encode(channelId[:])).Scan(&id)

	return id, err
}