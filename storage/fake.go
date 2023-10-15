package storage

import "context"

type FakeStorage struct{}

func (m *FakeStorage) StoreFile(_ context.Context, _ []byte) (string, error) {
	return "gs://voice-memo-files-sgzmd/myfile", nil
}

func (m *FakeStorage) ClearFile(_ context.Context, _ string) error {
	return nil
}
