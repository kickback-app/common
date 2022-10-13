package twilio_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/kickback-app/common/log"
	"github.com/kickback-app/common/twilio"
	"github.com/stretchr/testify/require"
	twilioapi "github.com/twilio/twilio-go/rest/api/v2010"
)

func TestMain(m *testing.M) {
	exitVal := m.Run()
	os.Exit(exitVal)
}

type mockTwilioClient struct {
	resp *twilioapi.ApiV2010Message
	err  error
}

func (mtc mockTwilioClient) CreateMessage(params *twilioapi.CreateMessageParams) (*twilioapi.ApiV2010Message, error) {
	return mtc.resp, mtc.err
}

func TestClientInitMissingAccountSID(t *testing.T) {
	_, err := twilio.NewClient(&twilio.ClientParams{
		AccountSID:      "",
		AccountToken:    "mock",
		FromPhonenumber: "+15104148622",
	})
	require.NotNil(t, err, "there should be an err for missing required val")
}

func TestClientInitMissingAccountToken(t *testing.T) {
	_, err := twilio.NewClient(&twilio.ClientParams{
		AccountSID:      "mock",
		AccountToken:    "",
		FromPhonenumber: "+15104148622",
	})
	require.NotNil(t, err, "there should be an err for missing required val")
}

func TestClientInitMissingFromPhonenumber(t *testing.T) {
	_, err := twilio.NewClient(&twilio.ClientParams{
		AccountSID:      "mock",
		AccountToken:    "mock",
		FromPhonenumber: "",
	})
	require.NotNil(t, err, "there should be an err for missing required val")
}

func TestCanSendSMS(t *testing.T) {
	sid := "mockSid"
	client, err := twilio.NewClient(&twilio.ClientParams{
		Logger:          log.StdOutLogger{},
		AccountSID:      "mockaccountsid",
		AccountToken:    "mockaccounttoken",
		FromPhonenumber: "+15104148622",
		Publisher: mockTwilioClient{
			resp: &twilioapi.ApiV2010Message{
				Sid: &sid,
			},
			err: nil,
		},
	})
	require.Nil(t, err, "new client err should be nil")
	err = client.SendSMS("hey", "+15104148611")
	require.Nil(t, err, "send sms err should be nil")
}

func TestInvalidPhonenumber(t *testing.T) {
	sid := "mockSid"
	client, err := twilio.NewClient(&twilio.ClientParams{
		Logger:          log.StdOutLogger{},
		AccountSID:      "mockaccountsid",
		AccountToken:    "mockaccounttoken",
		FromPhonenumber: "+15104148622",
		Publisher: mockTwilioClient{
			resp: &twilioapi.ApiV2010Message{
				Sid: &sid,
			},
			err: nil,
		},
	})
	require.Nil(t, err, "new client err should be nil")
	err = client.SendSMS("invalid phone number", "15104148611")
	require.NotNil(t, err, "should be an err if invlaid phonenumber (missing +) ")
}

func TestSMSErr(t *testing.T) {
	client, err := twilio.NewClient(&twilio.ClientParams{
		Logger:          log.StdOutLogger{},
		AccountSID:      "mockaccountsid",
		AccountToken:    "mockaccounttoken",
		FromPhonenumber: "+15104148622",
		Publisher: mockTwilioClient{
			resp: nil,
			err:  errors.New("mock err"),
		},
	})
	require.Nil(t, err, "new client err should be nil")
	err = client.SendSMS("hey", "+15104148611")
	require.NotNil(t, err, "send sms err should throw err")
}

func TestIsValidPhoneNumber(t *testing.T) {
	cases := []struct {
		phoneNumber    string
		ExpectedResult bool
	}{
		{
			"+13233",
			false,
		},
		{
			"",
			false,
		},
		{
			"15104148533",
			false,
		},
		{
			"5104148533",
			false,
		},
		{
			"+5104148533",
			false,
		},
		{
			"+15104148533",
			true,
		},
	}
	for i, c := range cases {
		result := twilio.IsValidPhoneNumber(c.phoneNumber)
		require.Equal(t, c.ExpectedResult, result, fmt.Sprintf("testing case %v", i))
	}
}
