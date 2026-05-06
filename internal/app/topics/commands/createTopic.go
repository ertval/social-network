package topiccommands

import (
	"context"

	"github.com/arnald/forum/internal/app/topics"
	"github.com/arnald/forum/internal/domain/topic"
	"github.com/arnald/forum/internal/domain/user"
)

type CreateTopicRequest struct {
	User        *user.User
	Title       string `json:"title"`
	Content     string `json:"content"`
	ImagePath   string `json:"imagePath"`
	CategoryIDs []int  `json:"categoryIds"`
}

type CreateTopicRequestHandler interface {
	Handle(ctx context.Context, req CreateTopicRequest) (*topic.Topic, error)
}

type createTopicRequestHandler struct {
	repo        topic.Repository
	fileStorage topics.FileStorageManager
}

func NewCreateTopicHandler(repo topic.Repository, fileStorage topics.FileStorageManager) CreateTopicRequestHandler {
	return &createTopicRequestHandler{
		repo:        repo,
		fileStorage: fileStorage,
	}
}

func (h *createTopicRequestHandler) Handle(ctx context.Context, req CreateTopicRequest) (*topic.Topic, error) {
	topic := &topic.Topic{
		UserID:      req.User.ID,
		CategoryIDs: req.CategoryIDs,
		Title:       req.Title,
		Content:     req.Content,
		ImagePath:   req.ImagePath,
	}

	//h.fileStorage.Upload(req.ImageFile,req.ImageFileName)
	err := h.repo.CreateTopic(ctx, topic)
	if err != nil {
		return nil, err
	}
	return topic, nil
}
