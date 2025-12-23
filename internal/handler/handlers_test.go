package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Oleg2210/goshortener/internal/config"
	"github.com/Oleg2210/goshortener/internal/repository"
	"github.com/Oleg2210/goshortener/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplacePOST(t *testing.T) {

	type testData struct {
		name string
		code int
	}

	test1 := testData{
		name: "Good case",
		code: http.StatusCreated,
	}

	t.Run(test1.name, func(t *testing.T) {
		repo := repository.NewMemoryRepository()
		shortenerService := service.NewShortenerService(
			repo,
			config.MinLength,
			config.MaxLength,
			config.ContextUserID,
		)
		app := App{
			ShortenerService: shortenerService,
		}

		requestBody := strings.NewReader("https://yandex.kz")
		request := httptest.NewRequest(http.MethodPost, "/", requestBody)
		request.Header.Set("Content-Type", "text/plain")

		responseRecorder := httptest.NewRecorder()
		app.HandlePost(responseRecorder, request)
		result := responseRecorder.Result()
		defer result.Body.Close()
		body, err := io.ReadAll(result.Body)

		assert.Equal(t, test1.code, result.StatusCode)
		require.NoError(t, err)
		assert.NotEmpty(t, string(body))
	})

}

func TestHandleGet(t *testing.T) {
	type testData struct {
		name string
		code int
	}

	test1 := testData{name: "No id", code: http.StatusBadRequest}

	t.Run(test1.name, func(t *testing.T) {
		repo := repository.NewMemoryRepository()
		shortenerService := service.NewShortenerService(
			repo,
			config.MinLength,
			config.MaxLength,
			config.ContextUserID,
		)
		app := App{
			ShortenerService: shortenerService,
		}

		request := httptest.NewRequest(http.MethodGet, "/test/", nil)

		responseRecorder := httptest.NewRecorder()
		app.HandleGet(responseRecorder, request)

		result := responseRecorder.Result()
		defer result.Body.Close()

		assert.Equal(t, test1.code, result.StatusCode)
	})
}
