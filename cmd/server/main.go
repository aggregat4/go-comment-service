package main

import (
	"aggregat4/go-commentservice/internal/domain"
	"aggregat4/go-commentservice/internal/repository"
	"aggregat4/go-commentservice/internal/server"
	"flag"
	"fmt"
	"github.com/aggregat4/go-baselib/lang"
	"github.com/kirsle/configdir"
	"github.com/kkyr/fig"
	"log"
)

func main() {
	var configFileLocation string
	flag.StringVar(&configFileLocation, "config", "", "The location of the configuration file if you do not want to default to the standard location")
	flag.Parse()
	defaultConfigLocation := configdir.LocalConfig("commentservice") + "/commentservice.json"

	var config domain.Config
	err := fig.Load(&config, fig.File(lang.IfElse(configFileLocation == "", defaultConfigLocation, configFileLocation)))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", config)

	var store repository.Store
	err = store.InitAndVerifyDb(repository.CreateFileDbUrl(config.DatabaseFilename))
	if err != nil {
		log.Fatalf("Error initializing database: %s", err)
	}
	defer store.Close()
	server.RunServer(
		server.Controller{
			Store:  &store,
			Config: config,
		})
}
