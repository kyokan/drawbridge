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
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/kyokan/drawbridge/internal/wallet"
	"github.com/kyokan/drawbridge/pkg/crypto"
)

type DepositResult struct {
}

type Client struct {
	keyManager   *wallet.KeyManager
	rpc          *rpc.Client
	client       *ethclient.Client
	lightning    *contracts.LightningERC20
	token        *contracts.ERC20
	utxoAddress  common.Address
	erc20Address common.Address
}

type DepositMultisigOpts struct {
	InputID        [32]byte
	OurAddress     common.Address
	TheirAddress   common.Address
	OurSignature   crypto.Signature
	TheirSignature crypto.Signature
}

func NewClient(keyManager *wallet.KeyManager, url string, address string) (*Client, error) {
	utxoAddress := common.HexToAddress(address)

	r, err := rpc.DialContext(context.Background(), url)

	if err != nil {
		return nil, err
	}

	conn := ethclient.NewClient(r)

	utxoContract, err := contracts.NewLightningERC20(utxoAddress, conn)

	if err != nil {
		return nil, err
	}

	tokenContractAddress, err := utxoContract.TokenAddress(nil)

	if err != nil {
		return nil, err
	}

	erc20Contract, err := contracts.NewERC20(tokenContractAddress, conn)

	if err != nil {
		return nil, err
	}

	wrapped := &Client{
		keyManager:   keyManager,
		rpc:          r,
		client:       conn,
		lightning:    utxoContract,
		token:        erc20Contract,
		utxoAddress:  utxoAddress,
		erc20Address: tokenContractAddress,
	}

	return wrapped, nil
}

func (c *Client) BlockHeight() (*big.Int, error) {
	var hexRes string
	err := c.rpc.Call(&hexRes, "eth_blockNumber")

	if err != nil {
		return nil, err
	}

	return hexutil.DecodeBig(hexRes)
}

func (c *Client) FilterContract(from *big.Int, to *big.Int) ([]ethtypes.Log, error) {
	q := ethereum.FilterQuery{
		FromBlock: from,
		ToBlock:   to,
		Addresses: []common.Address{
			c.utxoAddress,
		},
	}

	return c.client.FilterLogs(context.Background(), q)
}

func (c *Client) GetERC20Address() (common.Address) {
	return c.erc20Address
}

func (c *Client) ApproveERC20(tokens *big.Int) (*ethtypes.Transaction, error) {
	return c.token.Approve(c.keyManager.NewTransactor(0), c.utxoAddress, tokens)
}

func (c *Client) Deposit(tokens *big.Int) (*ethtypes.Transaction, error) {
	tx, err := c.lightning.Deposit(c.keyManager.NewTransactor(500000), tokens)

	if err != nil {
		return nil, err
	}

	return tx, err
}

func (c *Client) DepositMultisig(opts *DepositMultisigOpts) (*ethtypes.Transaction, error) {
	return nil, nil
}
