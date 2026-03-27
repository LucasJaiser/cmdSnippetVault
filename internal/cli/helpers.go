package cli

import (
	"lucasjaiser/goSnipperVault/internal/clipboard"
	"lucasjaiser/goSnipperVault/internal/domain"
	"lucasjaiser/goSnipperVault/internal/importer"
	"lucasjaiser/goSnipperVault/internal/service"
	"lucasjaiser/goSnipperVault/internal/storage/sqlite"
	"path/filepath"
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

func getImportForFileType(filename string) domain.Import {

	extension := filepath.Ext(filename)

	switch extension {
	case "yaml":
		return importer.NewYAMLImporter()
	case "yml":
		return importer.NewYAMLImporter()
	case "json":
		return importer.NewJSONImporter()
	}

	return nil
}
