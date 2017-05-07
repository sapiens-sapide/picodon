package explorators

import (
	"fmt"
	"github.com/sapiens-sapide/go-mastodon"
	"log"
)

// Connect to Instance's public feed via websocket
// and save all unknown usernames seen.
func (iw *InstanceWorker) MonitorPublicFeed() {
	fmt.Printf("starting MonitorPublicFeed worker for %s\n", iw.Instance.Domain)
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

			user, instance, err := splitUserAndInstance(acc.Acct, iw.Instance.Domain)
			if err != nil {
				fmt.Printf("error : %s\n", err)
				continue
			}
			acct := Account{
				Username: user,
				Instance: instance,
			}
			if instance != iw.Instance.Domain {
				id, err := GetRemoteAccountID(user, instance)
				if err == nil {
					acct.ID = uint(id)
				}
			} else {
				acct.ID = uint(acc.ID)
			}
			if acct.ID != 0 {
				iw.Backend.CreateAccountIfNotExist(acct)
			}
			iw.Backend.CreateInstanceIfNotExist(Instance{Domain: instance})
			// need to subhub to instance's local feed if it is a new one
		}
	}
}

/*
get public instance timeline
parse json to retreive a list of users' accounts belonging to the instance
save accounts
*/
