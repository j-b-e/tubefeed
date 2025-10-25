package meta

import (
	"fmt"
	"log/slog"
	"tubefeed/internal/models"
	"tubefeed/internal/provider"
	"tubefeed/internal/provider/registry"
	"tubefeed/internal/utils"

	"github.com/google/uuid"
)

type Source struct {
	provider provider.SourceProvider
	Status   models.Status
	Meta     provider.SourceMeta
	ID       uuid.UUID
	logger   *slog.Logger
}

type VideoProviderList map[string]provider.ProviderNewSourceFn

type Provider struct {
	List VideoProviderList
}

func (vm *Source) Download(path string) error {
	if vm.provider == nil {
		domain, err := utils.ExtractDomain(vm.Meta.URL)
		if err != nil {
			return err
		}
		new := registry.Get(domain)
		if new == nil {
			return fmt.Errorf("failed to Download")
		}
		provider, err := new(vm.Meta.URL, vm.logger)
		if err != nil {
			return err
		}
		vm.provider = provider
	}
	return vm.provider.Download(vm.ID, path)
}

func NewSource(id uuid.UUID, url string, logger *slog.Logger) (Source, error) {
	domain, _ := utils.ExtractDomain(url)
	new := registry.Get(domain)
	if new == nil {
		return Source{}, fmt.Errorf("domain not supported: %s", domain)
	}
	prov, err := new(url, logger)
	if err != nil {
		return Source{}, err
	}

	meta := provider.SourceMeta{
		URL:   prov.Url(),
		Title: "Loading...",
	}
	return Source{
		ID:       id,
		Meta:     meta,
		provider: prov,
		Status:   models.StatusNew,
		logger:   logger.With("id", id),
	}, nil
}

func (vm *Source) LoadMeta() error {

	videomd, err := vm.provider.LoadMetadata()
	if err != nil {
		return err
	}
	vm.Meta = *videomd

	return nil
}
