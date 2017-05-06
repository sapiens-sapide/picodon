package explorators

import (
	"context"
	"github.com/jinzhu/gorm"
	"time"
)

// our model for a mastodon user account
type Account struct {
	gorm.Model
	Username         string
	Instance         string    `gorm:"primary_key"`
	LastScan         time.Time // last time our worker scanned account's relationships
	LocalFollowers   uint
	LocalFollowings  uint
	RemoteFollowers  uint
	RemoteFollowings uint
}

type Instance struct {
	Domain            string `gorm:"primary_key"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time `sql:"index"`
	Username          string
	Password          string
	APIid             string
	APIsecret         string
	Is_registered     bool      // whether an account is created in the instance
	Is_authorized     bool      // whether an API key/secret has been gained from instance
	Users_count       uint      // number of accounts publicly announced by instance's API
	Last_count        time.Time // last time a user_count was obtained from instance's API
	Last_count_failed bool      // whether last try to get user_count was successfull
}

type InstanceWorker struct {
	Backend  Backend
	Context  context.Context
	Instance Instance
}
