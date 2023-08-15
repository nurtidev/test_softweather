package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateAccess(t *testing.T) {
	r, _ := http.NewRequest("GET", "/api/arithmetic", nil)
	assert.False(t, validateAccess(r))

	r.Header.Set("User-Access", "normaluser")
	assert.False(t, validateAccess(r))

	r.Header.Set("User-Access", "superuser")
	assert.True(t, validateAccess(r))
}

func TestValidateQuery(t *testing.T) {
	assert.True(t, validateQuery("2+2-3-5+1"))
	assert.True(t, validateQuery("1+2+3"))
	assert.True(t, validateQuery("123"))
	assert.False(t, validateQuery("invalid"))
	assert.False(t, validateQuery("2+2-3-5*1"))
	assert.False(t, validateQuery("2+2-3-5/1"))
	assert.False(t, validateQuery("こんにちは"))
}

func TestParseExpression(t *testing.T) {
	result, err := parseExpression("2 2-3-5 1")
	assert.Nil(t, err)
	assert.Equal(t, -3, result)

	result, err = parseExpression("1 2 3")
	assert.Nil(t, err)
	assert.Equal(t, 6, result)

	result, err = parseExpression("invalid")
	assert.Error(t, err)
}

func TestArithmeticHandler(t *testing.T) {
	r, _ := http.NewRequest("GET", "/api/arithmetic?q=2 2-3-5 1", nil)
	r.Header.Set("User-Access", "superuser")
	w := httptest.NewRecorder()
	arithmeticHandler(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, `{"result":-3}`, strings.TrimSpace(w.Body.String()))

	r, _ = http.NewRequest("GET", "/api/arithmetic", nil)
	r.Header.Set("User-Access", "superuser")
	w = httptest.NewRecorder()
	arithmeticHandler(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "Missing query parameter", strings.TrimSpace(w.Body.String()))

	r, _ = http.NewRequest("GET", "/api/arithmetic?q=invalid", nil)
	r.Header.Set("User-Access", "normaluser") // Здесь у нас нет доступа
	w = httptest.NewRecorder()
	arithmeticHandler(w, r)
	assert.Equal(t, http.StatusForbidden, w.Code) // Теперь мы ожидаем статус 403 Forbidden

	r, _ = http.NewRequest("GET", "/api/arithmetic?q=こんにちは", nil)
	r.Header.Set("User-Access", "superuser")
	w = httptest.NewRecorder()
	arithmeticHandler(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "Invalid query parameter", strings.TrimSpace(w.Body.String()))

	r, _ = http.NewRequest("GET", "/api/arithmetic?q="+strings.Repeat("1 ", 101), nil) // 202 символа
	r.Header.Set("User-Access", "superuser")
	w = httptest.NewRecorder()
	arithmeticHandler(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "Query parameter too long", strings.TrimSpace(w.Body.String()))
}

func BenchmarkValidateAccess(b *testing.B) {
	r, _ := http.NewRequest("GET", "/api/arithmetic", nil)
	r.Header.Set("User-Access", "superuser")
	for i := 0; i < b.N; i++ {
		validateAccess(r)
	}
}

func BenchmarkParseExpression(b *testing.B) {
	expr := "2 2-3-5 1"
	for i := 0; i < b.N; i++ {
		parseExpression(expr)
	}
}
