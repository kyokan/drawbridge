package db

import (
	"github.com/kyokan/drawbridge/pkg/types"
	"database/sql"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"bytes"
	"time"
	"math/big"
)

type UTXOs interface {
	SavePoll(utxos []*types.EthUTXO, blockNum *big.Int) error
	LastPoll() (*big.Int, error)
	FindById(id [32]byte) *types.EthUTXO
}

type PostgresUTXOs struct {
	db *sql.DB
}

func (p *PostgresUTXOs) SavePoll(utxos []*types.EthUTXO, blockNum *big.Int) error {
	if len(utxos) == 0 {
		return nil
	}

	return NewTransactor(p.db, func(tx *sql.Tx) error {
		insert, err := tx.Prepare("INSERT INTO eth_utxos(id, input_id, value, owner, block_number, tx_hash, spent, withdrawn) " +
			"VALUES($1, $2, $3, $4, $5, $6, $7, $8)")

		if err != nil {
			return err
		}

		updateSpent, err := tx.Prepare("UPDATE eth_utxos SET spent = $1 WHERE id = $2")

		if err != nil {
			return err
		}

		updateWithdrawn, err := tx.Prepare("UPDATE eth_utxos SET withdrawn = $1 WHERE id = $2")

		if err != nil {
			return err
		}

		for _, utxo := range utxos {
			if utxo.IsWithdrawn {
				_, err = updateWithdrawn.Exec(true, hexutil.Encode(utxo.ID[:]))
			} else {
				_, err = insert.Exec(
					hexutil.Encode(utxo.ID[:]),
					hexutil.Encode(utxo.InputID[:]),
					utxo.Value.Text(10),
					utxo.Owner.Hex(),
					utxo.BlockNumber.Text(10),
					utxo.TxHash.Hex(),
					utxo.IsSpent,
					utxo.IsWithdrawn,
				)

				if err == nil && !bytes.Equal(utxo.InputID[:], types.ZeroUTXOID[:]) {
					_, err = updateSpent.Exec(true, utxo.InputID)
				}
			}

			if err != nil {
				return err
			}
		}

		_, err = tx.Exec(
			"UPDATE eth_chainsaw_status SET (last_seen_block, last_polled_at) = ($1, $2)",
			blockNum.Text(10),
			time.Now().Unix(),
		)

		if err != nil {
			return err
		}

		return nil
	})
}

func (p *PostgresUTXOs) LastPoll() (*big.Int, error) {
	row := p.db.QueryRow("SELECT last_seen_block FROM eth_chainsaw_status LIMIT 1")
	var blockNum int64
	err := row.Scan(&blockNum)

	if err != nil {
		return nil, err
	}

	return big.NewInt(blockNum), nil
}

func (p *PostgresUTXOs) FindById(id [32]byte) *types.EthUTXO {
	return nil
}
