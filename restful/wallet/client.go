package wallet

import (
	"context"

	"bicycle/pkg/client"
)

func NewClient(baseUrl string, authKey string) *Client {
	return &Client{
		Client: client.NewClient(baseUrl, authKey),
	}
}

type Client struct {
	*client.Client
}

func (s *Client) BicycleNotification(ctx context.Context, req BicycleNotificationReq) (err error) {
	err = s.Post(ctx, s.BaseUrl+"/api/v1/wallet/bicycle/notification", req, nil)
	return
}
