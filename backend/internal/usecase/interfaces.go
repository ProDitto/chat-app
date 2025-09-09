package usecase

import (
	"context"
	"real-time-chat/internal/config"
	"real-time-chat/internal/domain"
	"time"
)

// AuthClaims holds user information from a validated JWT
type AuthClaims struct {
	UserID   string
	Username string
}

type UserUseCase interface {
	Register(ctx context.Context, email, username, password string) (*domain.User, error)
	GetUserByID(ctx context.Context, userID string) (*domain.User, error)
	VerifyEmail(ctx context.Context, token string) error
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, email, otp, newPassword string) error
	UploadProfilePicture(ctx context.Context, userID, filename string) (string, error)
}

type TokenUseCase interface {
	CreateTokens(ctx context.Context, user *domain.User) (*config.TokenDetails, error)
	Login(ctx context.Context, email, password string) (*config.TokenDetails, *domain.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (*config.TokenDetails, error)
	ValidateToken(tokenString string) (*AuthClaims, error)
	GenerateEmailVerificationToken(userID, email string) (string, error)
	StoreOTP(ctx context.Context, email, otp string) error
	GetOTP(ctx context.Context, email string) (string, error)
	DeleteOTP(ctx context.Context, email string) error
}

type MessageUseCase interface {
	SaveMessage(ctx context.Context, message *domain.Message) (*domain.Message, error)
	GetMessagesForConversation(ctx context.Context, conversationID, userID string, before time.Time, limit int) ([]*domain.Message, error)
	GetConversationLastMessage(ctx context.Context, conversationID string) (*domain.Message, error)
}

type ConversationUseCase interface {
	GetUserConversations(ctx context.Context, userID string) ([]*domain.Conversation, error)
	GetParticipantIDs(ctx context.Context, conversationID string) ([]string, error)
	MarkConversationAsRead(ctx context.Context, conversationID, userID string) error
	CreateOneToOneConversation(ctx context.Context, userID1, userID2 string) (string, error)
	DeleteOneToOneConversation(ctx context.Context, conversationID, userID string) error
	GetConversationByID(ctx context.Context, conversationID string) (*domain.Conversation, error)
	IsUserInConversation(ctx context.Context, conversationID, userID string) (bool, error)
	CreateGroupConversation(ctx context.Context, groupID string, participantIDs []string) error
	DeleteConversation(ctx context.Context, conversationID string) error
}

type FriendshipUseCase interface {
	SendRequest(ctx context.Context, requesterID, recipientUsername string) error
	RespondToRequest(ctx context.Context, userID, requestID string, status domain.FriendshipStatus) error
	GetPendingRequests(ctx context.Context, userID string) ([]*domain.FriendshipRequest, error)
	GetFriends(ctx context.Context, userID string) ([]*domain.User, error)
}

type GroupUseCase interface {
	CreateGroup(ctx context.Context, ownerID, name, slug string, initialMembers []string) (*domain.Group, error)
	GetGroupDetails(ctx context.Context, groupID string) (*domain.Group, []*domain.User, error)
	JoinGroup(ctx context.Context, groupID, userID string) error
	LeaveGroup(ctx context.Context, groupID, userID string) error
	RemoveGroupMember(ctx context.Context, performingUserID, groupID, memberToRemoveID string) error
	DeleteGroup(ctx context.Context, performingUserID, groupID string) error
}

type GameUseCase interface {
	InviteToTicTacToe(ctx context.Context, player1ID, player2Username string) (*domain.Game, error)
	GetGame(ctx context.Context, gameID string) (*domain.Game, error)
	MakeTicTacToeMove(ctx context.Context, gameID, playerID string, row, col int) (*domain.Game, error)
	RespondToGameInvite(ctx context.Context, gameID, userID string, accept bool) (*domain.Game, error)
}

type EventUseCase interface {
	CreateEvent(ctx context.Context, userID string, eventType domain.EventType, payload interface{}) error
	GetEventsForUser(ctx context.Context, userID string, sinceEventID string, limit int) ([]*domain.Event, error)
}
