package castvote

import (
	"context"
	"encoding/json"
	"net/http"
	"social-network/internal/config"
	"social-network/internal/domain/vote"
	"social-network/internal/infra/logger"
	"social-network/internal/infra/middleware"
	"social-network/internal/pkg/helpers"

	votecommands "social-network/internal/app/votes/commands"
)

type RequestModel struct {
	TopicID      *int `json:"topicId"`
	CommentID    *int `json:"commentId,omitempty"`
	ReactionType int  `json:"reactionType"`
}

type ResponseModel struct {
	Counts  *vote.Counts `json:"counts"`
	Message string       `json:"message"`
}

type Handler struct {
	castVote votecommands.CastVoteRequestHandler
	Config   *config.ServerConfig
	Logger   logger.Logger
}

func NewHandler(castvoteHandler votecommands.CastVoteRequestHandler, config *config.ServerConfig, logger logger.Logger) *Handler {
	return &Handler{
		castVote: castvoteHandler,
		Config:   config,
		Logger:   logger,
	}
}

func (h *Handler) CastVote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.Logger.PrintError(logger.ErrInvalidRequestMethod, nil)
		helpers.RespondWithError(
			w,
			http.StatusMethodNotAllowed,
			"Invalid request method",
		)
		return
	}

	user := middleware.GetUserFromContext(r)
	if user == nil {
		h.Logger.PrintError(logger.ErrUserNotFoundInContext, nil)
		helpers.RespondWithError(
			w,
			http.StatusUnauthorized,
			"Unauthorized",
		)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.Config.Timeouts.HandlerTimeouts.UserRegister)
	defer cancel()

	var req RequestModel
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.Logger.PrintError(err, nil)
		helpers.RespondWithError(
			w,
			http.StatusBadRequest,
			"invalid JSON",
		)
		return
	}

	target := vote.Target{
		TopicID:   req.TopicID,
		CommentID: req.CommentID,
	}

	err = h.castVote.Handle(ctx, votecommands.CastVoteRequest{
		UserID:       user.ID,
		NickName:     user.Nickname,
		Target:       target,
		ReactionType: req.ReactionType,
	})
	if err != nil {
		h.Logger.PrintError(err, nil)
		helpers.RespondWithError(
			w,
			http.StatusInternalServerError,
			"failed to cast vote",
		)
		return
	}

	Response := ResponseModel{
		Message: "Vote cast successfully",
	}

	helpers.RespondWithJSON(
		w,
		http.StatusOK,
		nil,
		Response,
	)
}
