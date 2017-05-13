package instancesWorker

import (
	"fmt"
	"github.com/sapiens-sapide/go-mastodon"
	expl "github.com/sapiens-sapide/picodon/tools/explorators"
	"log"
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
}

func (iw *InstanceWorker) SaveIfUnknown(acc mastodon.Account) (acct expl.Account, NewAccount, NewInstance bool) {
	//TODO: optimize lookup by maintaining in-memory accounts and instances tables
	user, instance, err := expl.SplitUserAndInstance(acc.Acct, iw.Instance.Domain)
	if err != nil {
		fmt.Printf("error : %s\n", err)
		return
	}
	acct = expl.Account{
		Username: user,
		Instance: instance,
	}
	if instance != iw.Instance.Domain {
		id, err := expl.GetRemoteAccountID(user, instance)
		if err == nil {
			acct.ID = uint(id)
		}
	} else {
		acct.ID = uint(acc.ID)
	}
	if acct.ID != 0 {
		iw.Backend.CreateAccountIfNotExist(acct)
	}
	iw.Backend.CreateInstanceIfNotExist(expl.Instance{Domain: instance})

	//TODO: feedback newaccount&newinstance
	return
}

/*
get public instance timeline
parse json to retreive a list of users' accounts belonging to the instance
save accounts
*/
