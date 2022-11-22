package applinks

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/kickback-app/common/log"
	"github.com/kickback-app/common/request"
)

const defaultAppURL = "https://kickbackapp.page.link/89eQ"

var googleCloudAPIKey = os.Getenv("GOOGLE_CLOUD_API_KEY")

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

	request := request.DefaultR(http.DefaultClient)
	request.SetHeader("Content-Type", "application/json")
	request.SetResult(&result) // Unmarshal response into struct automatically if status code >= 200 and <= 299.
	request.SetBody(body)

	resp, err := request.Post(url)

	if err != nil {
		fmt.Println("IN Err clause")
		log.Logger.Error(nil, "unable to create dynamicLinkInfo: %v", err)
		return defaultAppURL
	}
	if resp.IsError() {
		fmt.Println("In resp.IsErr clause")
		log.Logger.Error(nil, "unable to create dynamicLinkInfo due to bad status code (%v): %v", resp.StatusCode, resp.Error())
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
