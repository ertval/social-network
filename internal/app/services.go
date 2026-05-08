package app

import (
	activityQueries "github.com/arnald/forum/internal/app/activities/queries"
	categoryCommands "github.com/arnald/forum/internal/app/categories/commands"
	categoryQueries "github.com/arnald/forum/internal/app/categories/queries"
	chatapp "github.com/arnald/forum/internal/app/chat"
	chatcommands "github.com/arnald/forum/internal/app/chat/commands"
	chatqueries "github.com/arnald/forum/internal/app/chat/queries"
	commentCommands "github.com/arnald/forum/internal/app/comments/commands"
	commentQueries "github.com/arnald/forum/internal/app/comments/queries"
	"github.com/arnald/forum/internal/app/notifications"
	notificationcommands "github.com/arnald/forum/internal/app/notifications/commands"
	notificationqueries "github.com/arnald/forum/internal/app/notifications/queries"
	oauthservice "github.com/arnald/forum/internal/app/oauth"
	"github.com/arnald/forum/internal/app/topics"
	topicCommands "github.com/arnald/forum/internal/app/topics/commands"
	topicQueries "github.com/arnald/forum/internal/app/topics/queries"
	userCommands "github.com/arnald/forum/internal/app/user/commands"
	userQueries "github.com/arnald/forum/internal/app/user/queries"
	votecommands "github.com/arnald/forum/internal/app/votes/commands"
	voteQueries "github.com/arnald/forum/internal/app/votes/queries"
	"github.com/arnald/forum/internal/domain/activity"
	"github.com/arnald/forum/internal/domain/category"
	"github.com/arnald/forum/internal/domain/chat"
	"github.com/arnald/forum/internal/domain/comment"
	"github.com/arnald/forum/internal/domain/notification"
	"github.com/arnald/forum/internal/domain/oauth"
	"github.com/arnald/forum/internal/domain/topic"
	"github.com/arnald/forum/internal/domain/user"
	"github.com/arnald/forum/internal/domain/vote"
	"github.com/arnald/forum/internal/pkg/bcrypt"
	"github.com/arnald/forum/internal/pkg/uuid"
)

type Queries struct {
	UserLoginGithub    oauthservice.OAuthService
	GetTopic           topicQueries.GetTopicRequestHandler
	GetAllTopics       topicQueries.GetAllTopicsRequestHandler
	GetComment         commentQueries.GetCommentRequestHandler
	GetCommentsByTopic commentQueries.GetCommentsByTopicRequestHandler
	UserLoginEmail     userQueries.UserLoginEmailRequestHandler
	UserLoginUsername  userQueries.UserLoginUsernameRequestHandler
	GetCategoryByID    categoryQueries.GetCategoryByIDHandler
	GetAllCategories   categoryQueries.GetAllCategoriesRequestHandler
	GetCounts          voteQueries.GetCountsRequestHandler
	GetUserActivity    activityQueries.GetUserActivityHandler
	GetAllUsers        userQueries.GetAllUsersRequestHandler
	GetNotifications   notificationqueries.GetNotificationsHandler
	GetUnreadCount     notificationqueries.GetUnreadCountHandler
	GetChatHistory     chatqueries.GetChatHistoryHandler
	GetChatUsers       chatqueries.GetChatUsersHandler
}

type Commands struct {
	UserRegister          userCommands.UserRegisterRequestHandler
	CreateTopic           topicCommands.CreateTopicRequestHandler
	UpdateTopic           topicCommands.UpdateTopicRequestHandler
	DeleteTopic           topicCommands.DeleteTopicRequestHandler
	CreateComment         commentCommands.CreateCommentRequestHandler
	UpdateComment         commentCommands.UpdateCommentRequestHandler
	DeleteComment         commentCommands.DeleteCommentRequestHandler
	CreateCategory        categoryCommands.CreateCategoryRequestHandler
	UpdateCategory        categoryCommands.UpdateCategoryRequestHandler
	DeleteCategory        categoryCommands.DeleteCategoryRequestHandler
	CastVote              votecommands.CastVoteRequestHandler
	DeleteVote            votecommands.DeleteVoteRequestHandler
	CreateNotification    notificationcommands.CreateNotificationHandler
	OpenStream            notificationcommands.OpenStreamHandler
	MarkAsRead            notificationcommands.MarkAsReadHandler
	MarkAllAsRead         notificationcommands.MarkAllAsReadHandler
	InitChat              chatcommands.InitChatHandler
	MarkAsReadChatMessage chatcommands.MarkAsReadHandler
	SendChatMessage       chatcommands.SendChatHandler
}

type Services struct {
	Queries  Queries
	Commands Commands
}

func NewServices(userRepo user.Repository, categoryRepo category.Repository, topicRepo topic.Repository, commentRepo comment.Repository, voteRepo vote.Repository, oauthRepo oauth.Repository, activityRepo activity.Repository, chatRepo chat.Repository, notificationsRepo notification.Repository, notifier notifications.Notifier, broadcaster chatapp.Broadcaster, fileStorage topics.FileStorageManager) Services {
	uuidProvider := uuid.NewProvider()
	encryption := bcrypt.NewProvider()
	return Services{
		Queries: Queries{
			*oauthservice.NewOAuthService(oauthRepo, uuidProvider),
			topicQueries.NewGetTopicHandler(topicRepo, commentRepo),
			topicQueries.NewGetAllTopicsHandler(topicRepo, categoryRepo),
			commentQueries.NewGetCommentHandler(commentRepo),
			commentQueries.NewGetCommentsByTopicRequestHandler(commentRepo),
			userQueries.NewUserLoginEmailHandler(userRepo, encryption),
			userQueries.NewUserLoginUsernameHandler(userRepo, encryption),
			categoryQueries.NewGetCategoryByIDHandler(categoryRepo),
			categoryQueries.NewGetAllCategoriesHandler(categoryRepo),
			voteQueries.NewGetCountsRequestHandler(voteRepo),
			activityQueries.NewGetUserActivityHandler(activityRepo),
			userQueries.NewGetAllUsersRequestHandler(userRepo),
			notificationqueries.NewGetNotificationsHandler(notificationsRepo),
			notificationqueries.NewGetUnreadCountHandler(notificationsRepo),
			chatqueries.NewGetChatHistoryHandler(chatRepo),
			chatqueries.NewGetChatUsersHandler(chatRepo, userRepo, broadcaster),
		},
		Commands: Commands{
			userCommands.NewUserRegisterHandler(userRepo, uuidProvider, encryption),
			topicCommands.NewCreateTopicHandler(topicRepo, fileStorage),
			topicCommands.NewUpdateTopicHandler(topicRepo, fileStorage),
			topicCommands.NewDeleteTopicHandler(topicRepo, fileStorage),
			commentCommands.NewCreateCommentRequestHandler(commentRepo, topicRepo, notificationsRepo, notifier),
			commentCommands.NewUpdateCommentRequestHandler(commentRepo),
			commentCommands.NewDeleteCommentHandler(commentRepo),
			categoryCommands.NewCreateCategoryHandler(categoryRepo),
			categoryCommands.NewUpdateCategoryHandler(categoryRepo),
			categoryCommands.NewDeleteCategoryHandler(categoryRepo),
			votecommands.NewCastVoteHandler(voteRepo, topicRepo, commentRepo, notificationsRepo, notifier),
			votecommands.NewDeleteVoteHandler(voteRepo),
			notificationcommands.NewCreateNotificationHandler(notificationsRepo, notifier),
			notificationcommands.NewOpenStreamHandler(notificationsRepo, notifier),
			notificationcommands.NewMarkAsReadHandler(notificationsRepo),
			notificationcommands.NewMarkAllAsReadHandler(notificationsRepo),
			chatcommands.NewInitChatHandler(chatRepo),
			chatcommands.NewMarkAsReadHandler(chatRepo),
			chatcommands.NewSendChatHandler(chatRepo, broadcaster),
		},
	}
}
