package meta

import (
	"fmt"
	"tubefeed/internal/provider"
	"tubefeed/internal/provider/registry"
	"tubefeed/internal/utils"

	"github.com/google/uuid"
)

type Video struct {
	provider provider.VideoProvider
	Status   string
	Meta     provider.VideoMeta
}

type VideoProviderList map[string]provider.ProviderNewVideoFn

type Provider struct {
	List VideoProviderList
}

func (vm *Video) Download(path string) error {
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
	return vm.provider.Download(path)
}

func NewVideo(url string) (Video, error) {
	domain, _ := utils.ExtractDomain(url)
	new := registry.Get(domain)
	provider, err := new(url)
	if err != nil {
		return Video{}, err
	}
	meta, err := provider.LoadMetadata()
	meta.ID = uuid.New()
	if err != nil {
		return Video{}, err
	}
	return Video{
		Meta:     *meta,
		provider: provider,
	}, nil
}

func (vm *Video) LoadMeta() error {
	videomd, err := vm.provider.LoadMetadata()
	if err != nil {
		return err
	}
	vm.Meta = *videomd
	return nil
}
