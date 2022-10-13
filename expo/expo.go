package expo

import (
	"github.com/kickback-app/common/log"
	expoApi "github.com/navivix/exponent-server-sdk-golang/sdk"
)

type ExpoPublisher interface {
	Publish(message *expoApi.PushMessage) ([]expoApi.PushResponse, error)
}

type Client struct {
	logger    log.Logger
	publisher ExpoPublisher
}

func NewClient(logger log.Logger, publisher ExpoPublisher) *Client {
	var expopublisher ExpoPublisher
	expopublisher = expoApi.NewPushClient(nil)
	if publisher != nil {
		expopublisher = publisher
	}
	return &Client{
		logger:    logger,
		publisher: expopublisher,
	}
}

type Notification struct {
	ExpoPushTokens []string
	Title          string
	Body           string
	Data           map[string]string
	Sound          string
	Priority       string
}

type ExpoResult struct {
	Sent          int
	Failed        int
	InvalidTokens int
}

func (client *Client) SendPushNotification(msg *Notification) (*ExpoResult, error) {
	if len(msg.ExpoPushTokens) == 0 {
		client.logger.Info("no recipients set in the message...skipping")
		return nil, nil
	}

	invalidtokens := 0
	pushTokens := []expoApi.ExponentPushToken{}
	for _, etoken := range msg.ExpoPushTokens {
		token, err := expoApi.NewExponentPushToken(etoken)
		if err != nil {
			invalidtokens++
			continue
		}
		pushTokens = append(pushTokens, token)
	}
	sound := msg.Sound
	if sound == "" {
		sound = "default"
	}
	priority := msg.Priority
	if priority == "" {
		priority = expoApi.DefaultPriority
	}
	expoMsg := &expoApi.PushMessage{
		To:       pushTokens,
		Title:    msg.Title,
		Body:     msg.Body,
		Data:     msg.Data,
		Sound:    sound,
		Priority: priority}
	responses, err := client.publisher.Publish(expoMsg)
	if err != nil {
		client.logger.Error("unable to publish push notification: %v", err)
		return nil, err
	}
	failed := 0
	for _, res := range responses {
		client.logger.Debug("push notification details: %+v", res.Details)
		if err := res.ValidateResponse(); err != nil {
			failed++
			client.logger.Error("unable to to send push notification to %v: %v", res.PushMessage.To, err)
		}
	}
	client.logger.Info("sent push notifications to %v recipients", len(responses)-failed)
	return &ExpoResult{
		Sent:          len(responses) - failed,
		Failed:        failed,
		InvalidTokens: invalidtokens,
	}, nil
}
