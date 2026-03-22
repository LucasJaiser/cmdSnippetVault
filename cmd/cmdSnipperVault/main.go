package main

import (
	"fmt"
	"lucasjaiser/goSnipperVault/internal/config"
	"lucasjaiser/goSnipperVault/internal/service"
	"lucasjaiser/goSnipperVault/internal/storage/sqlite"
)

func main() {

	err := config.BindFlags(nil)
	if err != nil {
		fmt.Printf("error occurred: %s", err.Error())
		return
	}

	cfg, err := config.InitConfig(nil)

	if err != nil {
		fmt.Printf("error occurred: %s", err.Error())
		return
	}

	repo, err := sqlite.New(cfg.DatabasePath)
	if err != nil {
		fmt.Printf("error occurred: %s", err.Error())
		return
	}

	defer repo.Close()

	_ = service.NewSnippetService(repo, nil, cfg)

}
