package http_delivery

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"real-time-chat/internal/domain"
	"real-time-chat/internal/usecase"
	"real-time-chat/internal/utils"
	"strings"
	"time"

	"github.com/google/uuid"
)

type UserHandler struct {
	userService  usecase.UserUseCase
	tokenService usecase.TokenUseCase
	jwtSecret    string
	uploadDir    string
}

func NewUserHandler(us usecase.UserUseCase, ts usecase.TokenUseCase, jwtSecret, uploadDir string) *UserHandler {
	return &UserHandler{userService: us, tokenService: ts, jwtSecret: jwtSecret, uploadDir: uploadDir}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	user, err := h.userService.Register(r.Context(), req.Email, req.Username, req.Password)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSONResponse(w, http.StatusCreated, user)
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	tokens, user, err := h.tokenService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		ErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}
	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"user":         user,
	})
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *UserHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	tokens, err := h.tokenService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		ErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}
	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
	})
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userContextKey).(*domain.User)
	if !ok {
		ErrorResponse(w, http.StatusInternalServerError, "Could not retrieve user from context")
		return
	}
	JSONResponse(w, http.StatusOK, user)
}

type VerifyEmailRequest struct {
	Token string `json:"token"`
}

func (h *UserHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token") // Can be from query param or body
	if token == "" {
		var req VerifyEmailRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			ErrorResponse(w, http.StatusBadRequest, "Missing token")
			return
		}
		token = req.Token
	}

	if err := h.userService.VerifyEmail(r.Context(), token); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSONResponse(w, http.StatusOK, map[string]string{"message": "Email verified successfully"})
}

type RequestPasswordResetRequest struct {
	Email string `json:"email"`
}

func (h *UserHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req RequestPasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.userService.RequestPasswordReset(r.Context(), req.Email); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSONResponse(w, http.StatusOK, map[string]string{"message": "Password reset OTP sent to email"})
}

type ResetPasswordRequest struct {
	Email       string `json:"email"`
	OTP         string `json:"otp"`
	NewPassword string `json:"new_password"`
}

func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.userService.ResetPassword(r.Context(), req.Email, req.OTP, req.NewPassword); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSONResponse(w, http.StatusOK, map[string]string{"message": "Password reset successfully"})
}

func (h *UserHandler) UploadProfilePicture(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*domain.User)

	err := r.ParseMultipartForm(200 * 1024) // 200 KB limit for profile picture
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("File too large or invalid form: %v", err))
		return
	}

	file, handler, err := r.FormFile("profile_picture")
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Error retrieving the file")
		return
	}
	defer file.Close()

	// Generate a unique filename
	extension := filepath.Ext(handler.Filename)
	filename := uuid.NewString() + extension
	filePath := filepath.Join(h.uploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Error creating file on server")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Error saving file on server")
		return
	}

	profilePictureURL, err := h.userService.UploadProfilePicture(r.Context(), user.ID, filename)
	if err != nil {
		os.Remove(filePath) // Clean up uploaded file if DB update fails
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSONResponse(w, http.StatusOK, map[string]string{"profile_picture_url": profilePictureURL})
}

// FileServer serves static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	r.Get(path+"*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
