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
)

var csLog *zap.SugaredLogger

var utxoAbi abi.ABI

var ConfirmationCount = big.NewInt(0)

var DepositEventSignature = crypto.Keccak256Hash([]byte("Deposit(address,uint256,bytes32)"))

var OutputEventSignature = crypto.Keccak256Hash([]byte("Output(bytes32,address,uint256,bytes32)"))

var WithdrawalEventSignature = crypto.Keccak256Hash([]byte("Withdrawal(address,uint256,bytes32)"))

func init() {
	csLog = logger.Logger.Named("chainsaw")

	uAbi, err := abi.JSON(strings.NewReader(contracts.LightningERC20ABI))

	if err != nil {
		// can only happen if ABI generation is invalid during compilation
		panic(err)
	}

	utxoAbi = uAbi
}

type OutputEvent struct {
	InputId [32]byte
	Owner   common.Address
	Value   *big.Int
	Id      [32]byte
}

type WithdrawalEvent struct {
	Owner common.Address
	Value *big.Int
	Id    [32]byte
}

type Chainsaw struct {
	client    *Client
	lastBlock *big.Int
	db        *db.DB
	lastTick  time.Time
}

func NewChainsaw(client *Client, db *db.DB) *Chainsaw {
	return &Chainsaw{
		client:    client,
		lastBlock: big.NewInt(0),
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
		nextBlock := c.lastBlock.Add(c.lastBlock, big.NewInt(1))
		blockHeight, err := c.client.BlockHeight()
		if err != nil {
			csLog.Warnw("failed to get block height", "err", err.Error())
			continue
		}

		confirmedBlockHeight := blockHeight.Sub(blockHeight, ConfirmationCount)
		if confirmedBlockHeight.Cmp(nextBlock) <= 0 {
			csLog.Infow("already at latest block")
			continue
		}

		if err != nil {
			csLog.Warnw("failed to filter UTXO contract", "err", err.Error())
		}

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
