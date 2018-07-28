package db

import (
	"database/sql"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"time"
	"math/big"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/drawbridge/internal/conv"
)

type Outputs interface {
	SavePoll(outputs *PolledOutputs, blockNum uint64) error
	LastPoll() (uint64, error)
	FindById(id uint64) (*ETHOutput, error)
	FindSpendableByOwnerAmount(owner common.Address, amount *big.Int) (*ETHOutput, error)
}

type PostgresOutputs struct {
	db *sql.DB
}

type PolledOutputs struct {
	New []*ETHOutput
	Spent []uint64
	Withdrawn []uint64
}

func (p *PostgresOutputs) SavePoll(outputs *PolledOutputs, blockNum uint64) error {
	if len(outputs.New) == 0 && len(outputs.Spent) == 0 && len(outputs.Withdrawn) == 0 {
		return nil
	}

	return NewTransactor(p.db, func(tx *sql.Tx) error {
		if len(outputs.New) > 0 {
			stmt, err := tx.Prepare(`
				INSERT INTO eth_outputs (
					id,
					contract_address,
					amount,
					block_number,
					tx_hash,
					script,
					type,
					spent,
					withdrawn
				) VALUES (
					$1, $2, $3, $4, $5, $6, $7, $8, $9
				)
			`)
			if err != nil {
				return err
			}

			for _, out := range outputs.New {
				_, err := stmt.Exec(
					conv.BigToHex(out.ID),
					out.ContractAddress.Hex(),
					out.Amount.Text(10),
					blockNum,
					out.TxHash.Hex(),
					hexutil.Encode(out.Script),
					out.Type,
					false,
					false,
				)

				if err != nil {
					return err
				}
			}
		}

		if len(outputs.Spent) > 0 {
			stmt, err := tx.Prepare("UPDATE eth_outputs SET spent = $1 WHERE id = $2")
			if err != nil {
				return err
			}

			for _, id := range outputs.Spent {
				stmt.Exec(true, id)
			}
		}

		if len(outputs.Withdrawn) > 0 {
			stmt, err := tx.Prepare("UPDATE eth_outputs SET withdrawn = $1 WHERE id = $2")
			if err != nil {
				return err
			}

			for _, id := range outputs.Withdrawn {
				stmt.Exec(true, id)
			}
		}

		_, err := tx.Exec(
			"UPDATE eth_chainsaw_status SET (last_seen_block, last_polled_at) = ($1, $2)",
			blockNum,
			time.Now().Unix(),
		)
		if err != nil {
			return err
		}

		return nil
	})
}

func (p *PostgresOutputs) LastPoll() (uint64, error) {
	row := p.db.QueryRow("SELECT MAX(last_seen_block) FROM eth_chainsaw_status;")
	var blockNum uint64
	err := row.Scan(&blockNum)
	if err != nil {
		return 0, err
	}

	return blockNum, nil
}

func (p *PostgresOutputs) FindById(id uint64) (*ETHOutput, error) {
	row := p.db.QueryRow(`
		SELECT id, contract_address, amount, block_number, tx_hash, script, type, spent, withdrawn 
			FROM eth_outputs WHERE id = $1
	`, id)

	return deserOutputRow(row)
}

func (p *PostgresOutputs) FindSpendableByOwnerAmount(owner common.Address, amount *big.Int) (*ETHOutput, error) {
	row := p.db.QueryRow(`
		SELECT id, contract_address, amount, block_number, tx_hash, script, type, spent, withdrawn 
			FROM eth_outputs WHERE type = 1 AND script = $1 AND value = $3;
	`, owner.Hex(), amount.Text(10))
	return deserOutputRow(row)
}

type rawOutput struct {
	ID              string
	ContractAddress string
	Amount          string
	BlockNumber     uint64
	TxHash          string
	Script          string
	Type            uint8
	IsWithdrawn     bool
	IsSpent         bool
}

func deserOutputRow(row *sql.Row) (*ETHOutput, error) {
	raw := &rawOutput{}
	err := row.Scan(&raw.ID, &raw.ContractAddress, &raw.Amount, &raw.BlockNumber,
		&raw.TxHash, &raw.Script, &raw.Type, &raw.IsWithdrawn, &raw.IsSpent)
	if err != nil {
		return nil, err
	}

	id, err := conv.HexToBig(raw.ID)
	if err != nil {
		return nil, err
	}

	value, err := conv.StringToBig(raw.Amount)
	if err != nil {
		return nil, err
	}

	var txHash common.Hash
	buf, err := hexutil.Decode(raw.TxHash)
	if err != nil {
		return nil, err
	}
	copy(txHash[:], buf)

	script, err := hexutil.Decode(raw.Script)
	if err != nil {
		return nil, err
	}

	return &ETHOutput{
		ID:              id,
		ContractAddress: common.HexToAddress(raw.ContractAddress),
		Amount:          value,
		BlockNumber:     raw.BlockNumber,
		TxHash:          txHash,
		Script:          script,
		Type:            raw.Type,
		IsSpent:         raw.IsSpent,
		IsWithdrawn:     raw.IsWithdrawn,
	}, nil
}