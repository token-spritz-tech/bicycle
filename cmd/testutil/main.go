package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gobicycle/bicycle/blockchain"
	"github.com/gobicycle/bicycle/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xssnick/tonutils-go/address"
)

var (
	onlyMonitoring = true
	Version        = "dev"
)

const (
	depositsQty          = 10
	tonWithdrawAmount    = 550_000_000
	jettonWithdrawAmount = 550_000_000
	tonMinCutoff         = 10_000_000_000
)

func main() {
	log.Printf("App version: %s", Version)

	config.GetConfig()
	if circulation := os.Getenv("CIRCULATION"); circulation == "true" {
		onlyMonitoring = false
	}

	urlA := os.Getenv("HOST_A")
	if urlA == "" {
		log.Fatalf("empty HOST_A env var")
	}
	urlB := os.Getenv("HOST_B")
	if urlB == "" {
		log.Fatalf("empty HOST_B env var")
	}

	hotWalletA, err := address.ParseAddr(os.Getenv("HOT_WALLET_A"))
	if err != nil {
		log.Fatalf("invalid HOT_WALLET_A env var")
	}
	hotWalletB, err := address.ParseAddr(os.Getenv("HOT_WALLET_B"))
	if err != nil {
		log.Fatalf("invalid HOT_WALLET_B env var")
	}

	bcClient, err := blockchain.NewConnection(config.Config.NetworkConfigUrl)
	if err != nil {
		log.Fatalf("blockchain connection error: %v", err)
	}

	httpClient := NewClient(urlA, urlB, config.Config.APIToken, "TestClient")

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Fatal(http.ListenAndServe(":9101", nil))
	}()

	depositsA, err := httpClient.InitDeposits(urlA)
	if err != nil {
		log.Fatalf("can not init deposits: %v", err)
	}

	depositsB, err := httpClient.InitDeposits(urlB)
	if err != nil {
		log.Fatalf("can not init deposits: %v", err)
	}

	payerProc := NewPayerProcessor(context.TODO(), httpClient, bcClient, depositsA, depositsB, hotWalletA, hotWalletB)
	payerProc.Start()

	for {
		time.Sleep(time.Hour)
	}
}
