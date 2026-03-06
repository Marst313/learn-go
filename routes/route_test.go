package routes_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Marst/reminder-app/internal/database"
	"github.com/Marst/reminder-app/routes"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Setenv("JWT_SECRET", "test-secret-for-handlers")
	database.ConnectTestDB()
	code := m.Run()
	os.Exit(code)
}

// requireDB men-skip test jika database.DB nil.
func requireDB(t *testing.T) {
	t.Helper()
	if database.DB == nil {
		t.Skip("Skipping: TEST_DATABASE_URL not set or DB connection failed")
	}
}

// =============================================================
// HELPER
// =============================================================

// newRouter membuat gin.Engine dengan semua routes terdaftar.
func newRouter() *gin.Engine {
	r := gin.New()
	routes.RegisterRoutes(r)
	return r
}

// makeToken membuat JWT valid untuk test endpoint yang butuh auth.
func makeToken(t *testing.T) string {
	t.Helper()
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "test-secret-for-handlers"
	}
	claims := jwt.MapClaims{
		"user_id": 1,
		"email":   "test@test.example.com",
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	str, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("makeToken: %v", err)
	}
	return str
}

func doRequest(router *gin.Engine, method, path, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// =============================================================
// TEST: NoRoute (404)
// =============================================================

func TestNoRoute_Returns404(t *testing.T) {
	router := newRouter()

	w := doRequest(router, http.MethodGet, "/api/v1/tidak-ada", "")

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestNoRoute_RandomPath(t *testing.T) {
	router := newRouter()

	w := doRequest(router, http.MethodGet, "/random/path/yang/tidak/ada", "")

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// =============================================================
// TEST: Auth Routes — endpoint publik (tanpa token)
// =============================================================

// POST /api/v1/auth/register — publik, tidak butuh token
func TestRoute_Register_Exists(t *testing.T) {
	router := newRouter()

	// Kirim body kosong → handler akan return 400, bukan 404
	// Ini membuktikan route terdaftar dengan benar
	w := doRequest(router, http.MethodPost, "/api/v1/auth/register", "")

	assert.NotEqual(t, http.StatusNotFound, w.Code)
}

// POST /api/v1/auth/login — publik, tidak butuh token
func TestRoute_Login_Exists(t *testing.T) {
	router := newRouter()

	w := doRequest(router, http.MethodPost, "/api/v1/auth/login", "")

	assert.NotEqual(t, http.StatusNotFound, w.Code)
}

// POST /api/v1/auth/logout — publik, tidak butuh token
func TestRoute_Logout_Exists(t *testing.T) {
	router := newRouter()

	w := doRequest(router, http.MethodPost, "/api/v1/auth/logout", "")

	assert.Equal(t, http.StatusOK, w.Code)
}

// =============================================================
// TEST: Auth Routes — endpoint protected (butuh token)
// =============================================================

// GET /api/v1/auth/profile — tanpa token harus 401
func TestRoute_Profile_NoToken_Returns401(t *testing.T) {
	router := newRouter()

	w := doRequest(router, http.MethodGet, "/api/v1/auth/profile", "")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// GET /api/v1/auth/profile — dengan token valid harus masuk ke handler (bukan 401/404)
// Butuh DB karena token lolos middleware dan handler langsung query DB.
func TestRoute_Profile_WithToken_PassesAuth(t *testing.T) {
	requireDB(t)

	router := newRouter()
	token := makeToken(t)

	w := doRequest(router, http.MethodGet, "/api/v1/auth/profile", token)

	// Lolos middleware (bukan 401), route ada (bukan 404)
	// Handler return 500 karena user ID=1 tidak ada di DB test
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}

// GET /api/v1/auth/refresh-cookies — tanpa token harus 401
func TestRoute_RefreshCookies_NoToken_Returns401(t *testing.T) {
	router := newRouter()

	w := doRequest(router, http.MethodGet, "/api/v1/auth/refresh-cookies", "")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// GET /api/v1/auth/refresh-cookies — dengan token valid harus lolos middleware
func TestRoute_RefreshCookies_WithToken_PassesAuth(t *testing.T) {
	requireDB(t)

	router := newRouter()
	token := makeToken(t)

	w := doRequest(router, http.MethodGet, "/api/v1/auth/refresh-cookies", token)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}

// =============================================================
// TEST: Reminder Routes — semua butuh token
// =============================================================

// GET /api/v1/reminders — tanpa token harus 401
func TestRoute_GetReminders_NoToken_Returns401(t *testing.T) {
	router := newRouter()

	w := doRequest(router, http.MethodGet, "/api/v1/reminders", "")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// GET /api/v1/reminders — dengan token valid harus lolos middleware
func TestRoute_GetReminders_WithToken_PassesAuth(t *testing.T) {
	requireDB(t)

	router := newRouter()
	token := makeToken(t)

	w := doRequest(router, http.MethodGet, "/api/v1/reminders", token)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}

// POST /api/v1/reminders — tanpa token harus 401
func TestRoute_NewReminder_NoToken_Returns401(t *testing.T) {
	router := newRouter()

	w := doRequest(router, http.MethodPost, "/api/v1/reminders", "")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// PATCH /api/v1/reminders/:id — tanpa token harus 401
func TestRoute_UpdateReminder_NoToken_Returns401(t *testing.T) {
	router := newRouter()

	w := doRequest(router, http.MethodPatch, "/api/v1/reminders/1", "")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// DELETE /api/v1/reminders/:id — tanpa token harus 401
func TestRoute_DeleteReminder_NoToken_Returns401(t *testing.T) {
	router := newRouter()

	w := doRequest(router, http.MethodDelete, "/api/v1/reminders/1", "")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// PATCH /api/v1/reminders/:id/toggle-complete — tanpa token harus 401
func TestRoute_ToggleComplete_NoToken_Returns401(t *testing.T) {
	router := newRouter()

	w := doRequest(router, http.MethodPatch, "/api/v1/reminders/1/toggle-complete", "")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// PATCH /api/v1/reminders/:id/toggle-complete — dengan token valid harus lolos middleware
func TestRoute_ToggleComplete_WithToken_PassesAuth(t *testing.T) {
	requireDB(t)

	router := newRouter()
	token := makeToken(t)

	w := doRequest(router, http.MethodPatch, "/api/v1/reminders/1/toggle-complete", token)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}

// =============================================================
// TEST: Method Not Allowed
// =============================================================

// GET ke endpoint yang hanya menerima POST harus 404 atau 405
func TestRoute_WrongMethod_Register(t *testing.T) {
	router := newRouter()

	w := doRequest(router, http.MethodGet, "/api/v1/auth/register", "")

	// Gin return 404 untuk method yang tidak terdaftar di route tersebut
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRoute_WrongMethod_Login(t *testing.T) {
	router := newRouter()

	w := doRequest(router, http.MethodDelete, "/api/v1/auth/login", "")

	assert.Equal(t, http.StatusNotFound, w.Code)
}
