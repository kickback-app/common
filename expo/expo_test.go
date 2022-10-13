package expo_test

import (
	"os"
	"testing"

	"github.com/kickback-app/common/expo"
	"github.com/kickback-app/common/log"
	expoApi "github.com/navivix/exponent-server-sdk-golang/sdk"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	exitVal := m.Run()
	os.Exit(exitVal)
}

type mockPublisher struct {
	responses []expoApi.PushResponse
	err       error
}

func (mp mockPublisher) Publish(message *expoApi.PushMessage) ([]expoApi.PushResponse, error) {
	return mp.responses, mp.err
}

func TestCanSendPushNotification(t *testing.T) {
	client := expo.NewClient(log.StdOutLogger{}, mockPublisher{
		responses: []expoApi.PushResponse{
			{Status: expoApi.SuccessStatus},
		},
		err: nil,
	})

	res, err := client.SendPushNotification(&expo.Notification{
		ExpoPushTokens: []string{"ExponentPushToken_mock"},
		Title:          "mock",
		Body:           "mock",
	})
	require.Nil(t, err, "err should be nil")
	require.Equal(t, 0, res.InvalidTokens)
	require.Equal(t, 0, res.Failed)
	require.Equal(t, 1, res.Sent)
}

func TestInvalidExpoToken(t *testing.T) {
	client := expo.NewClient(log.StdOutLogger{}, mockPublisher{
		responses: []expoApi.PushResponse{
			{Status: expoApi.SuccessStatus},
		},
		err: nil,
	})

	res, err := client.SendPushNotification(&expo.Notification{
		ExpoPushTokens: []string{"ExponentPushToken_mock", "invalid"},
		Title:          "mock",
		Body:           "mock",
	})
	require.Nil(t, err, "err should be nil")
	require.Equal(t, 1, res.InvalidTokens)
	require.Equal(t, 0, res.Failed)
	require.Equal(t, 1, res.Sent)
}

func TestInvalidFailures(t *testing.T) {
	client := expo.NewClient(log.StdOutLogger{}, mockPublisher{
		responses: []expoApi.PushResponse{
			{Status: expoApi.SuccessStatus},
			{Status: "notOK"},
		},
		err: nil,
	})

	res, err := client.SendPushNotification(&expo.Notification{
		ExpoPushTokens: []string{"ExponentPushToken_mock", "invalid"},
		Title:          "mock",
		Body:           "mock",
	})
	require.Nil(t, err, "err should be nil")
	require.Equal(t, 1, res.InvalidTokens, "invalid token count")
	require.Equal(t, 1, res.Failed, "failure count")
	require.Equal(t, 1, res.Sent, "sent count")
}
