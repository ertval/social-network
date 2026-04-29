package topiccommands

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/arnald/forum/internal/domain/topic"
	"github.com/arnald/forum/internal/domain/user"
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

	fmt.Println(req.ImagePath)
	//add saving logic files lives in req.ImageFile.File
	topic := &topic.Topic{
		UserID:      req.User.ID,
		CategoryIDs: req.CategoryIDs,
		Title:       req.Title,
		Content:     req.Content,
		ImagePath:   req.ImagePath,
	}

	err := h.repo.CreateTopic(ctx, topic)
	if err != nil {
		return nil, err
	}
	return topic, nil
}
