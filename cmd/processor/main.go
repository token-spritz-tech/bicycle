package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"bicycle/api"
	"bicycle/blockchain"
	"bicycle/config"
	"bicycle/core"
	"bicycle/db"

	log "github.com/sirupsen/logrus"
	"github.com/zeromicro/go-zero/core/conf"
)

var Version = "dev"

var configFile = flag.String("f", "bicycle.yaml", "the config file")

func main() {
	log.Infof("App version: %s", Version)
	flag.Parse()

	conf.MustLoad(*configFile, &config.Config)
	config.LoadConfig()

	log.Info("config loaded")
	confStr, err := json.Marshal(config.Config)
	if err != nil {
		log.Fatalf("marshal config error: %v", err)
	}
	fmt.Println(string(confStr))

	confStr, err = json.Marshal(config.Config)
	if err != nil {
		log.Fatalf("marshal config error: %v", err)
	}
	fmt.Println(string(confStr))

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)
	wg := new(sync.WaitGroup)

	bcClient, err := blockchain.NewConnection(config.Config.NetworkConfigUrl)
	if err != nil {
		log.Fatalf("blockchain connection error: %v", err)
	}

	dbClient, err := db.NewConnection(config.Config.DatabaseURI)
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()

	err = dbClient.LoadAddressBook(ctx)
	if err != nil {
		log.Fatalf("address book loading error: %v", err)
	}

	isTimeSynced, err := bcClient.CheckTime(ctx, config.AllowableServiceToNodeTimeDiff)
	if err != nil {
		log.Fatalf("get node time err: %v", err)
	}
	if !isTimeSynced {
		log.Fatalf("Service and Node time not synced")
	}

	wallets, err := core.InitWallets(ctx, dbClient, bcClient, config.Config.Seed, config.Config.Jettons)
	if err != nil {
		log.Fatalf("Wallets initialization error: %v", err)
	}

	var tracker *blockchain.ShardTracker
	block, err := dbClient.GetLastSavedBlockID(ctx)
	if !errors.Is(err, core.ErrNotFound) && err != nil {
		log.Fatalf("Get last saved block error: %v", err)
	} else if errors.Is(err, core.ErrNotFound) {
		tracker = blockchain.NewShardTracker(wallets.Shard, nil, bcClient)
	} else {
		tracker = blockchain.NewShardTracker(wallets.Shard, block, bcClient)
	}

	blockScanner := core.NewBlockScanner(wg, dbClient, bcClient, wallets.Shard, tracker, config.WalletClient)

	withdrawalsProcessor := core.NewWithdrawalsProcessor(
		wg, dbClient, bcClient, wallets, config.Config.ColdWallet)
	withdrawalsProcessor.Start()

	apiMux := http.NewServeMux()
	h := api.NewHandler(dbClient, bcClient, wallets.Shard, *wallets.TonHotWallet.Address())
	api.RegisterHandlers(apiMux, h)
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", config.Config.APIPort), apiMux)
		if err != nil {
			log.Fatalf("api error: %v", err)
		}
	}()

	go func() {
		<-sigChannel
		log.Printf("SIGTERM received")
		blockScanner.Stop()
		withdrawalsProcessor.Stop()
	}()

	wg.Wait()
}
