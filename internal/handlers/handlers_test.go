package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Marst/reminder-app/internal/database"
	"github.com/Marst/reminder-app/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// =============================================================
// TestMain
// =============================================================

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	// Set JWT_SECRET agar GenerateJWT tidak gagal di test Login/Register
	os.Setenv("JWT_SECRET", "test-secret-for-handlers")

	database.ConnectTestDB()

	code := m.Run()

	database.CleanupTestDB()
	os.Exit(code)
}

// =============================================================
// HELPERS
// =============================================================

// counter untuk memastikan setiap email benar-benar unik
// meskipun t.Name() sama (parallel test, dll)
var emailCounter atomic.Int64

// uniqueEmail membuat email unik yang pasti tidak bentrok antar test.
// Format: {namatest}_{timestamp}_{counter}@test.example.com
// Domain "@test.example.com" dipakai CleanupTestDB untuk cleanup.
func uniqueEmail(t *testing.T) string {
	t.Helper()
	n := emailCounter.Add(1)
	// Sanitize nama test: hapus karakter yang tidak valid di email
	name := sanitizeForEmail(t.Name())
	return fmt.Sprintf("%s_%d_%d@test.example.com", name, time.Now().UnixMilli(), n)
}

// sanitizeForEmail mengganti karakter yang tidak valid di local part email.
func sanitizeForEmail(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			result = append(result, c)
		} else {
			result = append(result, '_')
		}
	}
	// Potong maksimal 30 karakter agar tidak terlalu panjang
	if len(result) > 30 {
		result = result[:30]
	}
	return string(result)
}

func setupRouter() *gin.Engine {
	return gin.New()
}

// setupRouterWithAuth mensimulasikan request yang sudah melewati AuthMiddleware.
func setupRouterWithAuth(userID int) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	return r
}

// requireDB men-skip test jika database.DB nil.
func requireDB(t *testing.T) {
	t.Helper()
	if database.DB == nil {
		t.Skip("Skipping: TEST_DATABASE_URL not set or DB connection failed")
	}
}

func jsonBody(t *testing.T, v interface{}) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("jsonBody: %v", err)
	}
	return bytes.NewBuffer(b)
}

func parseResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parseResponse: %v\nbody: %s", err, w.Body.String())
	}
	return resp
}

func newRequest(method, path string, body *bytes.Buffer) *http.Request {
	if body == nil {
		body = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	return req
}

// registerUser adalah helper untuk mendaftarkan user baru di DB test.
// Return email dan password yang dipakai agar bisa digunakan untuk login.
func registerUser(t *testing.T, router *gin.Engine) (email, password string) {
	t.Helper()
	email = uniqueEmail(t)
	password = "password123"

	body := jsonBody(t, map[string]string{
		"name": "Test User", "email": email, "password": password,
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, newRequest(http.MethodPost, "/register", body))

	if w.Code != http.StatusCreated {
		t.Fatalf("registerUser: expected 201, got %d — body: %s", w.Code, w.Body.String())
	}
	return email, password
}

// =============================================================
// TEST: Register
// =============================================================

// TestRegister_InvalidJSON — tidak butuh DB.
func TestRegister_InvalidJSON(t *testing.T) {
	router := setupRouter()
	router.POST("/register", handlers.Register)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, newRequest(http.MethodPost, "/register", bytes.NewBufferString("not{valid")))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, "Failed to parse struct", resp["message"])
}

// TestRegister_EmptyFields — butuh DB.
// Service validasi field kosong dan return error sebelum query INSERT.
func TestRegister_EmptyFields(t *testing.T) {
	requireDB(t)

	router := setupRouter()
	router.POST("/register", handlers.Register)

	body := jsonBody(t, map[string]string{"name": "", "email": "", "password": ""})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, newRequest(http.MethodPost, "/register", body))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, false, resp["success"])
	assert.NotEmpty(t, resp["message"])
}

// TestRegister_Success — butuh DB.
func TestRegister_Success(t *testing.T) {
	requireDB(t)

	router := setupRouter()
	router.POST("/register", handlers.Register)

	body := jsonBody(t, map[string]string{
		"name":     "Test User",
		"email":    uniqueEmail(t), // email unik, pasti belum ada di DB
		"password": "password123",
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, newRequest(http.MethodPost, "/register", body))

	assert.Equal(t, http.StatusCreated, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, "Successfully register new user!", resp["message"])
	assert.NotNil(t, resp["data"])
}

// TestRegister_DuplicateEmail — butuh DB.
func TestRegister_DuplicateEmail(t *testing.T) {
	requireDB(t)

	router := setupRouter()
	router.POST("/register", handlers.Register)

	email := uniqueEmail(t)

	// Register pertama → harus sukses
	w1 := httptest.NewRecorder()
	body1 := jsonBody(t, map[string]string{"name": "Test User", "email": email, "password": "pass123"})
	router.ServeHTTP(w1, newRequest(http.MethodPost, "/register", body1))
	assert.Equal(t, http.StatusCreated, w1.Code, "first register should succeed")

	// Register kedua dengan email sama → harus gagal
	w2 := httptest.NewRecorder()
	body2 := jsonBody(t, map[string]string{"name": "Test User", "email": email, "password": "pass123"})
	router.ServeHTTP(w2, newRequest(http.MethodPost, "/register", body2))

	assert.Equal(t, http.StatusBadRequest, w2.Code)
	resp := parseResponse(t, w2)
	assert.Equal(t, false, resp["success"])
}

// =============================================================
// TEST: Login
// =============================================================

// TestLogin_InvalidJSON — tidak butuh DB.
func TestLogin_InvalidJSON(t *testing.T) {
	router := setupRouter()
	router.POST("/login", handlers.Login)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, newRequest(http.MethodPost, "/login", bytes.NewBufferString("{bad")))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, "Failed to parse struct", resp["message"])
}

// TestLogin_EmptyFields — butuh DB.
func TestLogin_EmptyFields(t *testing.T) {
	requireDB(t)

	router := setupRouter()
	router.POST("/login", handlers.Login)

	body := jsonBody(t, map[string]string{"email": "", "password": ""})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, newRequest(http.MethodPost, "/login", body))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, false, resp["success"])
}

// TestLogin_UserNotFound — butuh DB.
func TestLogin_UserNotFound(t *testing.T) {
	requireDB(t)

	router := setupRouter()
	router.POST("/login", handlers.Login)

	body := jsonBody(t, map[string]string{
		"email":    "notexist_99999@test.example.com",
		"password": "anypassword",
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, newRequest(http.MethodPost, "/login", body))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, false, resp["success"])
}

// TestLogin_WrongPassword — butuh DB.
func TestLogin_WrongPassword(t *testing.T) {
	requireDB(t)

	router := setupRouter()
	router.POST("/register", handlers.Register)
	router.POST("/login", handlers.Login)

	// Register dulu dengan password yang benar
	email, _ := registerUser(t, router)

	// Login dengan password salah
	body := jsonBody(t, map[string]string{"email": email, "password": "WRONG_PASSWORD"})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, newRequest(http.MethodPost, "/login", body))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, false, resp["success"])
}

// TestLogin_Success — butuh DB.
func TestLogin_Success(t *testing.T) {
	requireDB(t)

	router := setupRouter()
	router.POST("/register", handlers.Register)
	router.POST("/login", handlers.Login)

	// Register dulu
	email, password := registerUser(t, router)

	// Login dengan kredensial yang benar
	body := jsonBody(t, map[string]string{"email": email, "password": password})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, newRequest(http.MethodPost, "/login", body))

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, "Successfully login!", resp["message"])
	data, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok, "data should be an object")
	assert.NotEmpty(t, data["token"], "token should not be empty")
}

// =============================================================
// TEST: Profile
// =============================================================

// TestProfile_NoToken — tidak butuh DB.
func TestProfile_NoToken(t *testing.T) {
	router := setupRouter()
	router.GET("/profile", handlers.Profile)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, newRequest(http.MethodGet, "/profile", nil))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, "Token is required", resp["message"])
}

// TestProfile_UserNotFound — butuh DB.
// Setelah fix `return` di auth_handler.go, harus return 500.
func TestProfile_UserNotFound(t *testing.T) {
	requireDB(t)

	router := setupRouterWithAuth(99999) // ID tidak ada di DB
	router.GET("/profile", handlers.Profile)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, newRequest(http.MethodGet, "/profile", nil))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, false, resp["success"])
}

// =============================================================
// TEST: RefreshCookies
// =============================================================

// TestRefreshCookies_NoToken — tidak butuh DB.
func TestRefreshCookies_NoToken(t *testing.T) {
	router := setupRouter()
	router.GET("/refresh", handlers.RefreshCookies)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, newRequest(http.MethodGet, "/refresh", nil))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, "Token is required", resp["message"])
}

// TestRefreshCookies_UserNotFound — butuh DB.
// Setelah fix `return` di auth_handler.go, harus return 500 tanpa double JSON.
func TestRefreshCookies_UserNotFound(t *testing.T) {
	requireDB(t)

	router := setupRouterWithAuth(99999)
	router.GET("/refresh", handlers.RefreshCookies)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, newRequest(http.MethodGet, "/refresh", nil))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := parseResponse(t, w) // hanya 1 JSON, tidak lagi double
	assert.Equal(t, false, resp["success"])
}

// =============================================================
// TEST: Logout
// =============================================================

// TestLogout_AlwaysSuccess — tidak butuh DB.
func TestLogout_AlwaysSuccess(t *testing.T) {
	router := setupRouter()
	router.POST("/logout", handlers.Logout)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, newRequest(http.MethodPost, "/logout", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, "Successfully logged out!", resp["message"])
}
