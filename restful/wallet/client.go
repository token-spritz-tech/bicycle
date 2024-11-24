package wallet

import (
	"context"

	"github.com/gobicycle/bicycle/pkg/client"
)

func NewClient(baseUrl string, authKey string) *Client {
	return &Client{
		Client: client.NewClient(baseUrl, authKey),
	}
}

type Client struct {
	*client.Client
}

func (s *Client) UserDepositAddress(ctx context.Context, req UserDepositAddressReq) (d UserDepositAddressResp, err error) {
	err = s.Get(ctx, s.BaseUrl+"/api/v1/wallet/deposit-address", req, &d)
	return
}

func (s *Client) CreateDepositAddress(ctx context.Context, req CreateDepositAddressReq) (err error) {
	err = s.Post(ctx, s.BaseUrl+"/api/v1/wallet/deposit-address", req, nil)
	return
}

func (s *Client) Withdraw(ctx context.Context, req WithdrawReq) (err error) {
	err = s.Post(ctx, s.BaseUrl+"/api/v1/wallet/withdraw", req, nil)
	return
}

func (s *Client) WithdrawRecords(ctx context.Context, req WithdrawRecordsReq) (d []WithdrawRecordItem, err error) {
	err = s.Get(ctx, s.BaseUrl+"/api/v1/wallet/withdraw/records", req, &d)
	return
}

func (s *Client) DepositRecords(ctx context.Context, req DepositRecordsReq) (d []DepositRecordItem, err error) {
	err = s.Get(ctx, s.BaseUrl+"/api/v1/wallet/deposit/records", req, &d)
	return
}

func (s *Client) DepositDetail(ctx context.Context, req DepositDetailReq) (d DepositRecordItem, err error) {
	err = s.Get(ctx, s.BaseUrl+"/api/v1/wallet/deposit/detail", req, &d)
	return
}

func (s *Client) WithdrawDetail(ctx context.Context, req WithdrawDetailReq) (d WithdrawRecordItem, err error) {
	err = s.Get(ctx, s.BaseUrl+"/api/v1/wallet/withdraw/detail", req, &d)
	return
}

func (s *Client) WithdrawApprove(ctx context.Context, req WithdrawApproveReq) (err error) {
	err = s.Post(ctx, s.BaseUrl+"/api/v1/wallet/withdraw/approve", req, nil)
	return
}

func (s *Client) WithdrawReject(ctx context.Context, req WithdrawRejectReq) (err error) {
	err = s.Post(ctx, s.BaseUrl+"/api/v1/wallet/withdraw/reject", req, nil)
	return
}
