package main

import (
	"fmt"
	"lucasjaiser/goSnipperVault/internal/storage/sqlite"
)

func main() {
	repo, err := sqlite.New("cmdvault.db")
	if err != nil {
		fmt.Printf("error occurred: %s", err.Error())
		return
	}

	defer repo.Close()
}
