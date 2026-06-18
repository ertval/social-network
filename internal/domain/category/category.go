package category

import "social-network/internal/domain/topic"

type Category struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	CreatedAt   string        `json:"createdAt"`
	CreatedBy   string        `json:"createdBy"`
	ImagePath   string        `json:"imagePath"`
	Color       string        `json:"color"`
	Slug        string        `json:"slug"`
	Topics      []topic.Topic `json:"topics"`
	ID          int           `json:"id"`
	TopicCount  int           `json:"topicsCount"`
}
