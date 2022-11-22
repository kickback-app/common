package applinks

import (
	"fmt"
	"net/url"
	"testing"

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
		resultEncoded := BuildDynamicLink(c.in.eventId, c.in.userId)
		result, err := url.QueryUnescape(resultEncoded)
		if err != nil {
			t.Fail()
		}
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
		resultEncoded := BuildInviteWebAppLink(c.in.eventId, c.in.userId)
		result, err := url.QueryUnescape(resultEncoded)
		if err != nil {
			t.Fail()
		}
		assert.Equal(t, c.out, result, fmt.Sprintf("testing => %+v", c.in))
	}
}
