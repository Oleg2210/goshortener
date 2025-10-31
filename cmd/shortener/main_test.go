package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplacePOST(t *testing.T) {

	type test_data struct {
		name string
		code int
	}

	test1 := test_data{
		name: "Good case",
		code: http.StatusCreated,
	}

	t.Run(test1.name, func(t *testing.T) {
		requestBody := strings.NewReader("https://yandex.kz")
		request := httptest.NewRequest(http.MethodPost, "/", requestBody)
		request.Header.Set("Content-Type", "text/plain")

		responseRecorder := httptest.NewRecorder()
		handlePost(responseRecorder, request)
		result := responseRecorder.Result()
		defer result.Body.Close()
		body, err := io.ReadAll(result.Body)

		assert.Equal(t, test1.code, result.StatusCode)
		require.NoError(t, err)
		assert.NotEmpty(t, string(body))
	})

}

func TestHandleGet(t *testing.T) {
	type test_data struct {
		name string
		code int
	}

	test1 := test_data{name: "No id", code: http.StatusBadRequest}

	t.Run(test1.name, func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/test/", nil)

		responseRecorder := httptest.NewRecorder()
		handleGet(responseRecorder, request)

		result := responseRecorder.Result()
		defer result.Body.Close()
		assert.Equal(t, test1.code, result.StatusCode)

	})

}
