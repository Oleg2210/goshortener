package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/Oleg2210/goshortener/internal/config"
	"github.com/Oleg2210/goshortener/internal/entities"
	"github.com/Oleg2210/goshortener/internal/serializers"
	"github.com/Oleg2210/goshortener/internal/service"
	"go.uber.org/zap"
)

type App struct {
	ShortenerService *service.ShortenerService
	Logger           *zap.Logger
	Deleter          *Deleter
}

func (a *App) HandlePost(w http.ResponseWriter, r *http.Request) {
	returnStatus := http.StatusCreated
	body, err := io.ReadAll(r.Body)

	if err != nil {
		a.Logger.Error("failed to read request body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	fullURL := string(body)

	id, err := a.ShortenerService.Shorten(r.Context(), fullURL)

	if err != nil {
		if errors.Is(err, service.ErrURLExists) {
			returnStatus = http.StatusConflict
		} else {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
	}

	w.WriteHeader(returnStatus)

	resolveURL, err := url.JoinPath(config.ResolveAddress, id)
	if err != nil {
		a.Logger.Error("error while url join", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, resolveURL)
}

func (a *App) HandlePostJSON(w http.ResponseWriter, r *http.Request) {
	returnStatus := http.StatusCreated
	body, err := io.ReadAll(r.Body)

	if err != nil {
		a.Logger.Error("failed to read request body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var req serializers.Request

	if err := req.UnmarshalJSON(body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	id, err := a.ShortenerService.Shorten(r.Context(), req.URL)

	if err != nil {
		if errors.Is(err, service.ErrURLExists) {
			returnStatus = http.StatusConflict
		} else {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
	}

	resultURL, err := url.JoinPath(config.ResolveAddress, id)

	if err != nil {
		a.Logger.Error("error while url join", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := serializers.Response{
		Result: resultURL,
	}
	jsonBytes, _ := resp.MarshalJSON()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(returnStatus)
	w.Write(jsonBytes)
}

func (a *App) HandlePostBatchJSON(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		a.Logger.Error("failed to read request body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var reqItems serializers.BatchRequestItemSlice
	if err := reqItems.UnmarshalJSON(body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	records := make([]entities.URLRecord, 0, len(reqItems))
	for _, r := range reqItems {
		records = append(
			records,
			entities.URLRecord{
				OriginalURL: r.OriginalURL,
				Short:       r.CorrelationID,
			},
		)
	}

	err = a.ShortenerService.BatchShorten(r.Context(), records)
	if err != nil {
		a.Logger.Error("error in batch saving", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var respItems serializers.BatchResponseItemSlice
	for _, r := range records {
		resultURL, err := url.JoinPath(config.ResolveAddress, r.Short)

		if err != nil {
			a.Logger.Error("error while url join", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		response := serializers.BatchResponseItem{
			CorrelationID: r.Short,
			ShortURL:      resultURL,
		}
		respItems = append(respItems, response)
	}

	jsonBytes, err := respItems.MarshalJSON()
	if err != nil {
		a.Logger.Error("error in resonse serializing", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonBytes)
}

func (a *App) HandleGet(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[1:]
	url, err := a.ShortenerService.GetURL(r.Context(), id)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if url.IsDeleted {
		w.WriteHeader(http.StatusGone)
		return
	}

	w.Header().Set("Location", url.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (a *App) HandlePing(w http.ResponseWriter, r *http.Request) {
	if pinged := a.ShortenerService.Ping(r.Context()); !pinged {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *App) HandleUserUrls(w http.ResponseWriter, r *http.Request) {
	records, err := a.ShortenerService.GetUserShortens(r.Context())

	if err != nil {
		a.Logger.Error("error while GetUserShortens", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(records) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var respItems serializers.AllShortenResponseItemSlice
	for _, r := range records {
		resultURL, err := url.JoinPath(config.ResolveAddress, r.Short)

		if err != nil {
			a.Logger.Error("error while url join", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		response := serializers.AllShortenResponseItem{
			OriginalURL: r.OriginalURL,
			ShortURL:    resultURL,
		}
		respItems = append(respItems, response)
	}

	jsonBytes, err := respItems.MarshalJSON()
	if err != nil {
		a.Logger.Error("error in resonse serializing", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (a *App) HandleMarkDelete(w http.ResponseWriter, r *http.Request) {
	var req serializers.DeleteRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(config.ContextUserID).(string)

	for _, short := range req {
		a.Deleter.queue <- DeleteTask{UserID: userID, Short: short}
	}

	w.WriteHeader(http.StatusAccepted)
}
