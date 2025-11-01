package downloader

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"tubefeed/internal/models"
	"tubefeed/internal/provider"
	"tubefeed/internal/utils"

	_ "tubefeed/internal/provider/ytdlp" // load ytdlp provider

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

type ErrDownloader error

func (vm *Source) Download(ctx context.Context, path string, progress chan<- int) error {
	if vm.provider == nil {
		domain, err := utils.ExtractDomain(vm.URL)
		if err != nil {
			vm.logger.ErrorContext(ctx, "failed to extract domain")
			return ErrDownloader(err)
		}
		newprovider := provider.Get(domain)
		if newprovider == nil {
			vm.logger.ErrorContext(ctx, "no provider found")
			return ErrDownloader(errors.New("no provider found"))
		}
		provider, err := newprovider(vm.URL, vm.logger)
		if err != nil {
			vm.logger.ErrorContext(ctx, fmt.Sprintf("provider failed: %v", err))
			return ErrDownloader(err)
		}
		vm.provider = provider
	}

	_, err := os.Stat(path)
	if err != nil {
		vm.logger.ErrorContext(ctx, fmt.Sprintf("stat failed: %v", err))
		return ErrDownloader(err)
	}
	return vm.provider.Download(ctx, vm.ID, path, progress)
}

// NewSource initializes a new downloader based on the URL domain
func NewSource(id uuid.UUID, url string, logger *slog.Logger) (Source, error) {
	domain, err := utils.ExtractDomain(url)
	if err != nil {
		return Source{}, err
	}
	newprovider := provider.Get(domain)
	if newprovider == nil {
		return Source{}, fmt.Errorf("domain not supported: %s", domain)
	}
	provider, err := newprovider(url, logger)
	if err != nil {
		return Source{}, err
	}
	return Source{
		ID:       id,
		URL:      url,
		provider: provider,
		logger:   logger.With("id", id),
	}, nil
}

// LoadMeta loads metadata for the given request
func (vm *Source) LoadMeta(ctx context.Context, request *models.Request) error {

	metadata, err := vm.provider.LoadMetadata(ctx)
	if err != nil {
		return err
	}
	request.Title = metadata.Title
	request.Status = models.StatusMeta
	request.SourceURL = metadata.URL
	return nil
}
