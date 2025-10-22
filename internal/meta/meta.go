package meta

import (
	"fmt"
	"tubefeed/internal/provider"
	"tubefeed/internal/provider/registry"
	"tubefeed/internal/utils"

	"github.com/google/uuid"
)

type Source struct {
	provider provider.SourceProvider
	Status   Status
	Meta     provider.SourceMeta
	ID       uuid.UUID
}

type VideoProviderList map[string]provider.ProviderNewSourceFn

type Provider struct {
	List VideoProviderList
}

type Status string

var (
	StatusNew     Status = "New"
	StatusMeta    Status = "FetchingMeta"
	StatusLoading Status = "Downloading"
	StatusReady   Status = "Available"
	StatusError   Status = "Error"
)

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
		provider, err := new(vm.Meta.URL)
		if err != nil {
			return err
		}
		vm.provider = provider
	}
	return vm.provider.Download(vm.ID, path)
}

func NewVideo(url string) (Source, error) {
	domain, _ := utils.ExtractDomain(url)
	new := registry.Get(domain)
	prov, err := new(url)
	if err != nil {
		return Source{}, err
	}

	meta := provider.SourceMeta{
		URL:   prov.Url(),
		Title: "Loading...",
	}
	return Source{
		ID:       uuid.New(),
		Meta:     meta,
		provider: prov,
		Status:   StatusNew,
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
