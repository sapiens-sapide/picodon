package accounts_explorator

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/sapiens-sapide/go-mastodon"
	"log"
	"time"
)

// worker that continuously goes through accounts in db
// to retreive accounts' relationships of an instance and
// save new discovered users and instances.
func (inst *Instance) ScanUsers(ctx context.Context, bck Backend) {
	fmt.Printf("starting ScanUsers worker for %s\n", inst.Domain)
	c := mastodon.NewClient(&mastodon.Config{
		Server:       "https://" + inst.Domain,
		ClientID:     inst.APIid,
		ClientSecret: inst.APIsecret,
	})
	for {
		err := c.Authenticate(ctx, inst.Username, inst.Password)
		if err != nil {
			log.Printf("[ScanInstanceUsers] : auth against instance %s failed with error : %s", inst.Domain, err)
		}
		accounts, err := bck.FindAccountsToScan(inst)
		for _, account := range accounts {
			followers, err := c.GetAccountFollowers(ctx, int64(account.ID))
			if err != nil {
				log.Printf("[ScanInstanceUsers] error when getting followers for account %d : %s", account.ID, err)
			} else {
				account.LocalFollowers, account.RemoteFollowers = inst.iterateAccounts(account.ID, followers, bck)
			}
			followings, err := c.GetAccountFollowing(ctx, int64(account.ID))
			if err != nil {
				log.Printf("[ScanInstanceUsers] error when getting followings for account %d : %s", account.ID, err)
			} else {
				account.LocalFollowings, account.RemoteFollowings = inst.iterateAccounts(account.ID, followings, bck)
			}
			account.LastScan = time.Now()
			bck.SaveAccount(account)
		}

		time.Sleep(5 * time.Minute)
	}
}

func (inst *Instance) iterateAccounts(accountID uint, accts []*mastodon.Account, bck Backend) (local, remote uint) {
	for _, mastodonAcct := range accts {
		user, instance, err := splitUserAndInstance(mastodonAcct.Acct, inst.Domain)
		if err != nil {
			fmt.Printf("error : %s\n", err)
			continue
		}
		acct := Account{
			Model:    gorm.Model{ID: uint(mastodonAcct.ID)},
			Username: user,
			Instance: instance,
		}
		bck.SaveAccount(acct)
		bck.SaveInstance(Instance{Domain: instance})
		if instance == inst.Domain {
			local++
		} else {
			remote++
		}
	}
	return
}
