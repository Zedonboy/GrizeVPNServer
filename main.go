package main

import (
	"log"
	"os"
	"sync"

	"github.com/xitongsys/pangolin/config"
	pangolinServer "github.com/xitongsys/pangolin/server"
	"github.com/zedonboy/grize-vpn-server/server"
)

func main() {
	config := config.NewConfig()
	lm, err := pangolinServer.NewLoginManager(config)
	if err != nil {
		log.Printf("Login Manger could not be instantiated")
		os.Exit(1)
	}

	us, err := server.NewInstance(config.ServerAddr, lm)
	if err != nil {
		log.Printf("Login Manger could not be instantiated")
		os.Exit(1)
	}

	us.StartWithAuth()

	var wg sync.WaitGroup

	wg.Add(1)
	wg.Wait()
}
