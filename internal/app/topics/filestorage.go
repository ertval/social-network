package topics

import "context"

type FileStorageManager interface {
	Upload(ctx context.Context, file []byte, path string) error
	Delete(ctx context.Context, path string) error
}
