package explorators

import (
	"context"
	"github.com/jinzhu/gorm"
	"time"
)

// our model for a mastodon user account
type Account struct {
	ID               uint `gorm:"primary_key"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Username         string
	Instance         string    `gorm:"primary_key"`
	LastScan         time.Time // last time our worker scanned account's relationships
	LocalFollowers   uint
	LocalFollowings  uint
	RemoteFollowers  uint
	RemoteFollowings uint
}

type Instance struct {
	gorm.Model
	Domain       string `gorm:"primary_key"`
	Username     string
	Password     string
	APIid        string
	APIsecret    string
	IsRegistered bool      // whether an account is created in the instance
	IsAuthorized bool      // whether an API key/secret has been gained from instance
	UsersCount   uint      // number of accounts publicly announced by instance's API
	LastCount    time.Time // last time a user_count was obtained from instance's API
	CountFailed  bool      // whether last try to get user_count was successfull
}
