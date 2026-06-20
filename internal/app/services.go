package app

import (
	activityQueries "social-network/internal/app/activities/queries"
	categoryCommands "social-network/internal/app/categories/commands"
	categoryQueries "social-network/internal/app/categories/queries"
	chatapp "social-network/internal/app/chat"
	chatcommands "social-network/internal/app/chat/commands"
	chatqueries "social-network/internal/app/chat/queries"
	commentCommands "social-network/internal/app/comments/commands"
	commentQueries "social-network/internal/app/comments/queries"
	"social-network/internal/app/notifications"
	notificationcommands "social-network/internal/app/notifications/commands"
	notificationqueries "social-network/internal/app/notifications/queries"
	oauthservice "social-network/internal/app/oauth"
	"social-network/internal/app/topics"
	topicCommands "social-network/internal/app/topics/commands"
	topicQueries "social-network/internal/app/topics/queries"
	userCommands "social-network/internal/app/user/commands"
	userQueries "social-network/internal/app/user/queries"
	votecommands "social-network/internal/app/votes/commands"
	voteQueries "social-network/internal/app/votes/queries"
	"social-network/internal/domain/activity"
	"social-network/internal/domain/category"
	"social-network/internal/domain/chat"
	"social-network/internal/domain/comment"
	"social-network/internal/domain/notification"
	"social-network/internal/domain/oauth"
	"social-network/internal/domain/topic"
	"social-network/internal/domain/user"
	"social-network/internal/domain/vote"
	"social-network/internal/pkg/bcrypt"
	"social-network/internal/pkg/uuid"
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
