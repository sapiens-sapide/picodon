package main

import (
	"context"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	. "github.com/sapiens-sapide/picodon/tools/accounts-explorator"
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
	instances := make(map[string]*Instance)
	nstncs := []Instance{}
	backend.DB.Where("is_registred = true AND is_authorized = true").Find(&nstncs)
	for _, nstnc := range nstncs {
		instances[nstnc.Domain] = &nstnc
	}

	// launch instances workers
	// TODO: manage workers start/stop/resume
	ctx := context.Background()
	for _, instance := range instances {
		go instance.MonitorPublicFeed(ctx, &backend)
		go instance.ScanUsers(ctx, &backend)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	//block until a signal is received
	<-c

}
