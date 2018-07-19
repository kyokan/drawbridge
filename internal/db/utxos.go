package db

import (
	"github.com/kyokan/drawbridge/pkg/types"
	"database/sql"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"time"
	"math/big"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/drawbridge/internal/conv"
)

type UTXOs interface {
	SavePoll(utxos []*types.EthUTXO, blockNum *big.Int) error
	LastPoll() (*big.Int, error)
	FindById(id [32]byte) *types.EthUTXO
	FindSpendableByOwnerAmount(amount *big.Int, owner common.Address) (*types.EthUTXO, error)
}

type PostgresUTXOs struct {
	db *sql.DB
}

type rawUtxo struct {
	Owner       string
	Value       string
	BlockNumber string
	TxHash      string
	ID          string
	InputID     string
	IsWithdrawn bool
	IsSpent     bool
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

				if err == nil && utxo.InputID != types.ZeroUTXOID {
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

func (p *PostgresUTXOs) FindSpendableByOwnerAmount(amount *big.Int, owner common.Address) (*types.EthUTXO, error) {
	raw := &rawUtxo{}

	row := p.db.QueryRow(
		"SELECT id, owner, value, block_number, tx_hash, input_id, withdrawn, spent FROM eth_utxos WHERE owner = $1 AND value = $2",
		owner.Hex(),
		amount.Text(10),
	)
	err := row.Scan(&raw.ID, &raw.Owner, &raw.Value, &raw.BlockNumber, &raw.TxHash, &raw.InputID, &raw.IsWithdrawn, &raw.IsSpent)

	if err != nil {
		return nil, err
	}

	value, err := conv.StringToBig(raw.Value)

	if err != nil {
		return nil, err
	}

	blockNumber, err := conv.StringToBig(raw.BlockNumber)

	if err != nil {
		return nil, err
	}

	txHash, err := conv.HexToBytes32(raw.TxHash)

	if err != nil {
		return nil, err
	}

	id, err := conv.HexToBytes32(raw.ID)

	if err != nil {
		return nil, err
	}

	inputId, err := conv.HexToBytes32(raw.InputID)

	if err != nil {
		return nil, err
	}

	utxo := &types.EthUTXO{
		Owner:       common.HexToAddress(raw.Owner),
		Value:       value,
		BlockNumber: blockNumber,
		TxHash:      txHash,
		ID:          id,
		InputID:     inputId,
		IsWithdrawn: raw.IsWithdrawn,
		IsSpent:     raw.IsSpent,
	}

	return utxo, nil
}
