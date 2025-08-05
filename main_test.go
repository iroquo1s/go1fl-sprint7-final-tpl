package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCafeNegative tests various negative scenarios for the cafe handler.
// It checks that the handler returns appropriate status codes and error messages
// for unknown cities and incorrect count parameters.
//
// Test cases include:
// - Request without city parameter
// - Request with unknown city
// - Request with invalid count parameter
func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()

		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		count int // передаваемое значение count
		want  int // ожидаемое количество кафе в ответе
	}{
		{0, 0},                         // count=0 — нет кафе
		{1, 1},                         // count=1 — одно кафе
		{2, 2},                         // count=2 — два кафе
		{100, len(cafeList["moscow"])}, // count=100 — максимум из общих кафе Москвы
	}

	for _, v := range requests {
		t.Run(fmt.Sprintf("count=%d", v.count), func(t *testing.T) {
			query := "/cafe?city=moscow&count=" + strconv.Itoa(v.count)

			response := httptest.NewRecorder()
			req := httptest.NewRequest("GET", query, nil)
			handler.ServeHTTP(response, req)

			assert.Equal(t, http.StatusOK, response.Code)

			body := response.Body.String()
			cafes := []string{}
			if body != "" {
				cafes = strings.Split(body, ",")
			}

			if v.want == 0 {
				assert.Empty(t, cafes, "expected empty list of cafes")
			} else {
				assert.Len(t, cafes, v.want, "the number of cafes doesn't match")
			}
		})
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		search    string // передаваемое значение search
		wantCount int    // ожидаемое количество найденных кафе
	}{
		{"фасоль", 0}, // не нашлось кафе с этим поиском
		{"кофе", 2},   // найдено два кафе
		{"вилка", 1},  // найдено одно кафе
	}

	for _, v := range requests {
		t.Run(fmt.Sprintf("search='%s'", v.search), func(t *testing.T) {
			query := "/cafe?city=moscow&search=" + v.search

			response := httptest.NewRecorder()
			req := httptest.NewRequest("GET", query, nil)
			handler.ServeHTTP(response, req)

			assert.Equal(t, http.StatusOK, response.Code)

			body := response.Body.String()
			cafes := []string{}
			if body != "" {
				cafes = strings.Split(body, ",")
			}

			assert.Len(t, cafes, v.wantCount, "number of found cafes mismatch")

			lowerSearch := strings.ToLower(v.search)
			for _, cafe := range cafes {
				assert.True(t, strings.Contains(strings.ToLower(cafe), lowerSearch),
					"cafe '%s' does not contain substring '%s'", cafe, v.search)
			}
		})
	}
}
