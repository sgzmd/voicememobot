package storage

import (
	"cloud.google.com/go/storage"
	"context"
	"voicesummary/config"
)

type Storage interface {
	StoreFile(ctx context.Context, data []byte) (string, error)
	ClearFile(ctx context.Context, name string) error
}

func NewRealStorage(ctx context.Context, conf *config.Config) (*RealStorage, error) {
	// Creating client instance ahead of time to detect issues early
	client, err := storage.NewClient(ctx, conf.GetCredentialsOption())
	if err != nil {
		return nil, err
	}

	rs := &RealStorage{
		config: conf,
		client: client,
	}

	return rs, nil
}

func NewFakeStorage() *FakeStorage {
	return &FakeStorage{}
}
