package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/sapiens-sapide/go-mastodon"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"
)

const defaultInstance = "xxx"
const clientInstanceID = "xxx"
const clientInstanceSecret = "xxx"
const username = "xxx"
const password = "xxx"
const addrSep = "@"

const postgres = "localhost"
const pgsqlUser = "xxx"
const pgsqlDB = "xxx"

type Account struct {
	gorm.Model
	Username string
	Instance string
	LastScan time.Time // last time our worker scanned account's relationships
}

type Instance struct {
	Domain    string `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

type InstanceCredentials struct {
	Domain       string
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
}

var db *gorm.DB
var err error

func main() {

	// db jobs
	db, err = gorm.Open("postgres", "host="+postgres+" dbname="+pgsqlDB+" user="+pgsqlUser+" sslmode=disable")
	if err != nil {
		log.Fatalf("DB opening failed with error : %s", err)
	}
	defer db.Close()
	if !db.HasTable(&Account{}) {
		db.CreateTable(&Account{})
	}
	if !db.HasTable(&Instance{}) {
		db.CreateTable(&Instance{})
	}

	db.AutoMigrate(&Account{}, &Instance{}) //Migrate schemas if needed
	instance := Instance{Domain: defaultInstance}
	db.FirstOrCreate(&instance) // Add instance entry (do nothing if it exists)

	// launch stream listener
	ctx := context.Background()

	mamotInstance := InstanceCredentials{
		Domain:       defaultInstance,
		ClientID:     clientInstanceID,
		ClientSecret: clientInstanceSecret,
		Username:     username,
		Password:     password,
	}

	go monitorInstanceFeed(ctx, mamotInstance)

	go scanUsersWorker(ctx, mamotInstance)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	//block until a signal is received
	<-c

}

// Connect to Instance's public feed via websocket
// to save all unknown usernames seen.
func monitorInstanceFeed(ctx context.Context, cred InstanceCredentials) {
	c := mastodon.NewClient(&mastodon.Config{
		Server:       "https://" + cred.Domain,
		ClientID:     cred.ClientID,
		ClientSecret: cred.ClientSecret,
	})
	// Loop to reconnect if connection closed
	for {
		err = c.Authenticate(ctx, cred.Username, cred.Password)
		if err != nil {
			log.Fatalf("Authentication against mastodon instance failed with error : %s", err)
		}

		wsClient := c.NewWSClient()
		publicStream, _ := wsClient.StreamingWSPublic(ctx, true)

		for evt := range publicStream {
			var account mastodon.Account
			switch e := evt.(type) {
			case *mastodon.NotificationEvent:
				account = e.Notification.Account
			case *mastodon.UpdateEvent:
				account = e.Status.Account
			default:
				continue
			}

			user, instance, err := splitUserAndInstance(account.Acct, cred.Domain)
			if err != nil {
				fmt.Printf("error : %s\n", err)
				continue
			}
			saveAccount(account.ID, user, instance)
		}
	}
}

// worker that continuously goes through accounts in db
// to retreive accounts' relationships of an instance and
// save new discovered users and instances.
func scanUsersWorker(ctx context.Context, cred InstanceCredentials) {
	c := mastodon.NewClient(&mastodon.Config{
		Server:       "https://" + cred.Domain,
		ClientID:     cred.ClientID,
		ClientSecret: cred.ClientSecret,
	})
	// Loop to reconnect if connection closed
	for {
		err = c.Authenticate(ctx, cred.Username, cred.Password)
		aWeekAgo := time.Now().Add(-(7 * 24 * time.Hour))
		var accounts []Account
		db.Where("last_scan isnull OR last_scan < ?", aWeekAgo).Find(&accounts)
		for _, account := range accounts {
			if account.Instance == cred.Domain { // can only query instance's accounts via API.
				followers, err := c.GetAccountFollowers(ctx, int64(account.ID))
				iterateAccounts(account.ID, followers, err, cred.Domain)
				followings, err := c.GetAccountFollowing(ctx, int64(account.ID))
				iterateAccounts(account.ID, followings, err, cred.Domain)
				db.Model(&account).Update("last_scan", time.Now())
			}
		}
	}
}

func iterateAccounts(accountID uint, accts []*mastodon.Account, err error, defaultDomain string) {
	if err == nil {
		for _, mastodonAcct := range accts {
			user, instance, err := splitUserAndInstance(mastodonAcct.Acct, defaultDomain)
			if err != nil {
				fmt.Printf("error : %s\n", err)
				continue
			}
			saveAccount(mastodonAcct.ID, user, instance)
		}
	} else {
		fmt.Printf("Error when retreiving followers for account %d : %s\n", accountID, err)
	}
}

func splitUserAndInstance(acct, localInstance string) (user, instance string, err error) {
	switch strings.Count(acct, addrSep) {
	case 0:
		//a local user
		user = acct
		instance = localInstance
		return
	case 1:
		s := strings.Split(acct, addrSep)
		user = s[0]
		instance = s[1]
		return
	default:
		err = errors.New("invalid string")
		return
	}
}

func saveAccount(id int64, user, instance string) {
	account := Account{
		Username: user,
		Instance: instance,
	}
	account.ID = uint(id)
	db.FirstOrCreate(&account)
	if instance != defaultInstance {
		instance := Instance{
			Domain: instance,
		}
		db.FirstOrCreate(&instance)
	}
}
