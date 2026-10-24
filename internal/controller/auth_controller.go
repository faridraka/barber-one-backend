package controller

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"barber-one/backend/internal/dto"
	"barber-one/backend/internal/middleware"
	"barber-one/backend/internal/service"
	"golang.org/x/crypto/bcrypt"
)

// AuthController handles authentication requests
type AuthController struct {
	userService  *service.UserService
	sessionStore *middleware.SessionStore
}

func NewAuthController(userService *service.UserService, sessionStore *middleware.SessionStore) *AuthController {
	return &AuthController{userService: userService, sessionStore: sessionStore}
}

// Register handles POST /register
// @Summary      Register user baru
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RegisterRequest  true  "Data registrasi"
// @Success      201   {object}  dto.RegisteredUser
// @Failure      400   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Router       /register [post]
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Metode tidak diizinkan."})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 8*1024)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var input dto.RegisterRequest
	if err := decoder.Decode(&input); err != nil {
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Format data tidak valid."})
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Permintaan harus berisi satu objek JSON."})
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Phone = strings.TrimSpace(input.Phone)
	if input.Email != nil {
		cleanEmail := strings.TrimSpace(*input.Email)
		input.Email = &cleanEmail
		if cleanEmail == "" {
			input.Email = nil
		}
	}

	if input.Name == "" || len([]rune(input.Name)) > 100 {
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Nama wajib diisi dan maksimal 100 karakter."})
		return
	}
	if input.Phone == "" || len(input.Phone) > 20 {
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Nomor telepon wajib diisi dan maksimal 20 karakter."})
		return
	}
	if input.Email != nil {
		if len(*input.Email) > 100 {
			middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Email maksimal 100 karakter."})
			return
		}
		// TODO: Validate email format
	}
	if len(input.Password) < 8 || len(input.Password) > 72 {
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Kata sandi harus terdiri dari 8 sampai 72 byte."})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		middleware.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Gagal memproses pendaftaran."})
		return
	}

	userID, err := c.userService.Register(r.Context(), input.Name, input.Phone, input.Email, string(hashedPassword))
	if err != nil {
		middleware.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Pendaftaran gagal."})
		return
	}

	middleware.WriteJSON(w, http.StatusCreated, map[string]any{
		"message": "Pendaftaran berhasil.",
		"user": map[string]any{
			"id":    userID,
			"name":  input.Name,
			"phone": input.Phone,
			"email": input.Email,
			"role":  "user",
		},
	})
}

// Login handles POST /login
// @Summary      Login user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.LoginRequest  true  "Kredensial"
// @Success      200   {object}  map[string]any
// @Failure      401   {object}  map[string]string
// @Router       /login [post]
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Metode tidak diizinkan."})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 8*1024)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var input dto.LoginRequest
	if err := decoder.Decode(&input); err != nil {
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Format data tidak valid."})
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Permintaan harus berisi satu objek JSON."})
		return
	}

	input.Identifier = strings.TrimSpace(input.Identifier)
	if input.Identifier == "" || input.Password == "" {
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Nomor telepon/email dan kata sandi wajib diisi."})
		return
	}

	user, err := c.userService.Login(r.Context(), input.Identifier, input.Password)
	if err != nil {
		middleware.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "Nomor telepon/email atau kata sandi tidak sesuai."})
		return
	}

	// Create session
	token, err := c.sessionStore.Create(user["id"].(int64))
	if err != nil {
		middleware.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Gagal membuat sesi login."})
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]any{
		"message":      "Login berhasil.",
		"user": map[string]any{
			"id":    user["id"],
			"name":  user["name"],
			"phone": user["phone"],
			"email": user["email"],
			"role":  user["role"],
		},
		"access_token": token,
	})
}

// Profile handles GET /profile
// @Summary      Ambil profil user yang sedang login
// @Tags         user
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]any
// @Failure      401  {object}  map[string]string
// @Router       /profile [get]
func (c *AuthController) Profile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Metode tidak diizinkan."})
		return
	}

	userID, ok := c.sessionStore.UserIDFromRequest(r)
	if !ok {
		middleware.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid atau sudah kedaluwarsa. Silakan login kembali."})
		return
	}

	user, err := c.userService.GetProfile(r.Context(), userID)
	if err != nil {
		middleware.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Gagal mengambil profil user."})
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]any{"user": user})
}

// Logout handles POST /logout
// @Summary      Logout dan hapus sesi
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]string
// @Router       /logout [post]
func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Metode tidak diizinkan."})
		return
	}

	c.sessionStore.RevokeFromRequest(r)
	w.WriteHeader(http.StatusNoContent)
}
