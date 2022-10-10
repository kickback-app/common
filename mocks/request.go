package mocks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
)

type requestMock struct {
	callCount         int
	responses         []*http.Response
	errors            []error
	validators        []RequestValidator
	defaultStatusCode int
}

func (m *requestMock) CallCount() int { return m.callCount }

func (m *requestMock) ResetCallCount() { m.callCount = 0 }

type NewRequestMockOpts struct {
	Responses         []*http.Response
	Errors            []error
	Validators        []RequestValidator
	DefaultStatusCode int
}

func NewRequestMock(opts *NewRequestMockOpts) *requestMock {
	return &requestMock{
		callCount:         0,
		responses:         opts.Responses,
		errors:            opts.Errors,
		defaultStatusCode: opts.DefaultStatusCode,
		validators:        opts.Validators,
	}
}

type RequestValidator struct {
	Name string // easier for identification on error

	ExpectedURLPath string
	ExpectedMethod  string

	ExpectedCalledWith map[string]interface{}
}

func (v *RequestValidator) validate(req *http.Request) error {
	bodyBytes := make([]byte, 0)
	if req.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(req.Body) // we swallow error so the bodyBytes may be an empty array
	}
	// body := string(bodyBytes)
	if v.ExpectedCalledWith != nil {
		var container map[string]interface{}
		err := json.Unmarshal(bodyBytes, &container)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(container, v.ExpectedCalledWith) {
			return validationError{Reason: "objects did not match expected called with evaluation"}
		}
	}
	if v.ExpectedMethod != "" {
		if v.ExpectedMethod != req.Method {
			return validationError{Reason: fmt.Sprintf("unexpected method (name: %v)", v.Name)}
		}
	}
	if v.ExpectedURLPath != "" {
		path := req.URL.Path
		if v.ExpectedURLPath != path {
			return validationError{Reason: fmt.Sprintf("unexpected URL path (name: %v)", v.Name)}
		}
	}
	return nil
}

func (m *requestMock) Do(req *http.Request) (*http.Response, error) {
	currAttempt := m.callCount
	var resp *http.Response
	var err error
	if currAttempt < len(m.responses) {
		resp = m.responses[currAttempt]
	}
	if currAttempt < len(m.errors) {
		err = m.errors[currAttempt]
	}
	if currAttempt < len(m.validators) {
		err = m.validators[currAttempt].validate(req)
	}
	m.callCount++
	return resp, err
}

type validationError struct {
	Reason string
}

func (e validationError) Error() string {
	return "validation did not pass: " + e.Reason
}
