package topiccommands

import (
	"context"

	"github.com/arnald/forum/internal/app/topics"
	"github.com/arnald/forum/internal/domain/topic"
	"github.com/arnald/forum/internal/domain/user"
)

type DeleteTopicRequest struct {
	User    *user.User
	TopicID int `json:"topicId"`
}

type DeleteTopicRequestHandler interface {
	Handle(ctx context.Context, req DeleteTopicRequest) error
}

type deleteTopicRequestHandler struct {
	repo        topic.Repository
	fileStorage topics.FileStorageManager
}

func NewDeleteTopicHandler(repo topic.Repository, fileStorage topics.FileStorageManager) DeleteTopicRequestHandler {
	return &deleteTopicRequestHandler{
		repo:        repo,
		fileStorage: fileStorage,
	}
}

func (h *deleteTopicRequestHandler) Handle(ctx context.Context, req DeleteTopicRequest) error {
	err := h.repo.DeleteTopic(ctx, req.User.ID, req.TopicID)
	if err != nil {
		return err
	}
	//h.fileStorage.Delete(req.ImagePath)
	return nil
}
