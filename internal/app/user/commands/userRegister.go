package usercommands

import (
	"context"
	"time"

	"social-network/internal/domain/user"
	"social-network/internal/pkg/bcrypt"
	"social-network/internal/pkg/helpers"
	"social-network/internal/pkg/uuid"
)

type UserRegisterRequest struct {
	Nickname  string
	Gender    string
	FirstName string
	LastName  string
	Password  string
	Email     string
	Age       int
}

type UserRegisterRequestHandler interface {
	Handle(ctx context.Context, req UserRegisterRequest) (*user.User, error)
}

type userRegisterRequestHandler struct {
	uuidiProvider      uuid.Provider
	encryptionProvider bcrypt.Provider
	repo               user.Repository
}

func NewUserRegisterHandler(repo user.Repository, uuidProvider uuid.Provider, en bcrypt.Provider) UserRegisterRequestHandler {
	return userRegisterRequestHandler{
		repo:               repo,
		uuidiProvider:      uuidProvider,
		encryptionProvider: en,
	}
}

func (h userRegisterRequestHandler) Handle(ctx context.Context, req UserRegisterRequest) (*user.User, error) {
	user := &user.User{
		CreatedAt: time.Now(),
		Password:  req.Password,
		AvatarURL: nil,
		Nickname:  req.Nickname,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Age:       req.Age,
		Gender:    req.Gender,
		ID:        h.uuidiProvider.NewUUID(),
	}

	err := helpers.ValidateEmail(user.Email)
	if err != nil {
		return nil, err
	}

	encryptedPass, err := h.encryptionProvider.Generate(user.Password)
	if err != nil {
		return nil, err
	}

	user.Password = encryptedPass

	err = h.repo.UserRegister(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, err
}
