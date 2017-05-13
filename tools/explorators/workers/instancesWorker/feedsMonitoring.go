package instancesWorker

import (
	"encoding/json"
	"fmt"
	"github.com/sapiens-sapide/go-mastodon"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

// Connect to Instance's local feed via websocket
// and save all unknown usernames and instances seen.
func (iw *InstanceWorker) WSLocalFeedMonitoring() {
	fmt.Printf("starting local feed WS monitoring for %s\n", iw.Instance.Domain)
	c := mastodon.NewClient(&mastodon.Config{
		Server:       "https://" + iw.Instance.Domain,
		ClientID:     iw.Instance.APIid,
		ClientSecret: iw.Instance.APIsecret,
	})
	// Loop to reconnect if connection closed
	for {
		err := c.Authenticate(iw.Context, iw.Instance.Username, iw.Instance.Password)
		if err != nil {
			log.Printf("[MonitorInstanceFeed] : auth against instance %s failed with error : %s\n", iw.Instance.Domain, err)
			//TODO: fallback to fetch timeline monitoring
			iw.Instance.IsAuthorized = false
			iw.Instance.APIid = ""
			iw.Instance.APIsecret = ""
			iw.Backend.SaveInstance(iw.Instance)
			return
		}

		wsClient := c.NewWSClient()
		publicStream, _ := wsClient.StreamingWSPublic(iw.Context, true)

		for evt := range publicStream {
			var acc mastodon.Account
			switch e := evt.(type) {
			case *mastodon.NotificationEvent:
				acc = e.Notification.Account
			case *mastodon.UpdateEvent:
				acc = e.Status.Account
			default:
				continue
			}

			iw.SaveIfUnknown(acc)
			// need to subhub to instance's local feed if it is a new one
		}
	}
}

// Fetch Instance's local feed via REST API
// and save all unknown usernames and instances seen
func (iw *InstanceWorker) APILocalFeedMonitoring() {
	fmt.Printf("starting local feed API monitoring for %s\n", iw.Instance.Domain)
	var timeFrame time.Duration
	u := url.URL{
		Scheme:   "https",
		Host:     iw.Instance.Domain,
		Path:     "/api/v1/timelines/public",
		RawQuery: "local=true&limit=50",
	}
	for {
		resp, err := http.Get(u.String())
		if err == nil {
			body, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err == nil {
				var statuses []mastodon.Status
				if err = json.Unmarshal(body, &statuses); err == nil {
					if len(statuses) > 10 { // don't lose time with sleeping instances…
						oldest := time.Now()
						most_recent := time.Now()
						for _, status := range statuses {
							iw.SaveIfUnknown(status.Account)
							if status.CreatedAt.After(most_recent) {
								most_recent = status.CreatedAt
							}
							if status.CreatedAt.Before(oldest) {
								oldest = status.CreatedAt
							}
						}
						timeFrame = most_recent.Sub(oldest)
					} else {
						timeFrame = 17 * time.Hour
					}
				}
			}
		}
		timeFrame = time.Duration(float64(timeFrame) * 0.75)
		if timeFrame > (12 * time.Hour) {
			timeFrame = 12 * time.Hour
		}
		time.Sleep(timeFrame)
	}
	/*
		- fetch json from instance's URL
		- partially unmarshal json to retrieve toots' id and accounts
		- lookup worker's toots map
		- if toot not found, add it to worker's map and launch 'SaveIfUnknown' func
		- cleanup vars (json, etc.)
	*/
}

/*
get public instance timeline
parse json to retreive a list of users' accounts belonging to the instance
save accounts
*/
