package ethclient

import (
	"github.com/kyokan/drawbridge/pkg/contracts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/kyokan/drawbridge/internal/wallet"
	"github.com/kyokan/drawbridge/pkg/crypto"
	"github.com/kyokan/drawbridge/internal/conv"
	"github.com/kyokan/drawbridge/pkg/txout"
		)

type DepositResult struct {
}

type Client struct {
	keyManager       *wallet.KeyManager
	rpc              *rpc.Client
	client           *ethclient.Client
	lightning        *contracts.LightningERC20
	token            *contracts.ERC20
	lightningAddress common.Address
	erc20Address     common.Address
}

func NewClient(keyManager *wallet.KeyManager, url string, address string) (*Client, error) {
	lightningAddress := common.HexToAddress(address)
	r, err := rpc.DialContext(context.Background(), url)
	if err != nil {
		return nil, err
	}

	conn := ethclient.NewClient(r)
	lightning, err := contracts.NewLightningERC20(lightningAddress, conn)
	if err != nil {
		return nil, err
	}

	tokenContractAddress, err := lightning.TokenAddress(nil)
	if err != nil {
		return nil, err
	}

	erc20Contract, err := contracts.NewERC20(tokenContractAddress, conn)
	if err != nil {
		return nil, err
	}

	wrapped := &Client{
		keyManager:       keyManager,
		rpc:              r,
		client:           conn,
		lightning:        lightning,
		token:            erc20Contract,
		lightningAddress: lightningAddress,
		erc20Address:     tokenContractAddress,
	}

	return wrapped, nil
}

func (c *Client) BlockHeight() (uint64, error) {
	var hex string
	err := c.rpc.Call(&hex, "eth_blockNumber")
	if err != nil {
		return 0, err
	}

	blockHeight, err := conv.HexToBig(hex)
	if err != nil {
		return 0, err
	}

	return blockHeight.Uint64(), err
}

func (c *Client) FilterContract(from uint64, to uint64) ([]ethtypes.Log, error) {
	q := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(from)),
		ToBlock:   big.NewInt(int64(to)),
		Addresses: []common.Address{
			c.lightningAddress,
		},
	}

	return c.client.FilterLogs(context.Background(), q)
}

func (c *Client) GetERC20Address() (common.Address) {
	return c.erc20Address
}

func (c *Client) ApproveERC20(tokens *big.Int) (*ethtypes.Transaction, error) {
	return c.token.Approve(c.keyManager.NewTransactor(0), c.lightningAddress, tokens)
}

func (c *Client) Deposit(tokens *big.Int) (*ethtypes.Transaction, error) {
	tx, err := c.lightning.Deposit(c.keyManager.NewTransactor(500000), tokens)
	if err != nil {
		return nil, err
	}

	return tx, err
}

func (c *Client) DepositMultisig(req *txout.SpendRequest, sig crypto.Signature) (*ethtypes.Transaction, error) {
	inputs, outputs, err := txout.WireData(req, sig)
	if err != nil {
		return nil, err
	}

	tx, err := c.lightning.Spend(c.keyManager.NewTransactor(3000000), inputs, outputs)
	if err != nil {
		return nil, err
	}

	return tx, nil
}
