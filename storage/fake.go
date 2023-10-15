package storage

import "context"

type FakeStorage struct{}

func (m *FakeStorage) StoreFile(_ context.Context, fileName string, _ []byte) (string, error) {
	return "gs://voice-memo-files-sgzmd/" + fileName, nil
}

func (m *FakeStorage) ClearFile(_ context.Context, _, _ string) error {
	return nil
}
