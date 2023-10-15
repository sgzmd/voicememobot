package storage

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/google/uuid"
	"log"
	"time"
	"voicesummary/config"
)

type RealStorage struct {
	client *storage.Client
	config *config.Config
}

func (r *RealStorage) getClient(_ context.Context) *storage.Client {
	return r.client
}

func (r *RealStorage) StoreFile(ctx context.Context, data []byte) (string, error) {
	client := r.getClient(ctx)
	bucketObj := client.Bucket(r.config.GoogleStorageBucket)

	// Create file name with format rec-<timestamp>-<uuid>.wav
	objName := fmt.Sprintf("rec-%s-%s.wav", time.Now().Format("20060102-150405.999Z"), uuid.NewString())
	obj := bucketObj.Object(objName)
	writer := obj.NewWriter(ctx)
	defer writer.Close()

	_, err := writer.Write(data)
	if err != nil {
		log.Printf("Failed to write file: %+v", err)
		return "", err
	}

	return fmt.Sprintf("gs://%s/%s", r.config.GoogleStorageBucket, objName), nil
}

func (r *RealStorage) ClearFile(ctx context.Context, name string) error {
	return r.getClient(ctx).Bucket(r.config.GoogleStorageBucket).Object(name).Delete(ctx)
}
