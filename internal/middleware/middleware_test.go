package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Marst/reminder-app/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// =============================================================
// HELPERS
// =============================================================

// routeWithAuth membuat router dengan AuthMiddleware.
// Handler akhir set status 200 + kembalikan user_id & email dari context
// agar bisa di-assert di test.
func routeWithAuth() *gin.Engine {
	r := gin.New()
	r.GET("/protected", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		email, _ := c.Get("email")
		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"email":   email,
		})
	})
	return r
}

// signToken membuat JWT HS256 dengan secret dan exp yang diberikan.
func signToken(t *testing.T, secret string, userID int, email string, exp time.Time) string {
	t.Helper()
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	str, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("signToken: %v", err)
	}
	return str
}

// decodeBody decode JSON response body ke map.
func decodeBody(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("decodeBody: %v — raw: %s", err, w.Body.String())
	}
	return m
}

// =============================================================
// TEST: AuthMiddleware
// =============================================================

// TestAuthMiddleware_MissingHeader — tidak ada header Authorization sama sekali.
// Middleware harus return 401 dengan pesan "Token is missing!".
func TestAuthMiddleware_MissingHeader(t *testing.T) {
	router := routeWithAuth()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	resp := decodeBody(t, w)
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, "Token is missing!", resp["message"])
}

// TestAuthMiddleware_EmptyBearerValue — header ada tapi value setelah "Bearer " kosong.
// strings.Split akan menghasilkan slice dengan index[1] = "".
func TestAuthMiddleware_EmptyBearerValue(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	router := routeWithAuth()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Token string kosong → ValidateJWT return error
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	resp := decodeBody(t, w)
	assert.Equal(t, false, resp["success"])
}

// TestAuthMiddleware_RandomString — token bukan format JWT sama sekali.
func TestAuthMiddleware_RandomString(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	router := routeWithAuth()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer notajwttoken")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	resp := decodeBody(t, w)
	assert.Equal(t, false, resp["success"])
}

// TestAuthMiddleware_ExpiredToken — token valid formatnya tapi sudah expired.
// Ini juga men-trigger kondisi `!token.Valid` di ValidateJWT.
func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	secret := "test-secret"
	os.Setenv("JWT_SECRET", secret)
	defer os.Unsetenv("JWT_SECRET")

	expiredToken := signToken(t, secret, 1, "user@example.com",
		time.Now().Add(-time.Hour), // exp 1 jam lalu
	)

	router := routeWithAuth()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	resp := decodeBody(t, w)
	assert.Equal(t, false, resp["success"])
}

// TestAuthMiddleware_WrongSecret — token ditandatangani dengan secret yang berbeda.
func TestAuthMiddleware_WrongSecret(t *testing.T) {
	// Generate dengan secret A
	tokenStr := signToken(t, "secret-A", 1, "user@example.com",
		time.Now().Add(time.Hour),
	)

	// Validasi dengan secret B
	os.Setenv("JWT_SECRET", "secret-B")
	defer os.Unsetenv("JWT_SECRET")

	router := routeWithAuth()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	resp := decodeBody(t, w)
	assert.Equal(t, false, resp["success"])
}

// TestAuthMiddleware_ValidToken — happy path.
// Middleware harus set user_id dan email ke gin.Context dengan benar.
func TestAuthMiddleware_ValidToken(t *testing.T) {
	secret := "test-secret"
	os.Setenv("JWT_SECRET", secret)
	defer os.Unsetenv("JWT_SECRET")

	tokenStr := signToken(t, secret, 42, "user@example.com",
		time.Now().Add(time.Hour),
	)

	router := routeWithAuth()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBody(t, w)
	// JSON decode angka jadi float64
	assert.Equal(t, float64(42), resp["user_id"])
	assert.Equal(t, "user@example.com", resp["email"])
}
