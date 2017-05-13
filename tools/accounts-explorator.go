package main

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/sapiens-sapide/go-mastodon"
	. "github.com/sapiens-sapide/picodon/tools/explorators"
	w "github.com/sapiens-sapide/picodon/tools/explorators/workers"
	iw "github.com/sapiens-sapide/picodon/tools/explorators/workers/instancesWorker"
	"log"
	"os"
	"os/signal"
	"sync"
)

const postgres = "localhost"
const pgsqlUser = "concierge"
const pgsqlDB = "mastodon"

func main() {

	var err error
	var backend PsqlDB

	// db jobs
	backend.DB, err = gorm.Open("postgres", "host="+postgres+" dbname="+pgsqlDB+" user="+pgsqlUser+" sslmode=disable")
	backend.DB.DB().SetMaxIdleConns(10)
	backend.DB.DB().SetMaxOpenConns(1000)
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
	InstancesWorkers := make(map[string]*iw.InstanceWorker)
	nstncs := []Instance{}
	// retreive all known instances that were up within last hour
	backend.DB.Where("count_failed = false").Find(&nstncs)
	for _, nstnc := range nstncs {
		ctx := context.Background()
		var err error
		var app *mastodon.Application
		// Get auth token if missing
		if nstnc.IsRegistered && !nstnc.IsAuthorized {
			app, err = mastodon.RegisterApp(ctx, &mastodon.AppConfig{
				Server:     "https://" + nstnc.Domain,
				ClientName: "concierge-bot",
				Scopes:     "read write follow",
			})
			if err == nil {
				nstnc.APIid = app.ClientID
				nstnc.APIsecret = app.ClientSecret
				nstnc.IsAuthorized = true
				backend.SaveInstance(nstnc)
			} else {
				fmt.Println(err)
			}
		}
		InstancesWorkers[nstnc.Domain] = &iw.InstanceWorker{
			Backend:       &backend,
			Context:       ctx,
			Instance:      nstnc,
			SeenLock:      new(sync.Mutex),
			AccountsSeen:  make(map[string]bool),
			InstancesSeen: make(map[string]bool),
			IsWSConnected: false,
		}
	}

	// launch instances workers
	// TODO: manage workers start/stop/resume
	/*
		every hour :
			- retreive 'up' instances list from db
			- cancel workers for instances not retreived
			- for each instance :
				check if there is a worker currently running
				if not => create one
				if yes => check if connection is possible
					if yes => check if connected
						if not => try to connect
			if connection possible : try to connect via WS
			if connection failed or not possible : launch API fetcher
	*/

	for _, worker := range InstancesWorkers {
		if worker.Instance.IsAuthorized {
			go worker.WSLocalFeedMonitoring()
			go worker.ScanUsers()
		} else {
			go worker.APILocalFeedMonitoring()
		}
	}
	/*
		TODO :connect to authorized instances
		TODO :subhub to other instances
		TODO :better output handling
	*/
	go w.InstancesUsersCount(&backend)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	//block until a signal is received
	<-c

}
