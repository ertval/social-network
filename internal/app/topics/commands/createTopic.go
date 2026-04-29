package topiccommands

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/arnald/forum/internal/domain/topic"
	"github.com/arnald/forum/internal/domain/user"
)

const (
	savingDir       = "frontend/static/images/uploads"
	imagepathprefix = "/static/images/uploads"
	uploadDirPerm   = 0o750
)

type CreateTopicRequest struct {
	User        *user.User
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	ImagePath   string     `json:"imagePath"`
	ImageFile   TopicImage `json:"topicImage"`
	CategoryIDs []int      `json:"categoryIds"`
}
type TopicImage struct {
	File   *multipart.File
	Header *multipart.FileHeader
}

type CreateTopicRequestHandler interface {
	Handle(ctx context.Context, req CreateTopicRequest) (*topic.Topic, error)
}

type createTopicRequestHandler struct {
	repo topic.Repository
}

func NewCreateTopicHandler(repo topic.Repository) CreateTopicRequestHandler {
	return &createTopicRequestHandler{
		repo: repo,
	}
}

func (h *createTopicRequestHandler) Handle(ctx context.Context, req CreateTopicRequest) (*topic.Topic, error) {

	destPath := filepath.Join(savingDir, req.ImagePath)
	destPath = filepath.Clean(destPath)

	if !strings.HasPrefix(destPath, filepath.Clean(savingDir)+string(os.PathSeparator)) {
		return nil, errors.New("Invalid File Path")
	}

	var destFile *os.File
	destFile, err := os.Create(destPath)
	if err != nil {
		return nil, err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, *req.ImageFile.File)
	if err != nil {
		return nil, err
	}
	req.ImagePath = filepath.Join(imagepathprefix, req.ImagePath)

	topic := &topic.Topic{
		UserID:      req.User.ID,
		CategoryIDs: req.CategoryIDs,
		Title:       req.Title,
		Content:     req.Content,
		ImagePath:   req.ImagePath,
	}

	err = h.repo.CreateTopic(ctx, topic)
	if err != nil {
		return nil, err
	}
	return topic, nil
}
