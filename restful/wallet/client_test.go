package wallet

import (
	"context"
	"fmt"
	"testing"
)

func TestClient_CreateDepositAddress(t *testing.T) {
	c := NewClient("http://127.0.0.1:8080", "tokenspritz")
	c.Client.DevMode()
	req := CreateDepositAddressReq{}
	req.CoinID = 1
	req.ChainID = 1
	req.UserID = 3
	err := c.CreateDepositAddress(context.Background(), req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func TestClient_UserDepositAddress(t *testing.T) {
	c := NewClient("http://47.238.247.230:8083", "tokenspritz")
	c.Client.DevMode()
	req := UserDepositAddressReq{}
	req.CoinID = 3
	req.ChainID = 1
	req.UserID = 102
	d, err := c.UserDepositAddress(context.Background(), req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(d)
}

func TestClient_DepositRecords(t *testing.T) {
	c := NewClient("http://47.238.247.230:8083", "tokenspritz")
	c.Client.DevMode()
	req := DepositRecordsReq{
		CoinID:  1,
		ChainID: 1,
	}
	req.CoinID = 2
	req.ChainID = 1
	req.UserID = 1
	req.PageSize = 100
	req.Page = 1
	d, err := c.DepositRecords(context.Background(), req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(d)
}

func TestClient_WithdrawRecords(t *testing.T) {
	c := NewClient("http://47.238.247.230:8083", "tokenspritz")
	c.Client.DevMode()
	req := WithdrawRecordsReq{}
	req.CoinID = 2
	req.ChainID = 1
	req.UserID = 1
	req.PageSize = 100
	req.Page = 1
	d, err := c.WithdrawRecords(context.Background(), req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(d)
}

func TestClient_WithdrawReq(t *testing.T) {
	c := NewClient("http://47.238.247.230:8083", "tokenspritz")
	c.Client.DevMode()
	req := WithdrawReq{}
	req.CoinID = 2
	req.ChainID = 1
	req.UserID = 1
	req.Address = "0x0b4923A701357fb4b33574e76b9Be43ccb9BC4e3"
	req.Volume = "0.05"
	err := c.Withdraw(context.Background(), req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
