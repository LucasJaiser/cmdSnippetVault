package cli

import (
	"lucasjaiser/goSnipperVault/internal/clipboard"
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

	if appCfg.Clipboard {
		snippetService = service.NewSnippetService(repo, clipboard.NewSystemClipboard())
	} else {
		snippetService = service.NewSnippetService(repo, clipboard.NewNoopClipboard())
	}

	return nil
}
