package topiccommands

import (
	"context"
)

type mockFileStorage struct {
	UploadFunc func(ctx context.Context, file []byte, path string) error
	DeleteFunc func(ctx context.Context, path string) error
}

func (m *mockFileStorage) Upload(ctx context.Context, file []byte, path string) error {
	if m.UploadFunc != nil {
		return m.UploadFunc(ctx, file, path)
	}
	return nil
}

func (m *mockFileStorage) Delete(ctx context.Context, path string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, path)
	}
	return nil
}
