package core

import (
	"bicycle/config"
	"encoding/hex"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

// JettonProxy is a special contract wrapper that allow to control jetton wallet from TON wallet.
// It is possible create few jetton proxies for single TON wallet (as owner) and control multiple jetton wallets.
// Read about JettonProxy smart contract at README.md and https://github.com/gobicycle/ton-proxy-contract
type JettonProxy struct {
	Owner       *address.Address
	SubwalletID uint32
	address     *address.Address
	stateInit   *tlb.StateInit
}

func NewJettonProxy(subwalletId uint32, owner *address.Address) (*JettonProxy, error) {
	if owner == nil {
		return nil, fmt.Errorf("nil owner")
	}
	stateInit := buildJettonProxyStateInit(subwalletId, owner)
	stateCell, err := tlb.ToCell(stateInit)
	if err != nil {
		return nil, fmt.Errorf("failed to get state cell: %w", err)
	}
	addr := address.NewAddress(0, DefaultWorkchain, stateCell.Hash())

	return &JettonProxy{
		Owner:       owner,
		SubwalletID: subwalletId,
		address:     addr,
		stateInit:   stateInit,
	}, nil
}

func buildJettonProxyStateInit(subwalletId uint32, owner *address.Address) *tlb.StateInit {
	h, err := hex.DecodeString(config.JettonProxyContractCode)
	if err != nil {
		log.Fatalf("decode JettonProxyContractCode hex error: %v", err)
	}
	code, err := cell.FromBOC(h)
	if err != nil {
		log.Fatalf("parsing JettonProxyContractCode boc error: %v", err)
	}
	data := cell.BeginCell().
		MustStoreAddr(owner).
		MustStoreUInt(uint64(subwalletId), 32).
		EndCell()
	res := &tlb.StateInit{
		Code: code,
		Data: data,
	}
	return res
}

// Address returns address of jetton proxy contract
func (p *JettonProxy) Address() *address.Address {
	return p.address
}

// StateInit returns state init structure of jetton proxy contract
func (p *JettonProxy) StateInit() *tlb.StateInit {
	return p.stateInit
}

// BuildMessage wraps custom body payload to resend by proxy contract
func (p *JettonProxy) BuildMessage(destination *address.Address, body *cell.Cell) *tlb.InternalMessage {
	return &tlb.InternalMessage{
		IHRDisabled: true,
		Bounce:      true,
		DstAddr:     destination,
		Amount:      tlb.FromNanoTONU(0), // proxy sends all TONs with mode == 128 + 32
		Body:        body,
	}
}
