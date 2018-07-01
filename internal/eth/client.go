package eth

import (
	"github.com/kyokan/drawbridge/pkg/contracts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"github.com/ethereum/go-ethereum/core/types"
)

type Client struct {
	keyManager    *KeyManager
	conn          *ethclient.Client
	utxoContract  *contracts.UTXOToken
	erc20Contract *contracts.ERC20
	utxoAddress  common.Address
	erc20Address common.Address
}

func NewClient(keyManager *KeyManager, url string, address string) (*Client, error) {
	utxoAddress := common.HexToAddress(address)
	conn, err := ethclient.Dial(url)

	if err != nil {
		return nil, err
	}

	utxoContract, err := contracts.NewUTXOToken(utxoAddress, conn)

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
		keyManager:    keyManager,
		conn:          conn,
		utxoContract:  utxoContract,
		erc20Contract: erc20Contract,
		utxoAddress:   utxoAddress,
		erc20Address:  tokenContractAddress,
	}

	return wrapped, nil
}

func (c *Client) GetERC20Address() (common.Address) {
	return c.erc20Address
}

func (c *Client) ApproveERC20(tokens *big.Int) (*types.Transaction, error) {
	return c.erc20Contract.Approve(c.keyManager.NewTransactor(0), c.utxoAddress, tokens)
}

func (c *Client) Deposit(tokens *big.Int) (*types.Transaction, error) {
	return c.utxoContract.Deposit(c.keyManager.NewTransactor(500000), tokens)
}
