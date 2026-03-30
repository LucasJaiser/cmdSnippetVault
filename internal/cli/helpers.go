package cli

import (
	"lucasjaiser/goSnipperVault/internal/clipboard"
	"lucasjaiser/goSnipperVault/internal/domain"
	"lucasjaiser/goSnipperVault/internal/exporter"
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

func getImportForFileType(filename string, formatOverride string) domain.Importer {

	if formatOverride == "" {
		formatOverride = filepath.Ext(filename)
	}

	switch formatOverride {
	case "yaml", ".yaml":
		return importer.NewYAMLImporter()
	case "yml", ".yml":
		return importer.NewYAMLImporter()
	case "json", ".json":
		return importer.NewJSONImporter()
	default:
		return importer.NewJSONImporter()

	}

}

func getExporterForType(typeString string) domain.Exporter {

	switch typeString {
	case "yaml", ".yaml":
		return exporter.NewYAMLExporter()
	case "yml", ".yml":
		return exporter.NewYAMLExporter()
	case "json", ".json":
		return exporter.NewJSONExporter()
	default:
		return exporter.NewJSONExporter()
	}

}
