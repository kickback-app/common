package applinks

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/kickback-app/common/mocks"
	"github.com/stretchr/testify/assert"
)

func TestBuildDynamicLink(t *testing.T) {
	cases := []struct {
		in struct {
			eventId string
			userId  string
		}
		out string
	}{
		{
			struct {
				eventId string
				userId  string
			}{"", ""},
			"https://kickbackapp.page.link/?link=https://kickbackapp.io/invited?eventId=&userId=&ibi=com.kickbackapp&isi=1607393773",
		},
		{
			struct {
				eventId string
				userId  string
			}{"123", "937495"},
			"https://kickbackapp.page.link/?link=https://kickbackapp.io/invited?eventId=123&userId=937495&ibi=com.kickbackapp&isi=1607393773",
		},
		{
			struct {
				eventId string
				userId  string
			}{"68349dkoife", "dwqdwq42352"},
			"https://kickbackapp.page.link/?link=https://kickbackapp.io/invited?eventId=68349dkoife&userId=dwqdwq42352&ibi=com.kickbackapp&isi=1607393773",
		},
	}
	for _, c := range cases {
		result := BuildDynamicLink(c.in.eventId, c.in.userId)
		assert.Equal(t, c.out, result, fmt.Sprintf("testing => %+v", c.in))
	}
}

func TestBuildInviteWebAppLink(t *testing.T) {
	cases := []struct {
		in struct {
			eventId string
			userId  string
		}
		out string
	}{
		{
			struct {
				eventId string
				userId  string
			}{"", ""},
			"https://kickbackapp.io/invite?eventid=&userid=",
		},
		{
			struct {
				eventId string
				userId  string
			}{"123", "937495"},
			"https://kickbackapp.io/invite?eventid=123&userid=937495",
		},
		{
			struct {
				eventId string
				userId  string
			}{"68349dkoife", "dwqdwq42352"},
			"https://kickbackapp.io/invite?eventid=68349dkoife&userid=dwqdwq42352",
		},
	}
	for _, c := range cases {
		result := BuildInviteWebAppLink(c.in.eventId, c.in.userId)
		assert.Equal(t, c.out, result, fmt.Sprintf("testing => %+v", c.in))
	}
}

func TestDynamicAppURL(t *testing.T) {
	fireBaseUrl = "mockURL/v1/path"
	DefaultClient = mocks.NewRequestMock(&mocks.NewRequestMockOpts{
		Responses: []*http.Response{
			{
				StatusCode: 200,
				Body: ioutil.NopCloser(bytes.NewReader([]byte(`{
					"shortLink": "mockShortLink",
					"previewLink": "mockPreviewLink"
				  }`))),
			},
		},
		Validators: []mocks.RequestValidator{
			{
				ExpectedMethod:  "POST",
				ExpectedURLPath: fireBaseUrl,
			},
		},
	})

	cases := []struct {
		in struct {
			eventId string
		}
		out string
	}{
		{
			struct {
				eventId string
			}{"123"},
			"mockShortLink",
		},
	}

	for _, c := range cases {
		result := DynamicAppURL(c.in.eventId)
		assert.Equal(t, c.out, result, fmt.Sprintf("testing => %+v", c.in))
	}
}
