package createcomment

import (
	"context"
	commentCommands "social-network/internal/app/comments/commands"
	"social-network/internal/config"
	"social-network/internal/infra/logger"
	"social-network/internal/infra/middleware"
	"social-network/internal/pkg/helpers"
	"social-network/internal/pkg/validator"
	"net/http"
	"strconv"
)

type RequestModel struct {
	Content string `json:"content"`
	TopicID int    `json:"topicId"`
}

type ResponseModel struct {
	Message   string `json:"message"`
	CommentID int    `json:"commentId"`
}

type Handler struct {
	createComment commentCommands.CreateCommentRequestHandler
	Config        *config.ServerConfig
	Logger        logger.Logger
}

func NewHandler(createComment commentCommands.CreateCommentRequestHandler, config *config.ServerConfig, logger logger.Logger) *Handler {
	return &Handler{
		createComment: createComment,
		Config:        config,
		Logger:        logger,
	}
}

func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.Logger.PrintError(logger.ErrInvalidRequestMethod, nil)
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	user := middleware.GetUserFromContext(r)
	if user == nil {
		h.Logger.PrintError(logger.ErrUserNotFoundInContext, nil)
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.Config.Timeouts.HandlerTimeouts.UserRegister)
	defer cancel()

	var commentToCreate RequestModel

	commentAny, err := helpers.ParseBodyRequest(r, &commentToCreate)
	if err != nil {
		helpers.RespondWithError(w,
			http.StatusBadRequest,
			"Invalid request payload",
		)

		h.Logger.PrintError(err, nil)
		return
	}
	defer r.Body.Close()

	v := validator.New()

	validator.ValidateCreateComment(v, commentAny)

	if !v.Valid() {
		helpers.RespondWithError(
			w,
			http.StatusBadRequest,
			v.ToStringErrors(),
		)

		h.Logger.PrintError(logger.ErrValidationFailed, v.Errors)
		return
	}

	comment, err := h.createComment.Handle(ctx, commentCommands.CreateCommentRequest{
		TopicID: commentToCreate.TopicID,
		Content: commentToCreate.Content,
		User:    user,
	})
	if err != nil {
		helpers.RespondWithError(w,
			http.StatusInternalServerError,
			"Failed to create comment",
		)

		h.Logger.PrintError(err, nil)
		return
	}
	commentResponse := ResponseModel{
		CommentID: comment.ID,
		Message:   "Comment created successfully",
	}

	helpers.RespondWithJSON(
		w,
		http.StatusCreated,
		nil,
		commentResponse,
	)

	h.Logger.PrintInfo(
		"Comment created successfully",
		map[string]string{
			"user_id":    user.ID,
			"comment_id": strconv.Itoa(comment.ID),
		},
	)
}
