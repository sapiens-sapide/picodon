package accounts_explorator

import (
	"fmt"
	"log"
	"github.com/sapiens-sapide/go-mastodon"

	"github.com/jinzhu/gorm"
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
			acct := Account {
				Model: gorm.Model{ID: uint(acc.ID)},
				Username: user,
				Instance: instance,
			}
			iw.Backend.SaveAccount(acct)
			iw.Backend.SaveInstance(Instance{Domain: instance})
		}
	}
}