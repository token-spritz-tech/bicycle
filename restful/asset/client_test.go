package asset

import (
	"context"
	"fmt"
	"testing"
)

func TestClient_ChainList(t *testing.T) {
	c := NewClient("http://47.238.247.230:8081", "tokenspritz")
	c.Client.DevMode()
	req := ChainListReq{}
	req.Page = 1
	req.PageSize = 100
	d, err := c.ChainList(context.Background(), req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(d)
}

func TestClient_CoinList(t *testing.T) {
	c := NewClient("http://47.238.247.230:8081", "tokenspritz")
	c.Client.DevMode()
	req := CoinListReq{}
	req.Page = 1
	req.PageSize = 100
	d, err := c.CoinList(context.Background(), req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(d)
}

func TestClient_TokenList(t *testing.T) {
	c := NewClient("http://47.238.247.230:8081", "tokenspritz")
	c.Client.DevMode()
	req := TokenListReq{}
	req.Page = 1
	req.PageSize = 100
	d, err := c.TokenList(context.Background(), req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(d)
}
