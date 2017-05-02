package accounts_explorator

import (
	"fmt"
	"context"
	"log"
	"github.com/sapiens-sapide/go-mastodon"

	"github.com/jinzhu/gorm"
)



// Connect to Instance's public feed via websocket
// and save all unknown usernames seen.
func (inst *Instance) MonitorPublicFeed(ctx context.Context, bck Backend) {
	fmt.Printf("starting MonitorPublicFeed worker for %s\n", inst.Domain)
	c := mastodon.NewClient(&mastodon.Config{
		Server:       "https://" + inst.Domain,
		ClientID:     inst.APIid,
		ClientSecret: inst.APIsecret,
	})
	// Loop to reconnect if connection closed
	for {
		err := c.Authenticate(ctx, inst.Username, inst.Password)
		if err != nil {
			log.Printf("[MonitorInstanceFeed] : auth against instance %s failed with error : %s\n", inst.Domain, err)
			return
		}

		wsClient := c.NewWSClient()
		publicStream, _ := wsClient.StreamingWSPublic(ctx, true)

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

			user, instance, err := splitUserAndInstance(acc.Acct, inst.Domain)
			if err != nil {
				fmt.Printf("error : %s\n", err)
				continue
			}
			acct := Account {
				Model: gorm.Model{ID: uint(acc.ID)},
				Username: user,
				Instance: instance,
			}
			bck.SaveAccount(acct)
			bck.SaveInstance(Instance{Domain: instance})
		}
	}
}