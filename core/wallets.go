package core

import (
	"bicycle/audit"
	"bicycle/config"
	"context"
	"errors"
	"fmt"
	"math/big"
	"math/rand"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tonkeeper/tongo/boc"
	tongoTlb "github.com/tonkeeper/tongo/tlb"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/jetton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type Wallets struct {
	Shard            byte
	TonHotWallet     *wallet.Wallet
	TonBasicWallet   *wallet.Wallet // basic V3 wallet to make other wallets with different subwallet_id
	JettonHotWallets map[string]JettonWallet
}

// InitWallets
// Generates highload hot-wallet and map[currency]JettonWallet Jetton wallets, and saves to DB
// TON highload hot-wallet (for seed and default subwallet_id) must be already active for success initialization.
func InitWallets(
	ctx context.Context,
	db storage,
	bc blockchain,
	seed string,
	jettons map[string]config.Jetton,
) (Wallets, error) {
	if config.Config.ColdWallet != nil && config.Config.ColdWallet.IsBounceable() {
		_, status, err := bc.GetAccountCurrentState(ctx, config.Config.ColdWallet)
		if err != nil {
			return Wallets{}, err
		}
		log.Infof("Cold wallet status: %s", status)
		if status != tlb.AccountStatusActive {
			return Wallets{}, fmt.Errorf("cold wallet address must be non-bounceable for not active wallet")
		}
	}

	tonHotWallet, shard, subwalletId, err := initTonHotWallet(ctx, db, bc, seed)
	if err != nil {
		return Wallets{}, err
	}

	tonBasicWallet, _, _, err := bc.GenerateDefaultWallet(seed, false)
	if err != nil {
		return Wallets{}, err
	}
	// don't set TTL here because spec is not inherited by GetSubwallet method

	jettonHotWallets := make(map[string]JettonWallet)
	for currency, j := range jettons {
		w, err := initJettonHotWallet(ctx, db, bc, tonHotWallet.Address(), j.Master, currency, subwalletId)
		if err != nil {
			return Wallets{}, err
		}
		jettonHotWallets[currency] = w
	}

	return Wallets{
		Shard:            shard,
		TonHotWallet:     tonHotWallet,
		TonBasicWallet:   tonBasicWallet,
		JettonHotWallets: jettonHotWallets,
	}, nil
}

func initTonHotWallet(
	ctx context.Context,
	db storage,
	bc blockchain,
	seed string,
) (
	tonHotWallet *wallet.Wallet,
	shard byte,
	subwalletId uint32,
	err error,
) {
	tonHotWallet, shard, subwalletId, err = bc.GenerateDefaultWallet(seed, true)
	if err != nil {
		return nil, 0, 0, err
	}
	hotSpec := tonHotWallet.GetSpec().(*wallet.SpecHighloadV2R2)
	hotSpec.SetMessagesTTL(uint32(config.ExternalMessageLifetime.Seconds()))

	addr := AddressMustFromTonutilsAddress(tonHotWallet.Address())
	alreadySaved := false
	addrFromDb, err := db.GetTonHotWalletAddress(ctx)
	if err == nil && addr != addrFromDb {
		audit.Log(audit.Error, string(TonHotWallet), InitEvent,
			fmt.Sprintf("Hot TON wallet address is not equal to the one stored in the database. Maybe seed was being changed. %s != %s",
				tonHotWallet.Address().String(), addrFromDb.ToTonutilsAddressStd(0).String()))
		return nil, 0, 0,
			fmt.Errorf("saved hot wallet not equal generated hot wallet. Maybe seed was being changed")
	} else if !errors.Is(err, ErrNotFound) && err != nil {
		return nil, 0, 0, err
	} else if err == nil {
		alreadySaved = true
	}

	log.Infof("Shard: %v", shard)
	log.Infof("TON hot wallet address: %v", tonHotWallet.Address().String())

	balance, status, err := bc.GetAccountCurrentState(ctx, tonHotWallet.Address())
	if err != nil {
		return nil, 0, 0, err
	}
	hotWalletMin := decimal.NewFromFloat(0.1).Mul(decimal.NewFromInt(10).Pow(decimal.NewFromInt(9)))
	if balance.Cmp(hotWalletMin.BigInt()) == -1 { // hot wallet balance < TonHotWalletMinimumBalance
		return nil, 0, 0,
			fmt.Errorf("hot wallet balance must be at least %v nanoTON", hotWalletMin)
	}
	if status != tlb.AccountStatusActive {
		err = bc.DeployTonWallet(ctx, tonHotWallet)
		if err != nil {
			return nil, 0, 0, err
		}
	}
	if !alreadySaved {
		err = db.SaveTonWallet(ctx, WalletData{
			SubwalletID: uint32(wallet.DefaultSubwallet),
			Currency:    TonSymbol,
			Type:        TonHotWallet,
			Address:     addr,
		})
		if err != nil {
			return nil, 0, 0, err
		}
	}
	return tonHotWallet, shard, subwalletId, nil
}

func initJettonHotWallet(
	ctx context.Context,
	db storage,
	bc blockchain,
	tonHotWallet, jettonMaster *address.Address,
	currency string,
	subwalletId uint32,
) (JettonWallet, error) {
	// not init or check balances of Jetton wallets, it is not required for the service to work
	a, err := bc.GetJettonWalletAddress(ctx, tonHotWallet, jettonMaster)
	if err != nil {
		return JettonWallet{}, err
	}
	res := JettonWallet{Address: a, Currency: currency}
	log.Infof("%v jetton hot wallet address: %v", currency, a.String())

	ownerAddr, err := AddressFromTonutilsAddress(tonHotWallet)
	if err != nil {
		return JettonWallet{}, err
	}
	jettonWalletAddr, err := AddressFromTonutilsAddress(a)
	if err != nil {
		return JettonWallet{}, err
	}

	walletData, isPresented, err := db.GetJettonWallet(ctx, jettonWalletAddr)
	if err != nil {
		return JettonWallet{}, err
	}

	if isPresented && walletData.Currency == currency {
		return res, nil
	} else if isPresented && walletData.Currency != currency {
		audit.Log(audit.Error, string(JettonHotWallet), InitEvent,
			fmt.Sprintf("Hot Jetton wallets %s and %s have the same address %s",
				walletData.Currency, currency, a.String()))
		return JettonWallet{}, fmt.Errorf("jetton hot wallet address duplication")
	}

	err = db.SaveJettonWallet(
		ctx,
		ownerAddr,
		WalletData{
			SubwalletID: subwalletId,
			Currency:    currency,
			Type:        JettonHotWallet,
			Address:     jettonWalletAddr,
		},
		true,
	)
	if err != nil {
		return JettonWallet{}, err
	}
	return res, nil
}

func buildComment(comment string) *cell.Cell {
	root := cell.BeginCell().MustStoreUInt(0, 32)
	if err := root.StoreStringSnake(comment); err != nil {
		log.Fatalf("memo must fit into cell")
	}
	return root.EndCell()
}

func LoadComment(cell *cell.Cell) string {
	if cell == nil {
		return ""
	}
	l := cell.BeginParse()
	if val, err := l.LoadUInt(32); err == nil && val == 0 {
		str, err := l.LoadStringSnake()
		if err != nil {
			log.Errorf("load comment error: %v", err)
			return ""
		}
		return str
	}
	return ""
}

// WithdrawTONs
// Send all TON from one wallet (and deploy it if needed) to another and destroy "from" wallet contract.
// Wallet must be not empty.
func WithdrawTONs(ctx context.Context, from, to *wallet.Wallet, comment string) error {
	if from == nil || to == nil || to.Address() == nil {
		return fmt.Errorf("nil wallet")
	}
	var body *cell.Cell
	if comment != "" {
		body = buildComment(comment)
	}
	return from.Send(ctx, &wallet.Message{
		Mode: 128 + 32, // 128 + 32 send all and destroy
		InternalMessage: &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      false,
			DstAddr:     to.Address(),
			Amount:      tlb.FromNanoTONU(0),
			Body:        body,
		},
	}, false)
}

func WithdrawJettons(
	ctx context.Context,
	from, to *wallet.Wallet,
	jettonWallet *address.Address,
	forwardAmount tlb.Coins,
	amount Coins,
	comment string,
) error {
	if from == nil || to == nil || to.Address() == nil {
		return fmt.Errorf("nil wallet")
	}
	body := MakeJettonTransferMessage(
		to.Address(),
		to.Address(),
		amount.BigInt(),
		forwardAmount,
		rand.Int63(),
		comment,
		"",
	)
	return from.Send(ctx, &wallet.Message{
		Mode: 128 + 32, // 128 + 32 send all and destroy
		InternalMessage: &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      true,
			DstAddr:     jettonWallet, // jetton wallet address
			Amount:      tlb.FromNanoTONU(0),
			Body:        body,
		},
	}, false)
}

func MakeJettonTransferMessage(
	destination, responseDest *address.Address,
	amount *big.Int,
	forwardAmount tlb.Coins,
	queryId int64,
	comment string,
	binaryComment string,
) *cell.Cell {
	forwardPayload := cell.BeginCell().EndCell()

	if binaryComment != "" {
		c, err := decodeBinaryComment(binaryComment)
		if err != nil {
			log.Fatalf("decode binary comment error : %s", err.Error())
		}
		forwardPayload = c
	} else if comment != "" {
		forwardPayload = buildComment(comment)
	}

	payload, err := tlb.ToCell(jetton.TransferPayload{
		QueryID:             uint64(queryId),
		Amount:              tlb.FromNanoTON(amount),
		Destination:         destination,
		ResponseDestination: responseDest,
		CustomPayload:       nil,
		ForwardTONAmount:    forwardAmount,
		ForwardPayload:      forwardPayload,
	})
	if err != nil {
		log.Fatalf("jetton transfer message serialization error: %s", err.Error())
	}

	return payload
}

// decodeBinaryComment implements decoding of hex string and put it into cell with TLB scheme:
// `binary_comment#b3ddcf7d {n:#} data:(SnakeData ~n) = InternalMsgBody;`
func decodeBinaryComment(comment string) (*cell.Cell, error) {
	bitString, err := boc.BitStringFromFiftHex(comment)
	if err != nil {
		return nil, err
	}

	c := boc.NewCell()
	err = c.WriteUint(0xb3ddcf7d, 32) // binary_comment#b3ddcf7d
	if err != nil {
		return nil, err
	}

	err = tongoTlb.Marshal(c, tongoTlb.SnakeData(*bitString))
	if err != nil {
		return nil, err
	}

	b, err := c.ToBoc()
	if err != nil {
		return nil, err
	}

	return cell.FromBOC(b)
}

func BuildTonWithdrawalMessage(t ExternalWithdrawalTask) *wallet.Message {
	internalMessage := tlb.InternalMessage{
		IHRDisabled: true,
		Bounce:      t.Bounceable,
		DstAddr:     t.Destination.ToTonutilsAddressStd(0),
		Amount:      tlb.FromNanoTON(t.Amount.BigInt()),
	}

	if t.BinaryComment != "" {
		c, err := decodeBinaryComment(t.BinaryComment)
		if err != nil {
			log.Fatalf("decode binary comment error : %s", err.Error())
		}
		internalMessage.Body = c
	} else if t.Comment != "" {
		internalMessage.Body = buildComment(t.Comment)
	} else {
		internalMessage.Body = cell.BeginCell().EndCell()
	}

	return &wallet.Message{
		Mode:            3,
		InternalMessage: &internalMessage,
	}
}

func BuildJettonWithdrawalMessage(
	t ExternalWithdrawalTask,
	highloadWallet *wallet.Wallet,
	fromJettonWallet *address.Address,
) *wallet.Message {
	body := MakeJettonTransferMessage(
		t.Destination.ToTonutilsAddressStd(0),
		highloadWallet.Address(),
		t.Amount.BigInt(),
		config.JettonForwardAmount,
		t.QueryID,
		t.Comment,
		t.BinaryComment,
	)

	return &wallet.Message{
		Mode: 3,
		InternalMessage: &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      true,
			DstAddr:     fromJettonWallet,
			Amount:      config.JettonTransferTonAmount,
			Body:        body,
		},
	}
}

func BuildJettonProxyWithdrawalMessage(
	proxy JettonProxy,
	jettonWallet, tonWallet *address.Address,
	forwardAmount tlb.Coins,
	amount *big.Int,
	comment string,
) *wallet.Message {
	jettonTransferPayload := MakeJettonTransferMessage(
		tonWallet,
		tonWallet,
		amount,
		forwardAmount,
		rand.Int63(),
		comment,
		"",
	)

	msg, err := tlb.ToCell(proxy.BuildMessage(jettonWallet, jettonTransferPayload))
	if err != nil {
		log.Fatalf("build proxy message cell error: %v", err)
	}
	body := cell.BeginCell().MustStoreRef(msg).EndCell()
	return &wallet.Message{
		Mode: 3,
		InternalMessage: &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      true,
			DstAddr:     proxy.Address(),
			Amount:      config.JettonTransferTonAmount,
			Body:        body,
			StateInit:   proxy.StateInit(),
		},
	}
}

func buildJettonProxyServiceTonWithdrawalMessage(
	proxy JettonProxy,
	tonWallet *address.Address,
	memo uuid.UUID,
) *wallet.Message {
	msg, err := tlb.ToCell(proxy.BuildMessage(tonWallet, buildComment(memo.String())))
	if err != nil {
		log.Fatalf("build proxy message cell error: %v", err)
	}
	body := cell.BeginCell().MustStoreRef(msg).EndCell()
	return &wallet.Message{
		Mode: 3,
		InternalMessage: &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      true,
			DstAddr:     proxy.Address(),
			Amount:      config.JettonTransferTonAmount,
			Body:        body,
			StateInit:   proxy.StateInit(),
		},
	}
}

func buildTonFillMessage(
	to *address.Address,
	amount tlb.Coins,
	memo uuid.UUID,
) *wallet.Message {
	return &wallet.Message{
		Mode: 3,
		InternalMessage: &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      false,
			DstAddr:     to,
			Amount:      amount,
			Body:        buildComment(memo.String()),
		},
	}
}
