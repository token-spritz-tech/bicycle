package wallet

import (
	"context"
	"fmt"
	"testing"
)

func TestClient_BicycleNotification(t *testing.T) {
	c := NewClient("http://127.0.0.1:8080", "tokenspritz")
	c.Client.DevMode()
	req := BicycleNotificationReq{}
	err := c.BicycleNotification(context.Background(), req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
