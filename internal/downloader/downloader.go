package downloader

import (
	"fmt"
	"log/slog"
	"os"
	"tubefeed/internal/models"
	"tubefeed/internal/provider"
	"tubefeed/internal/provider/registry"
	"tubefeed/internal/utils"

	"github.com/google/uuid"
)

type Source struct {
	provider provider.SourceProvider
	URL      string
	ID       uuid.UUID
	logger   *slog.Logger
}

type ProviderList map[string]provider.ProviderNewSourceFn

type Provider struct {
	List ProviderList
}

func (vm *Source) Download(path string) error {
	if vm.provider == nil {
		domain, err := utils.ExtractDomain(vm.URL)
		if err != nil {
			return err
		}
		new := registry.Get(domain)
		if new == nil {
			return fmt.Errorf("failed to Download")
		}
		provider, err := new(vm.URL, vm.logger)
		if err != nil {
			return err
		}
		vm.provider = provider
	}
	_, err := os.Stat(path)
	if err != nil {
		return err
	}
	return vm.provider.Download(vm.ID, path)
}

// NewSource initializes a new downloader based on the URL domain
func NewSource(id uuid.UUID, url string, logger *slog.Logger) (Source, error) {
	domain, err := utils.ExtractDomain(url)
	if err != nil {
		return Source{}, err
	}
	downloader := registry.Get(domain)
	if downloader == nil {
		return Source{}, fmt.Errorf("domain not supported: %s", domain)
	}
	prov, err := downloader(url, logger)
	if err != nil {
		return Source{}, err
	}
	return Source{
		ID:       id,
		URL:      url,
		provider: prov,
		logger:   logger.With("id", id),
	}, nil
}

// LoadMeta loads metadata for the given request
func (vm *Source) LoadMeta(request *models.Request) error {

	metadata, err := vm.provider.LoadMetadata()
	if err != nil {
		return err
	}
	request.Title = metadata.Title
	request.Status = models.StatusMeta
	return nil
}
