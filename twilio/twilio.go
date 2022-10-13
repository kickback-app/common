package twilio

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/kickback-app/common/log"
	"github.com/kickback-app/common/request"
	"github.com/kickback-app/common/storage"
	"github.com/kickback-app/common/utils/gsmutils"
	"github.com/twilio/twilio-go"
	twilioapi "github.com/twilio/twilio-go/rest/api/v2010"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	baseurl = "https://api.twilio.com"
)

var defaultclient = &http.Client{Timeout: 15 * time.Second}

type Client struct {
	logger          log.Logger
	dbclient        storage.Manager
	httpclient      request.HTTPClient
	accountSID      string
	accountToken    string
	fromPhonenumber string
}

type ClientParams struct {
	Logger          log.Logger
	Dbclient        storage.Manager
	Httpclient      request.HTTPClient
	AccountSID      string
	AccountToken    string
	FromPhonenumber string
}

func NewClient(params *ClientParams) (*Client, error) {
	httpclient := params.Httpclient
	if httpclient == nil {
		httpclient = defaultclient
	}
	if params.AccountSID == "" {
		return nil, errors.New("accountSID cannot be empty")
	}
	if params.AccountToken == "" {
		return nil, errors.New("accountToken cannot be empty")
	}
	if params.FromPhonenumber == "" {
		return nil, errors.New("from phonenumber cannot be empty")
	}
	return &Client{
		logger:     params.Logger,
		dbclient:   params.Dbclient,
		httpclient: httpclient,

		accountSID:      params.AccountSID,
		accountToken:    params.AccountToken,
		fromPhonenumber: params.FromPhonenumber,
	}, nil
}

// SendSMS invokes twilio API to send sms message following their docs
// https://www.twilio.com/docs/sms/quickstart/go
func (client *Client) SendSMS(msg, phonenumber string) error {
	twilioClient := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: client.accountSID,
		Password: client.accountToken,
	})

	msg, err := NormalizeTwilioMsg(msg)
	if err != nil {
		client.logger.Error("unable to normalize sms message: %v", err)
		return err
	}

	params := &twilioapi.CreateMessageParams{}
	params.SetBody(msg)
	params.SetFrom(client.fromPhonenumber)
	params.SetTo(phonenumber)

	resp, err := twilioClient.Api.CreateMessage(params)
	if err != nil {
		client.logger.Error("unable to send sms message: %v", err)
		return err
	}
	respSid := ""
	if resp.Sid != nil {
		respSid = *resp.Sid
	}
	respErrCode := 0
	if resp.ErrorCode != nil {
		respErrCode = *resp.ErrorCode
	}
	respErrMsg := ""
	if resp.ErrorMessage != nil {
		respErrMsg = *resp.ErrorMessage
	}
	client.logger.Info("twilio sms summary to %v: [sid: %v] [errCode: %v] [errMsg: %v]", phonenumber, respSid, respErrCode, respErrMsg)
	return nil
}

// NormalizeTwilioMsg removes extraneous characters from the message to avoid high costs
func NormalizeTwilioMsg(in string) (string, error) {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	r := strings.NewReader(in)
	x := transform.NewReader(r, t)
	b, err := ioutil.ReadAll(x)
	if err != nil {
		return "", err
	}
	s := gsmutils.ReplaceRunesToGSM7(string(b))
	return s, nil
}
