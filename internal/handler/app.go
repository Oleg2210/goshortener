package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/Oleg2210/goshortener/internal/config"
	"github.com/Oleg2210/goshortener/internal/serializers"
	"github.com/Oleg2210/goshortener/internal/service"
)

type App struct {
	ShortenerService *service.ShortenerService
}

func (a *App) HandlePost(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	url := string(body)

	id, err := a.ShortenerService.Shorten(url)

	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	resolveAddress := config.ResolveAddress + "/%s"
	fmt.Fprintf(w, resolveAddress, id)
}

func (a *App) HandlePostJSON(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var req serializers.Request

	if err := req.UnmarshalJSON(body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	id, err := a.ShortenerService.Shorten(req.URL)

	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	resp := serializers.Response{
		Result: config.ResolveAddress + "/" + id,
	}
	jsonBytes, _ := resp.MarshalJSON()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonBytes)
}

func (a *App) HandleGet(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[1:]
	url, err := a.ShortenerService.GetUrl(id)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
