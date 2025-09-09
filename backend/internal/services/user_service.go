package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"real-time-chat/internal/config"
	"real-time-chat/internal/domain"
	"real-time-chat/internal/usecase"
	"real-time-chat/internal/utils"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrEmailExists           = errors.New("email already exists")
	ErrUsernameExists        = errors.New("username already exists")
	ErrTokenInvalid          = errors.New("token is invalid")
	ErrEmailNotVerified      = errors.New("email not verified")
	ErrOTPInvalidOrExpired   = errors.New("invalid or expired OTP")
	ErrPasswordTooShort      = errors.New("password must be at least 8 characters long")
	ErrAlreadyVerified       = errors.New("email already verified")
	ErrUserAlreadyExists     = errors.New("user with this email or username already exists")
	ErrProfilePictureInvalid = errors.New("invalid profile picture URL")
)

type UserService struct {
	userService
}

type userService struct {
	userRepo     domain.UserRepository
	TokenService usecase.TokenUseCase
	emailSender  utils.EmailSender
}

func NewUserService(userRepo domain.UserRepository, tokenService usecase.TokenUseCase, emailSender utils.EmailSender) usecase.UserUseCase {
	return &userService{
		userRepo:     userRepo,
		TokenService: tokenService,
		emailSender:  emailSender,
	}
}

func (s *userService) Register(ctx context.Context, email, username, password string) (*domain.User, error) {
	if len(password) < 8 {
		return nil, ErrPasswordTooShort
	}

	if _, err := s.userRepo.FindByEmail(ctx, email); err == nil {
		return nil, ErrEmailExists
	}
	if _, err := s.userRepo.FindByName(ctx, username); err == nil {
		return nil, ErrUsernameExists
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:           uuid.NewString(),
		Email:        email,
		Username:     username,
		PasswordHash: hashedPassword,
		IsVerified:   false, // Set to false initially
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Send email verification
	verificationToken, err := s.TokenService.GenerateEmailVerificationToken(user.ID, user.Email)
	if err != nil {
		log.Printf("Failed to generate email verification token for %s: %v", user.Email, err)
		// Don't fail registration, but log the error. User can request new verification later.
	} else {
		verificationLink := fmt.Sprintf("http://localhost:5173/verify-email?token=%s", verificationToken) // Frontend URL
		err = s.emailSender.SendEmail(user.Email, "Verify Your Email", fmt.Sprintf("Please verify your email by clicking on this link: %s", verificationLink))
		if err != nil {
			log.Printf("Failed to send verification email to %s: %v", user.Email, err)
		}
	}

	return user, nil
}

func (s *userService) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

func (s *userService) VerifyEmail(ctx context.Context, token string) error {
	claims, err := utils.ValidateEmailVerificationToken(token, s.TokenService.(*tokenService).jwtSecret) // Use specific validation and secret
	if err != nil {
		return ErrTokenInvalid
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return ErrUserNotFound
	}
	if user.IsVerified {
		return ErrAlreadyVerified
	}

	return s.userRepo.UpdateVerificationStatus(ctx, user.ID, true)
}

func (s *userService) RequestPasswordReset(ctx context.Context, email string) error {
	_, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return ErrUserNotFound
	}

	// Generate and store OTP
	otp := utils.GenerateOTP(6) // 6-digit OTP
	err = s.TokenService.StoreOTP(ctx, email, otp)
	if err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}

	// Send OTP via email
	err = s.emailSender.SendEmail(email, "Password Reset OTP", fmt.Sprintf("Your password reset OTP is: %s. It is valid for 30 minutes.", otp))
	if err != nil {
		return fmt.Errorf("failed to send OTP email: %w", err)
	}
	return nil
}

func (s *userService) ResetPassword(ctx context.Context, email, otp, newPassword string) error {
	if len(newPassword) < 8 {
		return ErrPasswordTooShort
	}

	storedOTP, err := s.TokenService.GetOTP(ctx, email)
	if err != nil {
		return ErrOTPInvalidOrExpired
	}
	if storedOTP != otp {
		return ErrOTPInvalidOrExpired
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return ErrUserNotFound
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	if err := s.userRepo.UpdatePassword(ctx, user.ID, hashedPassword); err != nil {
		return err
	}

	// Delete OTP after successful reset
	s.TokenService.DeleteOTP(ctx, email)
	return nil
}

func (s *userService) UploadProfilePicture(ctx context.Context, userID, filename string) (string, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", ErrUserNotFound
	}

	// Construct public URL
	publicURL := "/uploads/" + filename // This assumes /uploads is served statically

	if err := s.userRepo.UpdateProfilePicture(ctx, user.ID, publicURL); err != nil {
		return "", err
	}
	return publicURL, nil
}

// tokenService implements TokenUseCase for JWT and OTP management
type tokenService struct {
	userRepo        domain.UserRepository // Added to resolve user for claims during token creation
	tokenRepo       domain.TokenRepository
	jwtSecret       string
	accessDuration  time.Duration
	refreshDuration time.Duration
	otpExpiry       time.Duration // Used for storing OTP
}

func NewTokenService(userRepo domain.UserRepository, tokenRepo domain.TokenRepository, jwtSecret string, accessDuration, refreshDuration, otpExpiry time.Duration) usecase.TokenUseCase {
	return &tokenService{
		userRepo:        userRepo,
		tokenRepo:       tokenRepo,
		jwtSecret:       jwtSecret,
		accessDuration:  accessDuration,
		refreshDuration: refreshDuration,
		otpExpiry:       otpExpiry,
	}
}

func (s *tokenService) Login(ctx context.Context, email, password string) (*config.TokenDetails, *domain.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	// Check if user is verified
	if !user.IsVerified {
		return nil, nil, ErrEmailNotVerified
	}

	if err := utils.CheckPasswordHash(password, user.PasswordHash); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}
	td, err := s.CreateTokens(ctx, user)
	if err != nil {
		return nil, nil, err
	}
	return td, user, nil
}

func (s *tokenService) RefreshToken(ctx context.Context, refreshToken string) (*config.TokenDetails, error) {
	claims, err := utils.ValidateJWT(refreshToken, s.jwtSecret)
	if err != nil {
		return nil, ErrTokenInvalid
	}
	refreshUUID, ok := claims["refresh_uuid"].(string)
	if !ok {
		return nil, ErrTokenInvalid
	}
	userID, err := s.tokenRepo.GetRefreshTokenUserID(ctx, refreshUUID)
	if err != nil {
		return nil, ErrTokenInvalid
	}
	if err := s.tokenRepo.DeleteRefreshToken(ctx, refreshUUID); err != nil {
		log.Printf("could not delete old refresh token: %v\n", err)
	}
	// Need to fetch user to regenerate token with username
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound // Or appropriate error
	}
	td, err := s.CreateTokens(ctx, user)
	if err != nil {
		return nil, err
	}
	return td, nil
}

func (s *tokenService) CreateTokens(ctx context.Context, user *domain.User) (*config.TokenDetails, error) {
	td := &config.TokenDetails{
		AccessTokenExpires:  time.Now().Add(s.accessDuration),
		RefreshTokenExpires: time.Now().Add(s.refreshDuration),
		AccessUUID:          uuid.NewString(),
		RefreshUUID:         uuid.NewString(),
	}

	var err error
	td.AccessToken, err = utils.GenerateJWT(user.ID, user.Username, td.AccessUUID, s.jwtSecret, td.AccessTokenExpires)
	if err != nil {
		return nil, err
	}
	td.RefreshToken, err = utils.GenerateJWT(user.ID, user.Username, td.RefreshUUID, s.jwtSecret, td.RefreshTokenExpires)
	if err != nil {
		return nil, err
	}
	err = s.tokenRepo.StoreRefreshToken(ctx, user.ID, td.RefreshUUID, s.refreshDuration)
	if err != nil {
		return nil, err
	}
	return td, nil
}

func (s *tokenService) ValidateToken(tokenString string) (*usecase.AuthClaims, error) {
	claims, err := utils.ValidateJWT(tokenString, s.jwtSecret)
	if err != nil {
		return nil, err
	}
	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("invalid user_id in token")
	}
	username, ok := claims["username"].(string)
	if !ok {
		return nil, errors.New("invalid username in token")
	}
	return &usecase.AuthClaims{UserID: userID, Username: username}, nil
}

func (s *tokenService) GenerateEmailVerificationToken(userID, email string) (string, error) {
	// Use JWT for email verification token, with a shorter expiry
	expires := time.Now().Add(time.Hour * 24) // 24 hours for email verification
	claims := utils.NewVerificationClaims(userID, email, expires)
	return utils.GenerateJWTWithClaims(claims, s.jwtSecret)
}

func (s *tokenService) StoreOTP(ctx context.Context, email, otp string) error {
	return s.tokenRepo.StoreOTP(ctx, email, otp, s.otpExpiry)
}

func (s *tokenService) GetOTP(ctx context.Context, email string) (string, error) {
	otp, err := s.tokenRepo.GetOTP(ctx, email)
	if err != nil {
		return "", err
	}
	if otp == "" {
		return "", ErrOTPInvalidOrExpired
	}
	return otp, nil
}

func (s *tokenService) DeleteOTP(ctx context.Context, email string) error {
	return s.tokenRepo.DeleteOTP(ctx, email)
}
