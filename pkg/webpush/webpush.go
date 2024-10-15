package webpush

import (
	"github.com/SherClockHolmes/webpush-go"
	"github.com/injunweb/backend-server/internal/config"
)

type Subscription struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

func SendNotification(subscription Subscription, message string) error {
	s := &webpush.Subscription{
		Endpoint: subscription.Endpoint,
		Keys: webpush.Keys{
			P256dh: subscription.Keys.P256dh,
			Auth:   subscription.Keys.Auth,
		},
	}

	resp, err := webpush.SendNotification([]byte(message), s, &webpush.Options{
		VAPIDPublicKey:  config.AppConfig.VapidPublicKey,
		VAPIDPrivateKey: config.AppConfig.VapidPrivateKey,
		TTL:             30,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func GetVAPIDPublicKey() string {
	return config.AppConfig.VapidPublicKey
}
