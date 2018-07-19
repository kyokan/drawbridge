package db

import (
	"database/sql"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/kyokan/drawbridge/internal/conv"
	"strconv"
	"github.com/kyokan/drawbridge/pkg/types"
	"github.com/kyokan/drawbridge/pkg/crypto"
)

type Channels interface {
	CreateLocalChannel(msg *lnwire.OpenChannel, commitmentSeed []byte) error
	CreateRemoteChannel(open *lnwire.OpenChannel, accept *lnwire.AcceptChannel, commitmentSeed []byte) error
	AcceptLocalChannel(accept *lnwire.AcceptChannel) error
	GetPendingChannel(pendingChannelId [32]byte) (*types.Channel, error)
	FinalizeChannelId(pendingChannelId [32]byte, finalizedChannelId [32]byte, inputId []byte) error
	FinalizeChannelSignatures(finalizedChannelId [32]byte, ourSig crypto.Signature, theirSig crypto.Signature) error
	GetFinalizedChannel(finalizedChannelId [32]byte) (*types.Channel, error)
}

type PostgresChannels struct {
	db *sql.DB
}

type rawChannel struct {
	OurFundingAddress   string
	TheirFundingAddress string
	FundingAmount       string
	OurSignature        string
	TheirSignature      string
	InputID             string
}

func (p *PostgresChannels) CreateLocalChannel(msg *lnwire.OpenChannel, commitmentSeed []byte) error {
	return NewTransactor(p.db, func(tx *sql.Tx) error {
		insert, err := tx.Prepare(`
			INSERT INTO channels (
				chain_id,
				temporary_id,
				funding_amount,
				push_amount,
				dust_limit,
				max_value_in_flight,
				channel_reserve,
				htlc_minimum,
				fee_per_kw,
				csv_delay,
				our_funding_key,
				our_revocation_point,
				our_payment_point,
				our_delayed_payment_point,
				our_htlc_point,
				our_commitment_seed,
				self_originated				
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		`)

		if err != nil {
			return err
		}

		_, err = insert.Exec(
			hexutil.Encode(msg.ChainHash[:]),
			hexutil.Encode(msg.PendingChannelID[:]),
			msg.FundingAmount,
			strconv.FormatUint(uint64(msg.PushAmount), 10),
			msg.DustLimit,
			strconv.FormatUint(uint64(msg.MaxValueInFlight), 10),
			msg.ChannelReserve,
			strconv.FormatUint(uint64(msg.HtlcMinimum), 10),
			msg.FeePerKiloWeight,
			msg.CsvDelay,
			crypto.BTCECToCompressedHex(msg.FundingKey),
			crypto.BTCECToCompressedHex(msg.RevocationPoint),
			crypto.BTCECToCompressedHex(msg.PaymentPoint),
			crypto.BTCECToCompressedHex(msg.DelayedPaymentPoint),
			crypto.BTCECToCompressedHex(msg.HtlcPoint),
			hexutil.Encode(commitmentSeed),
			true,
		)

		return err
	})
}

func (p *PostgresChannels) CreateRemoteChannel(open *lnwire.OpenChannel, accept *lnwire.AcceptChannel, commitmentSeed []byte) error {
	return NewTransactor(p.db, func(tx *sql.Tx) error {
		insert, err := tx.Prepare(`
			INSERT INTO channels (
				chain_id,
				temporary_id,
				funding_amount,
				push_amount,
				dust_limit,
				max_value_in_flight,
				channel_reserve,
				htlc_minimum,
				fee_per_kw,
				csv_delay,
				our_funding_key,
				our_revocation_point,
				our_payment_point,
				our_delayed_payment_point,
				our_htlc_point,
				their_funding_key,
				their_revocation_point,
				their_payment_point,
				their_delayed_payment_point,
				their_htlc_point,
				their_first_commitment_point,
				our_commitment_seed,
				self_originated				
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)
		`)

		if err != nil {
			return err
		}

		_, err = insert.Exec(
			hexutil.Encode(open.ChainHash[:]),
			hexutil.Encode(accept.PendingChannelID[:]),
			open.FundingAmount,
			strconv.FormatUint(uint64(open.PushAmount), 10),
			accept.DustLimit,
			strconv.FormatUint(uint64(accept.MaxValueInFlight), 10),
			accept.ChannelReserve,
			strconv.FormatUint(uint64(accept.HtlcMinimum), 10),
			open.FeePerKiloWeight,
			accept.CsvDelay,
			crypto.BTCECToCompressedHex(accept.FundingKey),
			crypto.BTCECToCompressedHex(accept.RevocationPoint),
			crypto.BTCECToCompressedHex(accept.PaymentPoint),
			crypto.BTCECToCompressedHex(accept.DelayedPaymentPoint),
			crypto.BTCECToCompressedHex(accept.HtlcPoint),
			crypto.BTCECToCompressedHex(open.FundingKey),
			crypto.BTCECToCompressedHex(open.RevocationPoint),
			crypto.BTCECToCompressedHex(open.PaymentPoint),
			crypto.BTCECToCompressedHex(open.DelayedPaymentPoint),
			crypto.BTCECToCompressedHex(open.HtlcPoint),
			crypto.BTCECToCompressedHex(open.FirstCommitmentPoint),
			hexutil.Encode(commitmentSeed),
			false,
		)

		return err
	})
}

func (p *PostgresChannels) AcceptLocalChannel(accept *lnwire.AcceptChannel) error {
	return NewTransactor(p.db, func(tx *sql.Tx) error {
		update, err := tx.Prepare(`
			UPDATE channels SET (
				their_funding_key,
				their_revocation_point,
				their_payment_point,
				their_delayed_payment_point,
				their_htlc_point,
				their_first_commitment_point
			) = ($1, $2, $3, $4, $5, $6) WHERE temporary_id = $7
		`)

		if err != nil {
			return err
		}

		_, err = update.Exec(
			crypto.BTCECToCompressedHex(accept.FundingKey),
			crypto.BTCECToCompressedHex(accept.RevocationPoint),
			crypto.BTCECToCompressedHex(accept.PaymentPoint),
			crypto.BTCECToCompressedHex(accept.DelayedPaymentPoint),
			crypto.BTCECToCompressedHex(accept.HtlcPoint),
			crypto.BTCECToCompressedHex(accept.FirstCommitmentPoint),
			hexutil.Encode(accept.PendingChannelID[:]),
		)

		return err
	})
}

func (p *PostgresChannels) GetPendingChannel(pendingChannelId [32]byte) (*types.Channel, error) {
	raw := &rawChannel{}

	err := p.db.QueryRow("SELECT our_funding_key, their_funding_key, funding_amount FROM channels WHERE temporary_id = $1",
		hexutil.Encode(pendingChannelId[:])).Scan(&raw.OurFundingAddress, &raw.TheirFundingAddress, &raw.FundingAmount)

	if err != nil {
		return nil, err
	}

	out, err := raw.ToChannel()

	if err != nil {
		return nil, err
	}

	return out, nil
}

func (p *PostgresChannels) FinalizeChannelId(pendingChannelId [32]byte, finalizedChannelId [32]byte, inputId []byte) error {
	return NewTransactor(p.db, func(tx *sql.Tx) error {
		var err error

		if inputId == nil {
			_, err = tx.Exec(
				"UPDATE channels SET finalized_id = $1 WHERE temporary_id = $2",
				hexutil.Encode(finalizedChannelId[:]),
				hexutil.Encode(pendingChannelId[:]),
			)
		} else {
			_, err = tx.Exec(
				"UPDATE channels SET (finalized_id, input_id) = ($1, $2) WHERE temporary_id = $3",
				hexutil.Encode(finalizedChannelId[:]),
				hexutil.Encode(inputId[:]),
				hexutil.Encode(pendingChannelId[:]),
			)
		}

		return err
	})
}

func (p *PostgresChannels) FinalizeChannelSignatures(finalizedChannelId [32]byte, ourSig crypto.Signature, theirSig crypto.Signature) error {
	hexId := hexutil.Encode(finalizedChannelId[:])

	return NewTransactor(p.db, func(tx *sql.Tx) error {
		var err error

		if ourSig == nil {
			_, err = tx.Exec(
				"UPDATE channels SET (their_signature) = ($1) WHERE finalized_id = $2",
				hexutil.Encode(theirSig.Bytes()),
				hexId,
			)
		} else if theirSig == nil {
			_, err = tx.Exec(
				"UPDATE channels SET (our_signature) = ($1) WHERE finalized_id = $2",
				hexutil.Encode(ourSig.Bytes()),
				hexId,
			)
		} else {
			_, err = tx.Exec(
				"UPDATE channels SET (our_signature, their_signature) = ($1, $2) WHERE finalized_id = $3",
				hexutil.Encode(ourSig.Bytes()),
				hexutil.Encode(theirSig.Bytes()),
				hexId,
			)
		}

		return err
	})
}

func (p *PostgresChannels) GetFinalizedChannel(finalizedChannelId [32]byte) (*types.Channel, error) {
	raw := &rawChannel{}

	err := p.db.QueryRow("SELECT our_funding_key, their_funding_key, funding_amount, our_signature, their_signature, input_id FROM channels WHERE finalized_id = $1",
		hexutil.Encode(finalizedChannelId[:])).Scan(&raw.OurFundingAddress, &raw.TheirFundingAddress, &raw.FundingAmount, &raw.OurSignature, &raw.TheirSignature, &raw.InputID)

	if err != nil {
		return nil, err
	}

	out, err := raw.ToChannel()

	if err != nil {
		return nil, err
	}

	return out, nil
}

func (r *rawChannel) ToChannel() (*types.Channel, error) {
	ours, err := crypto.PublicFromCompressedHex(r.OurFundingAddress)

	if err != nil {
		return nil, err
	}

	theirs, err := crypto.PublicFromCompressedHex(r.TheirFundingAddress)

	if err != nil {
		return nil, err
	}

	amount, err := conv.StringToBig(r.FundingAmount)

	if err != nil {
		return nil, err
	}

	var inputId [32]byte

	if r.InputID != "" {
		inputId, err = conv.HexToBytes32(r.InputID)

		if err != nil {
			return nil, err
		}
	}

	var ourSig crypto.Signature
	var theirSig crypto.Signature

	if r.OurSignature != "" {
		if ourSig, err = crypto.SignatureFromHex(r.OurSignature); err != nil {
			return nil, err
		}
	}

	if r.TheirSignature != "" {
		if theirSig, err = crypto.SignatureFromHex(r.TheirSignature); err != nil {
			return nil, err
		}
	}

	return &types.Channel{
		OurFundingAddress:   ours,
		TheirFundingAddress: theirs,
		FundingAmount:       amount,
		OurSignature:        ourSig,
		TheirSignature:      theirSig,
		InputID:             inputId,
	}, nil
}
