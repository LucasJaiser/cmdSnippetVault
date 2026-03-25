package cli

import (
	"lucasjaiser/goSnipperVault/internal/service"
	"lucasjaiser/goSnipperVault/internal/storage/sqlite"
)

func getService() error {
	if snippetService != nil {
		return nil
	}

	if appCfg == nil {
		InitEssential()
	}

	repo, err := sqlite.New(appCfg.DatabasePath)
	if err != nil {
		return err
	}

	Cleanup = func() { repo.Close() }

	snippetService = service.NewSnippetService(repo, nil)

	return nil
}
