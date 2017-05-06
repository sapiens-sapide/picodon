package main

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/sapiens-sapide/go-mastodon"
	. "github.com/sapiens-sapide/picodon/tools/explorators"
	"log"
	"os"
	"os/signal"
)

const postgres = "localhost"
const pgsqlUser = "concierge"
const pgsqlDB = "mastodon"

func main() {

	var err error
	var backend PsqlDB

	// db jobs
	backend.DB, err = gorm.Open("postgres", "host="+postgres+" dbname="+pgsqlDB+" user="+pgsqlUser+" sslmode=disable")
	if err != nil {
		log.Fatalf("DB opening failed with error : %s", err)
	}
	defer backend.DB.Close()
	if !backend.DB.HasTable(&Account{}) {
		backend.DB.CreateTable(&Account{})
	}
	if !backend.DB.HasTable(&Instance{}) {
		backend.DB.CreateTable(&Instance{})
	}

	backend.DB.AutoMigrate(&Account{}, &Instance{}) //Migrate schemas if needed

	// instances to monitor
	instances := make(map[string]*InstanceWorker)
	nstncs := []Instance{}
	backend.DB.Where("is_registered = true").Find(&nstncs)

	for _, nstnc := range nstncs {
		ctx := context.Background()
		var err error
		var app *mastodon.Application
		if !nstnc.Is_authorized {
			app, err = mastodon.RegisterApp(ctx, &mastodon.AppConfig{
				Server:     "https://" + nstnc.Domain,
				ClientName: "concierge-bot",
				Scopes:     "read write follow",
			})
			if err == nil {
				nstnc.APIid = app.ClientID
				nstnc.APIsecret = app.ClientSecret
				nstnc.Is_authorized = true
				backend.SaveInstance(nstnc)
			} else {
				fmt.Println(err)
			}
		}
		if err == nil {
			instances[nstnc.Domain] = &InstanceWorker{
				Backend:  &backend,
				Context:  ctx,
				Instance: nstnc,
			}
		}
	}

	// launch instances workers
	// TODO: manage workers start/stop/resume
	for _, worker := range instances {
		go worker.MonitorPublicFeed()
		go worker.ScanUsers()
	}

	go InstancesUsersCount(&backend)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	//block until a signal is received
	<-c

}
