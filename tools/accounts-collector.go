package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/mattn/go-mastodon"
	"log"
	"strings"
)

const localInstance = "xxx"
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
}

type Instance struct {
	gorm.Model
	Domain string
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
	instance := Instance{Domain: localInstance}
	db.Create(&instance) // Add instance entry (do nothing if it exists)

	// launch stream listener
	ctx := context.Background()

	c := mastodon.NewClient(&mastodon.Config{
		Server:       "https://" + localInstance,
		ClientID:     clientInstanceID,
		ClientSecret: clientInstanceSecret,
	})
	err = c.Authenticate(ctx, username, password)
	if err != nil {
		log.Fatalf("Authentication against mastodon instance failed with error : %s", err)
	}

	wsClient := c.NewWSClient()
	publicStream, _ := wsClient.StreamingWSPublicLocal(ctx)

	for evt := range publicStream {

		var account mastodon.Account
		switch e := evt.(type) {
		case *mastodon.NotificationEvent:
			fmt.Printf("%+v\n", e.Notification)
			account = e.Notification.Account
		case *mastodon.UpdateEvent:
			account = e.Status.Account
		default:
			continue
		}

		user, instance, err := splitUserAndInstance(account.Acct)
		if err != nil {
			fmt.Printf("error : %s\n", err)
			continue
		}
		if instance == localInstance {
			saveAccount(account.ID, user, instance)
		}
	}
}

func splitUserAndInstance(acct string) (user, instance string, err error) {
	switch strings.Count(acct, addrSep) {
	case 0:
		//acct a local user
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
}
