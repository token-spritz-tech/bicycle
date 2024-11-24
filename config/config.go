package config

import (
	"context"
	"errors"
	"log"
	"math/big"
	"time"

	"bicycle/restful/asset"
	"bicycle/restful/wallet"

	"github.com/shopspring/decimal"
	"github.com/tonkeeper/tongo/boc"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
)

const MaxJettonForwardTonAmount = 20_000_000

var (
	JettonTransferTonAmount     = tlb.FromNanoTONU(100_000_000)
	JettonForwardAmount         = tlb.FromNanoTONU(MaxJettonForwardTonAmount) // must be < JettonTransferTonAmount
	JettonInternalForwardAmount = tlb.FromNanoTONU(1)

	DefaultHotWalletHysteresis = decimal.NewFromFloat(0.95) // `hot_wallet_residual_balance` = `hot_wallet_max_balance` * `hysteresis`

	ExternalMessageLifetime = 50 * time.Second

	ExternalWithdrawalPeriod  = 80 * time.Second // must be ExternalWithdrawalPeriod > ExternalMessageLifetime and some time for balance update
	InternalWithdrawalPeriod  = 80 * time.Second
	ExpirationProcessorPeriod = 5 * time.Second

	AllowableBlockchainLagging     = 40 * time.Second // TODO: use env var
	AllowableServiceToNodeTimeDiff = 2 * time.Second

	AssetClient  *asset.Client
	WalletClient *wallet.Client
)

// JettonProxyContractCode source code at https://github.com/gobicycle/ton-proxy-contract
const JettonProxyContractCode = "B5EE9C72410102010037000114FF00F4A413F4BCF2C80B010050D33331D0D3030171B0915BE0FA4030ED44D0FA4030C705F2E1939320D74A97D4018100A0FB00E8301E8A9040"

const MaxCommentLength = 1000 // qty in chars

var Config = struct {
	Seed                     string `json:""`
	DatabaseURI              string `json:""`
	APIPort                  int    `json:""`
	Testnet                  bool   `json:",default=true"`
	IsDepositSideCalculation bool   `json:",default=true"`
	QueueURI                 string `json:",optional"`
	QueueName                string `json:",optional"`
	QueueEnabled             bool   `json:",default=false"`
	NetworkConfigUrl         string `json:""`
	WebhookEndpoint          string `json:",optional"`
	WebhookToken             string `json:",optional"`
	AllowableLaggingSec      int    `json:",optional"`
	ForwardTonAmount         int    `json:",default=1"`
	WalletClientUrl          string `json:""`                     // TS钱包服务
	AssetClientUrl           string `json:""`                     // TS资产服务
	ClientKey                string `json:",default=tokenspritz"` // 客户端密钥
	Jettons                  map[string]Jetton
	Ton                      Cutoffs
	ColdWallet               *address.Address
	BlockchainConfig         *boc.Cell
	Coins                    map[int]asset.CoinListItem
	Chain                    asset.ChainListItem
	Tokens                   map[string]asset.TokenListItem
}{}

type Jetton struct {
	Master           *address.Address
	WithdrawalCutoff *big.Int
}

type Cutoffs struct {
	Withdrawal *big.Int
}

func LoadConfig() {
	if Config.ForwardTonAmount < 0 || Config.ForwardTonAmount > MaxJettonForwardTonAmount {
		log.Fatalf("Forward TON amount for jetton transfer must be positive and less than %d", MaxJettonForwardTonAmount)
	} else {
		JettonForwardAmount = tlb.FromNanoTONU(uint64(Config.ForwardTonAmount))
	}

	if Config.AllowableLaggingSec != 0 {
		AllowableBlockchainLagging = time.Second * time.Duration(Config.AllowableLaggingSec)
	}

	AssetClient = asset.NewClient(Config.AssetClientUrl, Config.ClientKey)
	WalletClient = wallet.NewClient(Config.WalletClientUrl, Config.ClientKey)

	// 加载TON链、Token、币种
	err := LoadTonChain()
	if err != nil {
		log.Fatalf("can not load ton chain: %v", err)
		panic(err)
	}
	err = LoadCoins()
	if err != nil {
		log.Fatalf("can not load coins: %v", err)
		panic(err)
	}
	err = LoadTokens()
	if err != nil {
		log.Fatalf("can not load tokens: %v", err)
		panic(err)
	}
	if len(Config.Tokens) == 0 {
		log.Fatalf("no tokens found")
		panic("no tokens found")
	}
	// 解析Token
	parseTokens()
}

// 加载币种
func LoadCoins() (err error) {
	coinList, err := AssetClient.CoinList(context.Background(), asset.CoinListReq{
		Page:     1,
		PageSize: 10000,
	})
	if err != nil {
		return
	}
	for _, coin := range coinList.Records {
		Config.Coins[coin.ID] = coin
	}
	return nil
}

// 加载Token
func LoadTokens() (err error) {
	if Config.Chain.ID == 0 {
		return errors.New("chain id is not set")
	}
	if len(Config.Coins) == 0 {
		return errors.New("coins are not set")
	}
	tokenList, err := AssetClient.TokenList(context.Background(), asset.TokenListReq{
		Page:     1,
		PageSize: 10000,
		ChainID:  Config.Chain.ID,
	})
	if err != nil {
		return
	}
	for _, token := range tokenList.Records {
		if _, ok := Config.Coins[int(token.CoinID)]; !ok {
			log.Fatalf("coin %v not found", token.CoinID)
			continue
		}
		Config.Tokens[Config.Coins[int(token.CoinID)].Name] = token
	}
	return
}

// 加载TON链
func LoadTonChain() (err error) {
	chainList, err := AssetClient.ChainList(context.Background(), asset.ChainListReq{
		Page:     1,
		PageSize: 10000,
	})
	if err != nil {
		return
	}
	for _, chain := range chainList {
		if chain.Name == "TON" {
			Config.Chain = chain
			return
		}
	}
	err = errors.New("ton chain not found")
	return
}

func parseTokens() {
	if len(Config.Tokens) == 0 {
		panic("no tokens found")
	}
	for _, token := range Config.Tokens {
		if _, ok := Config.Coins[int(token.CoinID)]; !ok {
			log.Fatalf("coin %v not found", token.CoinID)
			continue
		}
		c := Config.Coins[int(token.CoinID)]
		if c.Name == "TON" {
			minWithdrawVolume, err := decimal.NewFromString(token.MinWithdrawVolume)
			withdrawalCutoff := minWithdrawVolume.Mul(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(token.Decimal))))
			if err != nil {
				log.Fatalf("invalid %v jetton withdrawal cutoff: %v", token.ID, err)
			}
			Config.Ton = Cutoffs{
				Withdrawal: withdrawalCutoff.BigInt(),
			}
		} else {
			addr, err := address.ParseAddr(token.Address)
			if err != nil {
				log.Fatalf("invalid jetton address: %v", err)
			}
			minWithdrawVolume, err := decimal.NewFromString(token.MinWithdrawVolume)
			withdrawalCutoff := minWithdrawVolume.Mul(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(token.Decimal))))
			if err != nil {
				log.Fatalf("invalid %v jetton withdrawal cutoff: %v", token.ID, err)
			}
			Config.Jettons[c.Name] = Jetton{
				Master:           addr,
				WithdrawalCutoff: withdrawalCutoff.BigInt(),
			}
		}
	}
}
