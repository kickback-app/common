package applinks

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/kickback-app/common/log"
	"github.com/kickback-app/common/request"
)

const defaultAppURL = "https://kickbackapp.page.link/89eQ"

var googleCloudAPIKey = os.Getenv("GOOGLE_CLOUD_API_KEY")

var DefaultClient request.HTTPClient = &http.Client{Timeout: 15 * time.Second}

var fireBaseUrl = fmt.Sprintf("https://firebasedynamiclinks.googleapis.com/v1/shortLinks?key=%v", googleCloudAPIKey)

func DynamicAppURL(eventID string) string {

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

	var reason struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Status  string `json:"status"`
			Details []struct {
				Type     string `json:"@type"`
				Reason   string `json:"reason"`
				Domain   string `json:"domain"`
				Metadata struct {
					Service string `json:"service"`
				} `json:"metadata"`
			} `json:"details"`
		} `json:"error"`
	}

	request := request.DefaultR(DefaultClient)
	request.SetHeader("Content-Type", "application/json")
	request.SetResult(&result) // Unmarshal response into struct automatically if status code >= 200 and <= 299.
	request.SetBody(body)
	request.SetReason(&reason)

	resp, err := request.Post(fireBaseUrl)

	if err != nil {
		log.Logger.Error(nil, "unable to create dynamicLinkInfo: %v", err)
		return defaultAppURL
	}
	if resp.IsError() {
		log.Logger.Error(nil, "unable to create dynamicLinkInfo due to bad status code (%v): %v", resp.StatusCode, reason)
		return defaultAppURL
	}
	return result.ShortLink
}

func BuildDynamicLink(eventId string, userId string) string {
	link := fmt.Sprintf("https://kickbackapp.io/invited?eventId=%v&userId=%v", eventId, userId)
	return fmt.Sprintf("https://kickbackapp.page.link/?link=%v&ibi=com.kickbackapp&isi=1607393773", link)
}

func BuildInviteWebAppLink(eventId string, userId string) string {
	return fmt.Sprintf("https://kickbackapp.io/invite?eventid=%v&userid=%v", eventId, userId)
}
