package asset

import (
	"bicycle/pkg/client"
	"context"
)

func NewClient(baseUrl string, authKey string) *Client {
	return &Client{
		Client: client.NewClient(baseUrl, authKey),
	}
}

type Client struct {
	*client.Client
}

func (s *Client) CoinList(ctx context.Context, req CoinListReq) (d CoinListResp, err error) {
	err = s.Get(ctx, s.BaseUrl+"/api/v1/asset/coin/list", req, &d)
	return
}

func (s *Client) ChainList(ctx context.Context, req ChainListReq) (d []ChainListItem, err error) {
	err = s.Get(ctx, s.BaseUrl+"/api/v1/asset/chain/list", req, &d)
	return
}

func (s *Client) TokenList(ctx context.Context, req TokenListReq) (d TokenListResp, err error) {
	err = s.Get(ctx, s.BaseUrl+"/api/v1/asset/token/list", req, &d)
	return
}