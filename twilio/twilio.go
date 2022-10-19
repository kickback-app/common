package twilio

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"unicode"

	"github.com/kickback-app/common/log"
	"github.com/kickback-app/common/utils/gsmutils"
	"github.com/twilio/twilio-go"
	twilioapi "github.com/twilio/twilio-go/rest/api/v2010"
	openapi "github.com/twilio/twilio-go/rest/verify/v2"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const phonenumberRegex = `^\+\d{11,}$`

type TwilioPublisher interface {
	CreateMessage(params *twilioapi.CreateMessageParams) (*twilioapi.ApiV2010Message, error)
}

type TwilioVerifier interface {
	CreateVerification(ServiceID string, params *openapi.CreateVerificationParams) (*openapi.VerifyV2Verification, error)
	CreateVerificationCheck(ServiceID string, params *openapi.CreateVerificationCheckParams) (*openapi.VerifyV2VerificationCheck, error)
}

type Client struct {
	logger          log.Logger
	twilioclient    TwilioPublisher
	twilioVerifier  TwilioVerifier
	accountSID      string
	accountToken    string
	verifyServiceID string
	fromPhonenumber string
}

type ClientParams struct {
	Logger          log.Logger
	Publisher       TwilioPublisher
	Verifier        TwilioVerifier
	AccountSID      string
	AccountToken    string
	VerifyServiceID string
	FromPhonenumber string
}

func NewClient(params *ClientParams) (*Client, error) {
	if params.AccountSID == "" {
		return nil, errors.New("accountSID cannot be empty")
	}
	if params.AccountToken == "" {
		return nil, errors.New("accountToken cannot be empty")
	}
	if !IsValidPhoneNumber(params.FromPhonenumber) {
		return nil, fmt.Errorf("from phonenumber (%v) is invalid", params.FromPhonenumber)
	}
	var twilioclient TwilioPublisher
	twiliRestClient := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: params.AccountSID,
		Password: params.AccountToken,
	})
	twilioclient = twiliRestClient.Api
	if params.Publisher != nil {
		twilioclient = params.Publisher
	}
	var logger log.Logger
	logger = log.StdOutLogger{}
	if params.Logger != nil {
		logger = params.Logger
	}
	var twilioverifier TwilioVerifier
	twilioverifier = twiliRestClient.VerifyV2
	if params.Verifier != nil {
		twilioverifier = params.Verifier
	}
	return &Client{
		logger: logger,

		twilioclient:    twilioclient,
		twilioVerifier:  twilioverifier,
		accountSID:      params.AccountSID,
		accountToken:    params.AccountToken,
		verifyServiceID: params.VerifyServiceID,
		fromPhonenumber: params.FromPhonenumber,
	}, nil
}

// SendSMS invokes twilio API to send sms message following their docs
// https://www.twilio.com/docs/sms/quickstart/go
func (client *Client) SendSMS(msg, phonenumber string) error {
	if !IsValidPhoneNumber(phonenumber) {
		return fmt.Errorf("phonenumber (%v) is invalid", phonenumber)
	}
	msg, err := NormalizeTwilioMsg(msg)
	if err != nil {
		client.logger.Error("unable to normalize sms message: %v", err)
		return err
	}

	params := &twilioapi.CreateMessageParams{}
	params.SetBody(msg)
	params.SetFrom(client.fromPhonenumber)
	params.SetTo(phonenumber)

	resp, err := client.twilioclient.CreateMessage(params)
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

// SendOtp twilio OTP setup
// https://www.twilio.com/blog/sms-one-time-passcode-verification-otp-2fa-golang
func (client *Client) SendOtp(phonenumber string) error {
	if ok := IsValidPhoneNumber(phonenumber); !ok {
		return InvalidPhonenumberError{Phonenumber: phonenumber}
	}
	params := &openapi.CreateVerificationParams{}
	params.SetTo(phonenumber)
	params.SetChannel("sms")

	resp, err := client.twilioVerifier.CreateVerification(client.verifyServiceID, params)
	if err != nil {
		client.logger.Error("unable to send otp: %v", err)
		return err
	}
	client.logger.Info("successfully sent otp to %v: %v", *resp.Sid)
	return nil
}

func (client *Client) CheckOtp(phonenumber, code string) (bool, error) {
	params := &openapi.CreateVerificationCheckParams{}
	params.SetTo(phonenumber)
	params.SetCode(code)

	resp, err := client.twilioVerifier.CreateVerificationCheck(client.verifyServiceID, params)
	if err != nil {
		client.logger.Error("unable to check otp: %v", err)
		return false, err
	}
	if *resp.Status == "approved" {
		client.logger.Info("passed otp check")
		return true, nil
	}
	client.logger.Error("invalid otp code")
	return false, InvalidOtpCodeErr{}
}

// IsValidPhoneNumber returns true if the phone number is valid according to twilio rules
func IsValidPhoneNumber(phonenumber string) bool {
	if matches, _ := regexp.MatchString(phonenumberRegex, phonenumber); matches {
		return true
	}
	return false
}

/*
 * Associated errors
 *
 */

type InvalidPhonenumberError struct {
	Phonenumber string
}

func (e InvalidPhonenumberError) Error() string {
	return fmt.Sprintf("invalid phonenumber: %s: should match format of +15104158622", e.Phonenumber)
}

func (e InvalidPhonenumberError) Code() int {
	return http.StatusBadRequest
}

type ThrottleError struct {
	RetryAfter string
}

func (e ThrottleError) Error() string {
	if e.RetryAfter != "" {
		return fmt.Sprintf("too many requests: retry after %v", e.RetryAfter)
	}
	return "too many requests"
}

func (e ThrottleError) Code() int {
	return http.StatusTooManyRequests
}

type InvalidOtpCodeErr struct{}

func (e InvalidOtpCodeErr) Error() string {
	return "imvalid otp code"
}

func (e InvalidOtpCodeErr) Code() int {
	return http.StatusBadRequest
}
