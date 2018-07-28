package lndclient

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
	"fmt"
	"time"
	"golang.org/x/net/context"
	"github.com/kyokan/drawbridge/internal/logger"
	"go.uber.org/zap"
	"io/ioutil"
	"google.golang.org/grpc/credentials"
	"github.com/lightningnetwork/lnd/macaroons"
	"gopkg.in/macaroon.v2"
	"github.com/kyokan/drawbridge/pkg/crypto"
	"github.com/kyokan/drawbridge/internal/conv"
	"github.com/go-errors/errors"
	"crypto/sha256"
)

type LNDClient struct {
	client lnrpc.LightningClient
	ctx    context.Context
}

type LNDClientConfig struct {
	Host         string
	Port         string
	CertFile     string
	MacaroonFile string
	Context      context.Context
}

var log *zap.SugaredLogger

const InvoiceMemo = "Drawbridge Payment (v0)"
const InvoiceExpiry = 3600

func init() {
	log = logger.Logger.Named("lndclient")
}

func NewClient(config *LNDClientConfig) (*LNDClient, error) {
	creds, err := credentials.NewClientTLSFromFile(config.CertFile, "localhost")
	if err != nil {
		return nil, err
	}

	macaroonBytes, err := ioutil.ReadFile(config.MacaroonFile)
	if err != nil {
		return nil, err
	}

	mac := &macaroon.Macaroon{}
	if err = mac.UnmarshalBinary(macaroonBytes); err != nil {
		return nil, err
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithBlock(),
		grpc.WithPerRPCCredentials(macaroons.NewMacaroonCredential(mac)),
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", config.Host, config.Port), opts...)
	if err != nil {
		return nil, err
	}

	client := lnrpc.NewLightningClient(conn)

	return &LNDClient{
		client: client,
		ctx:    config.Context,
	}, nil
}

func (c *LNDClient) GetInfo() (*lnrpc.GetInfoResponse, error) {
	log.Infow("executing GetInfo RPC call")
	ctx, _ := context.WithTimeout(c.ctx, time.Second*10)
	return c.client.GetInfo(ctx, &lnrpc.GetInfoRequest{})
}

func (c *LNDClient) ConnectPeer(pubkey *crypto.PublicKey, host string) (error) {
	hasPeer, err := c.HasPeer(pubkey)
	if err != nil {
		return err
	}

	if hasPeer {
		log.Infow("already connected to peer", "pubkey", pubkey.CompressedHex())
		return nil
	}

	ctx, cancel := context.WithTimeout(c.ctx, time.Second*10)
	defer cancel()
	req := &lnrpc.ConnectPeerRequest{
		Addr: &lnrpc.LightningAddress{
			Pubkey: conv.Strip0x(pubkey.CompressedHex()),
			Host:   host,
		},
		Perm: true,
	}

	_, err = c.client.ConnectPeer(ctx, req)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hasPeer, err := c.HasPeer(pubkey)
			if err != nil {
				return err
			}

			if hasPeer {
				return nil
			}
		case <-ctx.Done():
			return errors.New("peer connection timed out")
		}
	}
}

func (c *LNDClient) HasPeer(pubkey *crypto.PublicKey) (bool, error) {
	ctx, cancel := context.WithTimeout(c.ctx, time.Second * 10)
	defer cancel()
	res, err := c.client.ListPeers(ctx, &lnrpc.ListPeersRequest{})
	if err != nil {
		return false, err
	}

	pub := conv.Strip0x(pubkey.CompressedHex())

	for _, peer := range res.Peers {
		if peer.PubKey == pub {
			return true, nil
		}
	}

	return false, nil
}

func (c *LNDClient) AddInvoice(amount int64, preimage []byte) (*lnrpc.AddInvoiceResponse, error) {
	descHash := sha256.Sum256(preimage)
	invoice := &lnrpc.Invoice{
		Memo: InvoiceMemo,
		RPreimage: preimage,
		Value: amount,
		DescriptionHash: descHash[:],
		Expiry: InvoiceExpiry,
	}

	ctx, cancel := context.WithTimeout(c.ctx, time.Second * 10)
	defer cancel()
	return c.client.AddInvoice(ctx, invoice)
}