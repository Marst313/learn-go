package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Pastikan tidak ada env JWT_SECRET yang bocor dari environment lain
	os.Unsetenv("JWT_SECRET")
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	if err := json.Unmarshal(w.Body.Bytes(), v); err != nil {
		t.Fatalf("decodeBody: %v — raw: %s", err, w.Body.String())
	}
}

// =============================================================
// TEST: HashedPassword & CompareHashAndPassword
// =============================================================

func TestHashedPassword_Success(t *testing.T) {
	hashed, err := HashedPassword("mySecret123")

	assert.NoError(t, err)
	assert.NotEmpty(t, hashed)
	assert.NotEqual(t, "mySecret123", hashed)
}

func TestHashedPassword_EmptyString(t *testing.T) {
	// bcrypt tetap bisa hash string kosong
	hashed, err := HashedPassword("")

	assert.NoError(t, err)
	assert.NotEmpty(t, hashed)
}

func TestCompareHashAndPassword_Match(t *testing.T) {
	password := "correctPassword"
	hashed, _ := HashedPassword(password)

	err := CompareHashAndPassword(hashed, password)
	assert.NoError(t, err)
}

func TestCompareHashAndPassword_WrongPassword(t *testing.T) {
	hashed, _ := HashedPassword("correctPassword")

	err := CompareHashAndPassword(hashed, "wrongPassword")
	assert.Error(t, err)
	assert.Equal(t, "Invalid email or password", err.Error())
}

func TestCompareHashAndPassword_EmptyPassword(t *testing.T) {
	hashed, _ := HashedPassword("somePassword")

	err := CompareHashAndPassword(hashed, "")
	assert.Error(t, err)
}

// =============================================================
// TEST: GenerateJWT
// =============================================================

func TestGenerateJWT_Success(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	token, err := GenerateJWT(1, "user@example.com")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	// Pastikan format JWT: header.payload.signature
	assert.Equal(t, 3, len(splitDot(token)))
}

func TestGenerateJWT_WithEmptySecret(t *testing.T) {
	os.Unsetenv("JWT_SECRET")
	// HS256 masih bisa sign dengan secret string kosong
	token, err := GenerateJWT(1, "user@example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

// =============================================================
// TEST: ValidateJWT
// =============================================================

// TestValidateJWT_Success — happy path: token valid, claims ter-extract dengan benar.
func TestValidateJWT_Success(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	token, _ := GenerateJWT(42, "test@example.com")
	claims, err := ValidateJWT(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, float64(42), claims["user_id"].(float64))
	assert.Equal(t, "test@example.com", claims["email"].(string))
}

// TestValidateJWT_NoSecret — JWT_SECRET tidak di-set → error "JWT_SECRET is not set".
func TestValidateJWT_NoSecret(t *testing.T) {
	os.Unsetenv("JWT_SECRET")

	claims, err := ValidateJWT("any.token.here")

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, "JWT_SECRET is not set", err.Error())
}

// TestValidateJWT_RandomString — token sama sekali bukan JWT.
func TestValidateJWT_RandomString(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	claims, err := ValidateJWT("not-a-jwt-at-all")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

// TestValidateJWT_WrongSecret — token ditandatangani dengan secret berbeda.
func TestValidateJWT_WrongSecret(t *testing.T) {
	os.Setenv("JWT_SECRET", "original-secret")
	token, _ := GenerateJWT(1, "user@example.com")

	// Validasi dengan secret yang berbeda
	os.Setenv("JWT_SECRET", "different-secret")
	defer os.Unsetenv("JWT_SECRET")

	claims, err := ValidateJWT(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

// TestValidateJWT_UnexpectedSigningMethod — token menggunakan RS256, bukan HS256.
// Ini men-trigger branch `Unexpected signing method` di keyfunc.
func TestValidateJWT_UnexpectedSigningMethod(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	// Generate RSA key dan buat token RS256
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	rsaClaims := jwt.MapClaims{
		"user_id": 1,
		"email":   "test@example.com",
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
	rsaToken := jwt.NewWithClaims(jwt.SigningMethodRS256, rsaClaims)
	tokenString, err := rsaToken.SignedString(privateKey)
	assert.NoError(t, err)

	claims, err := ValidateJWT(tokenString)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "Unexpected signing method")
}

// TestValidateJWT_ExpiredToken — token format valid dan signature benar, tapi sudah expired.
// Ini men-trigger kondisi `!token.Valid` di ValidateJWT.
//
// Teknik: buat JWT secara manual dengan exp di masa lalu + sign dengan HMAC-SHA256
// langsung (bukan via library) agar jwt.Parse bisa memverifikasi signature
// lalu mendeteksi token expired → token.Valid = false.
func TestValidateJWT_ExpiredToken(t *testing.T) {
	secret := "test-secret-key"
	os.Setenv("JWT_SECRET", secret)
	defer os.Unsetenv("JWT_SECRET")

	expiredToken := buildManualJWT(t, secret, map[string]interface{}{
		"user_id": 1,
		"email":   "test@example.com",
		"exp":     time.Now().Add(-2 * time.Hour).Unix(), // 2 jam lalu
	})

	claims, err := ValidateJWT(expiredToken)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

// =============================================================
// TEST: MapReminderCreateError
// =============================================================

func newPQError(constraint string) *pq.Error {
	return &pq.Error{Constraint: constraint}
}

// TestMapReminderCreateError_NonPQError — error bukan *pq.Error → 500.
func TestMapReminderCreateError_NonPQError(t *testing.T) {
	status, msg := MapReminderCreateError(assert.AnError)

	assert.Equal(t, http.StatusInternalServerError, status)
	assert.Empty(t, msg)
}

// TestMapReminderCreateError_CategoryConstraint — constraint mengandung "category".
func TestMapReminderCreateError_CategoryConstraint(t *testing.T) {
	err := newPQError("reminders_category_check")
	status, msg := MapReminderCreateError(err)

	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, msg, "category")
}

// TestMapReminderCreateError_PriorityConstraint — constraint mengandung "priority".
func TestMapReminderCreateError_PriorityConstraint(t *testing.T) {
	err := newPQError("reminders_priority_check")
	status, msg := MapReminderCreateError(err)

	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, msg, "priority")
}

// TestMapReminderCreateError_UserIDConstraint — constraint mengandung "user_id".
func TestMapReminderCreateError_UserIDConstraint(t *testing.T) {
	err := newPQError("reminders_user_id_check")
	status, msg := MapReminderCreateError(err)

	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, msg, "user")
}

// TestMapReminderCreateError_FKeyConstraint — constraint berakhiran "_fkey".
func TestMapReminderCreateError_FKeyConstraint(t *testing.T) {
	err := newPQError("some_table_fkey")
	status, msg := MapReminderCreateError(err)

	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, msg, "reference")
}

// TestMapReminderCreateError_CheckConstraint — constraint berakhiran "_check".
func TestMapReminderCreateError_CheckConstraint(t *testing.T) {
	err := newPQError("some_field_check")
	status, msg := MapReminderCreateError(err)

	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, msg, "Invalid value")
}

// TestMapReminderCreateError_NotNullConstraint — constraint berakhiran "_not_null".
func TestMapReminderCreateError_NotNullConstraint(t *testing.T) {
	err := newPQError("some_column_not_null")
	status, msg := MapReminderCreateError(err)

	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, msg, "required")
}

// TestMapReminderCreateError_UnknownConstraint — constraint tidak cocok rule apapun → 500.
func TestMapReminderCreateError_UnknownConstraint(t *testing.T) {
	err := newPQError("xyz_unknown_constraint")
	status, msg := MapReminderCreateError(err)

	assert.Equal(t, http.StatusInternalServerError, status)
	assert.Equal(t, "Failed to create reminder", msg)
}

// =============================================================
// INTERNAL HELPERS
// =============================================================

// buildManualJWT membuat JWT secara manual menggunakan crypto/hmac langsung.
// Digunakan untuk membuat token dengan exp di masa lalu (expired)
// yang tetap memiliki signature valid — sehingga jwt.Parse bisa
// memverifikasi signature sebelum memeriksa exp, lalu set token.Valid = false.
func buildManualJWT(t *testing.T, secret string, claimsMap map[string]interface{}) string {
	t.Helper()

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	payloadJSON, err := json.Marshal(claimsMap)
	if err != nil {
		t.Fatalf("buildManualJWT: marshal claims: %v", err)
	}
	payload := base64.RawURLEncoding.EncodeToString(payloadJSON)

	signingInput := fmt.Sprintf("%s.%s", header, payload)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("%s.%s", signingInput, sig)
}

// splitDot memecah string berdasarkan karakter titik.
func splitDot(s string) []string {
	var parts []string
	start := 0
	for i, c := range s {
		if c == '.' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

// =============================================================
// TEST: JSONError
// =============================================================

func TestJSONError_StatusAndBody(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	JSONError(c, http.StatusBadRequest, "something went wrong")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp ErrorResponse
	decodeBody(t, w, &resp)
	assert.False(t, resp.Success)
	assert.Equal(t, "something went wrong", resp.Message)
	assert.Equal(t, http.StatusBadRequest, resp.ErrorCode)
}

func TestJSONError_500(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	JSONError(c, http.StatusInternalServerError, "internal error")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp ErrorResponse
	decodeBody(t, w, &resp)
	assert.False(t, resp.Success)
	assert.Equal(t, "internal error", resp.Message)
	assert.Equal(t, http.StatusInternalServerError, resp.ErrorCode)
}

func TestJSONError_401(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	JSONError(c, http.StatusUnauthorized, "unauthorized")

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp ErrorResponse
	decodeBody(t, w, &resp)
	assert.False(t, resp.Success)
	assert.Equal(t, "unauthorized", resp.Message)
	assert.Equal(t, http.StatusUnauthorized, resp.ErrorCode)
}

// TestJSONError_AbortsRequest memastikan request di-abort setelah JSONError
// sehingga handler berikutnya tidak dieksekusi.
func TestJSONError_AbortsRequest(t *testing.T) {
	router := gin.New()
	nextCalled := false

	router.GET("/test", func(c *gin.Context) {
		JSONError(c, http.StatusBadRequest, "stop here")
	}, func(c *gin.Context) {
		// Handler ini tidak boleh dipanggil karena JSONError pakai AbortWithStatusJSON
		nextCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, nextCalled, "next handler should not be called after JSONError")
}

// =============================================================
// TEST: JSONSuccess
// =============================================================

func TestJSONSuccess_WithData(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := map[string]string{"token": "abc123"}
	JSONSuccess(c, http.StatusOK, "operation successful", data)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp SuccessResponse
	decodeBody(t, w, &resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "operation successful", resp.Message)
	assert.NotNil(t, resp.Data)
}

func TestJSONSuccess_201Created(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	JSONSuccess(c, http.StatusCreated, "created!", map[string]int{"id": 42})

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp SuccessResponse
	decodeBody(t, w, &resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "created!", resp.Message)
}

func TestJSONSuccess_NilData(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	JSONSuccess(c, http.StatusOK, "logged out", nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp SuccessResponse
	decodeBody(t, w, &resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "logged out", resp.Message)
	assert.Nil(t, resp.Data)
}

func TestJSONSuccess_EmptyStringData(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	JSONSuccess(c, http.StatusOK, "ok", "")

	assert.Equal(t, http.StatusOK, w.Code)

	var resp SuccessResponse
	decodeBody(t, w, &resp)
	assert.True(t, resp.Success)
}
