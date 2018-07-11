package db

import (
	"github.com/ethereum/go-ethereum/common"
	"database/sql"
	"math/big"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type Channels interface {
	Save(channel *ETHChannel) error
	FindById(chanId common.Hash) (*ETHChannel, error)
	FindByAmount(amount *big.Int) (*ETHChannel, error)
}

type PostgresChannels struct {
	db *sql.DB
}

func (p *PostgresChannels) Save(channel *ETHChannel) error {
	return NewTransactor(p.db, func(tx *sql.Tx) error {
		_, err := tx.Exec(
			"INSERT INTO eth_channels (id, funding_output, counterparty) VALUES ($1, $2, $3)",
			channel.ID.Hex(),
			channel.FundingOutput.Hex(),
			channel.Counterparty.Hex(),
		)
		return err
	})
}

func (p *PostgresChannels) FindById(chanId common.Hash) (*ETHChannel, error) {
	row := p.db.QueryRow(`
		SELECT e.id, e.funding_output, e.counterparty FROM eth_channels e
		WHERE e.id = $1 
	`, chanId.Hex())
	return deserChannelRow(row)
}

func (p *PostgresChannels) FindByAmount(amount *big.Int) (*ETHChannel, error) {
	row := p.db.QueryRow(`
		SELECT e.id, e.funding_output, e.counterparty FROM eth_channels e
		JOIN eth_outputs o ON e.funding_output = o.id
		WHERE o.amount = $1 
	`, amount.Text(10))
	return deserChannelRow(row)
}

type rawChannel struct {
	ID            string
	FundingOutput string
	Counterparty  string
}

func deserChannelRow(row *sql.Row) (*ETHChannel, error) {
	raw := &rawChannel{}
	err := row.Scan(&raw.ID, &raw.FundingOutput, &raw.Counterparty)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var id common.Hash
	buf, err := hexutil.Decode(raw.ID)
	if err != nil {
		return nil, err
	}
	copy(id[:], buf)

	var fundingOutput common.Hash
	buf, err = hexutil.Decode(raw.FundingOutput)
	if err != nil {
		return nil, err
	}
	copy(fundingOutput[:], buf)

	counterparty := common.HexToAddress(raw.Counterparty)

	return &ETHChannel{
		ID:            id,
		FundingOutput: fundingOutput,
		Counterparty:  counterparty,
	}, nil
}
