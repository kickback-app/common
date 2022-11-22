package applinks

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/imroc/req/v3"
	"github.com/kickback-app/common/log"
)

const defaultAppURL = "https://kickbackapp.page.link/89eQ"

var googleCloudAPIKey = os.Getenv("GOOGLE_CLOUD_API_KEY")
var client = req.C()

func DynamicAppURL(eventID string) string {
	url := fmt.Sprintf("https://firebasedynamiclinks.googleapis.com/v1/shortLinks?key=%v", googleCloudAPIKey)
	body := map[string]interface{}{
		"dynamicLinkInfo": map[string]interface{}{
			"domainUriPrefix": "https://kickbackapp.page.link",
			"link":            fmt.Sprintf("https://kickbackapp.io/invited?eventId=%v", eventID),
			"iosInfo": map[string]interface{}{
				"iosBundleId": "com.kickbackapp",
			},
		},
	}
	var result struct {
		ShortLink   string `json:"shortLink"`
		PreviewLink string `json:"previewLink"`
	}
	var errRes interface{}
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&result). // Unmarshal response into struct automatically if status code >= 200 and <= 299.
		SetError(&errRes).
		SetRetryCount(2).
		SetRetryFixedInterval(2 * time.Second).
		AddRetryCondition(func(resp *req.Response, err error) bool {
			return resp.StatusCode >= 500
		}).
		SetBody(body).
		Post(url)
	if err != nil {
		fmt.Println("IN Err clause")
		log.Logger.Error(nil, "unable to create dynamicLinkInfo: %v", err)
		return defaultAppURL
	}
	if resp.IsError() {
		fmt.Println("In resp.IsErr clause")
		log.Logger.Error(nil, "unable to create dynamicLinkInfo due to bad status code (%v): %v", resp.StatusCode, errRes)
		return defaultAppURL
	}
	return result.ShortLink
}

func BuildDynamicLink(eventId string, userId string) string {
	link := url.QueryEscape(fmt.Sprintf("https://kickbackapp.io/invited?eventId=%v&userId=%v", eventId, userId))

	return fmt.Sprintf("https://kickbackapp.page.link/?link=%v&ibi=com.kickbackapp&isi=1607393773", link)
}

func BuildInviteWebAppLink(eventId string, userId string) string {
	return fmt.Sprintf("https://kickbackapp.io/invite?eventid=%v&userid=%v", eventId, userId)
}
