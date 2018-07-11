package ethclient

import (
	"math/big"
	"go.uber.org/zap"
	"github.com/kyokan/drawbridge/internal/logger"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"strings"
	"github.com/kyokan/drawbridge/pkg/contracts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/drawbridge/internal/db"
	"time"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"context"
)

var csLog *zap.SugaredLogger

var lightningABI abi.ABI

const ConfirmationCount = 0

var CreateSignature = crypto.Keccak256Hash([]byte("Create(uint256,uint256,bytes,bytes32)"))

var SpendSignature = crypto.Keccak256Hash([]byte("Spend(uint256)"))

var WithdrawalSignature = crypto.Keccak256Hash([]byte("Withdrawal(address,uint256)"))

func init() {
	csLog = logger.Logger.Named("chainsaw")

	uAbi, err := abi.JSON(strings.NewReader(contracts.LightningERC20ABI))

	if err != nil {
		// can only happen if ABI generation is invalid during compilation
		panic(err)
	}

	lightningABI = uAbi
}

type CreateEvent struct {
	Owner    common.Address
	Value    *big.Int
	BlockNum *big.Int
	Script   []byte
	Id       [32]byte
}

type WithdrawalEvent struct {
	Owner common.Address
	Value *big.Int
}

type SpendEvent struct {
	Id [32]byte
}

type Chainsaw struct {
	client    *Client
	lastBlock uint64
	db        *db.DB
	lastTick  time.Time
}

func NewChainsaw(client *Client, db *db.DB) *Chainsaw {
	return &Chainsaw{
		client:    client,
		lastBlock: 0,
		db:        db,
	}
}

func (c *Chainsaw) Start() {
	csLog.Info("chainsaw started")

	c.lastTick = time.Now()

	lastBlock, err := c.db.Outputs.LastPoll()

	if err != nil {
		csLog.Errorw("failed to fetch initial poll data", "err", err.Error())
		return
	}

	c.lastBlock = lastBlock

	for {
		c.awaitNextTick()
		nextBlock := c.lastBlock + 1
		blockHeight, err := c.client.BlockHeight()
		if err != nil {
			csLog.Warnw("failed to get block height", "err", err.Error())
			continue
		}

		confirmedBlockHeight := blockHeight - ConfirmationCount
		if confirmedBlockHeight < nextBlock {
			csLog.Infow("already at latest block")
			continue
		}

		logs, err := c.client.FilterContract(nextBlock, confirmedBlockHeight)
		if err != nil {
			csLog.Warnw("failed to filter contract", "err", err)
			continue
		}

		results := &db.PolledOutputs{}

		for _, log := range logs {
			switch log.Topics[0] {
			case CreateSignature:
				out := &CreateEvent{}
				err := lightningABI.Unpack(out, "Create", log.Data)
				if err != nil {
					csLog.Errorw("failed to unpack event", "err", err.Error())
					continue
				}

				results.New = append(results.New, &db.ETHOutput{
					ID: out.Id,
					ContractAddress: log.Address,
					Amount: out.Value,
					BlockNumber: log.BlockNumber,
					TxHash: log.TxHash,
					Script: out.Script,
					Type: uint8(out.Script[0]),
					IsSpent: false,
					IsWithdrawn: false,
				})
				csLog.Infow("processed CreateEvent log", "id", hexutil.Encode(out.Id[:]))
			default:
				csLog.Infow("received unknown event", "topic", log.Topics[0].Hex())
			}
		}

		err = c.db.Outputs.SavePoll(results, confirmedBlockHeight)
		if err != nil {
			csLog.Errorw("failed to save poll", "err", err.Error())
			continue
		}

		csLog.Infow("finished poll", "blockHeight", confirmedBlockHeight)
		c.lastBlock = confirmedBlockHeight
	}
}

func (c *Chainsaw) awaitNextTick() {
	diff := time.Since(c.lastTick).Seconds()

	if diff < 15 {
		time.Sleep(time.Duration(15-diff) * time.Second)
	}

	c.lastTick = time.Now()
}

func AwaitOutput(ctx context.Context, d *db.DB, outputId common.Hash) (*db.ETHOutput, error) {
	ticker := time.NewTicker(time.Second * 10)

	for {
		select {
		case <- ticker.C:
			output, err := d.Outputs.FindById(outputId)
			if err != nil {
				return nil, err
			}
			if output != nil {
				return output, err
			}
			case <- ctx.Done():
				return nil, ctx.Err()
		}
	}
}