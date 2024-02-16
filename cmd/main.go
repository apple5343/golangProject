package main

import (
	"fmt"
	"net/http"
	"server/config"
	"server/server"
	"server/server/db"
)

func main() {
	if err := run(); err != nil {
		return
	}
}

func run() error {
	config, err := config.InitConfig()
	if err != nil {
		return err
	}
	database, err := db.NewDatabase(*config)
	if err != nil {
		return err
	}
	s, err := server.Init(database, config)
	if err != nil {
		return err
	}
	httpServer := &http.Server{
		Addr:    config.Server.Port,
		Handler: s,
	}
	s.Calculator.ContinueCalculations()
	fmt.Printf("Config: %+v\n", config)
	if err := httpServer.ListenAndServe(); err != nil {
		return err
	}
	return nil
}
