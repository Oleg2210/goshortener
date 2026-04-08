package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/Oleg2210/goshortener/internal/config"
	"github.com/Oleg2210/goshortener/internal/entities"
	proto_serializers "github.com/Oleg2210/goshortener/internal/protobuf"
	"github.com/Oleg2210/goshortener/internal/serializers"
	"github.com/Oleg2210/goshortener/internal/service"
	"github.com/Oleg2210/goshortener/pkg/middleware/cookies"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// App represents the main HTTP application with services, logger, deleter, and audit publisher.
type App struct {
	ShortenerService *service.ShortenerService
	Logger           *zap.Logger
	Deleter          *Deleter
	Publisher        *AuditPublisher
}

// HandlePost handles plain text POST requests for shortening a single URL.
// It returns 201 Created on success or 409 Conflict if the URL already exists.
func (a *App) HandlePost(w http.ResponseWriter, r *http.Request) {
	returnStatus := http.StatusCreated
	body, err := io.ReadAll(r.Body)

	if err != nil {
		a.Logger.Error("failed to read request body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	fullURL := string(body)

	userID, _ := cookies.GetUserIDFromContext(r.Context())

	id, err := a.ShortenerService.Shorten(r.Context(), fullURL, userID)

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

	if a.Publisher != nil {
		event := AuditEvent{
			TS:     time.Now().Unix(),
			Action: "shorten",
			UserID: userID,
			URL:    fullURL,
		}
		a.Publisher.Publish(r.Context(), event)
	}

	fmt.Fprint(w, resolveURL)
}

// HandlePostJSON handles JSON POST requests for shortening a single URL.
// Expects {"url": "<original_url>"} and returns {"result": "<short_url>"} in JSON.
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

	userID, _ := cookies.GetUserIDFromContext(r.Context())
	id, err := a.ShortenerService.Shorten(r.Context(), req.URL, userID)

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

	if a.Publisher != nil {
		event := AuditEvent{
			TS:     time.Now().Unix(),
			Action: "shorten",
			UserID: userID,
			URL:    req.URL,
		}
		a.Publisher.Publish(r.Context(), event)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(returnStatus)
	w.Write(jsonBytes)
}

// HandlePostBatchJSON handles batch JSON POST requests for shortening multiple URLs.
// Expects a slice of {OriginalURL, CorrelationID}, responds with corresponding short URLs.
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

	userID, _ := cookies.GetUserIDFromContext(r.Context())

	err = a.ShortenerService.BatchShorten(r.Context(), records, userID)
	if err != nil {
		a.Logger.Error("error in batch saving", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var respItems serializers.BatchResponseItemSlice = make([]serializers.BatchResponseItem, 0, len(records))
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

// HandleGet handles GET requests for a short URL.
// Redirects to the original URL (307 Temporary Redirect) or returns 410 Gone if deleted.
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

	if a.Publisher != nil {
		userID, _ := cookies.GetUserIDFromContext(r.Context())
		event := AuditEvent{
			TS:     time.Now().Unix(),
			Action: "follow",
			UserID: userID,
			URL:    url.OriginalURL,
		}
		a.Publisher.Publish(r.Context(), event)
	}

	w.Header().Set("Location", url.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// HandlePing checks the health of the ShortenerService and responds with 200 OK if healthy.
func (a *App) HandlePing(w http.ResponseWriter, r *http.Request) {
	if pinged := a.ShortenerService.Ping(r.Context()); !pinged {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleGetAllUserUrls returns all URLs shortened by the current user.
// Responds with 200 OK and JSON array, or 204 No Content if no URLs exist.
func (a *App) HandleGetAllUserUrls(w http.ResponseWriter, r *http.Request) {
	userID, _ := cookies.GetUserIDFromContext(r.Context())
	records, err := a.ShortenerService.GetUserShortens(r.Context(), userID)

	if err != nil {
		a.Logger.Error("error while GetUserShortens", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(records) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var respItems serializers.AllShortenResponseItemSlice = make([]serializers.AllShortenResponseItem, 0, len(records))
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

// HandleMarkDelete queues URLs for deletion for the current user.
// Accepts JSON array of short URLs and responds with 202 Accepted.
func (a *App) HandleMarkDelete(w http.ResponseWriter, r *http.Request) {
	var req serializers.DeleteRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	userID, _ := cookies.GetUserIDFromContext(r.Context())

	a.Deleter.queue <- DeleteTask{UserID: userID, Shorts: req}

	w.WriteHeader(http.StatusAccepted)
}

// HandleGetStatistic returns number of saved urls and users count
// Checks a subnet
func (a *App) HandleGetStatistic(w http.ResponseWriter, r *http.Request) {

	if config.TrustedSubnet == "" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	ipStr := r.Header.Get("X-Real-IP")
	if ipStr == "" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	ip := net.ParseIP(ipStr)
	_, ipNet, err := net.ParseCIDR(config.TrustedSubnet)
	if ip == nil || !ipNet.Contains(ip) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	urls, users, err := a.ShortenerService.GetInternalStatistic(r.Context())

	if err != nil {
		a.Logger.Error("error while GetInternalStatistic", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response := serializers.StatiscticResponse{
		Urls:  urls,
		Users: users,
	}

	jsonBytes, err := response.MarshalJSON()
	if err != nil {
		a.Logger.Error("error in resonse serializing", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

type GRPCServer struct {
	proto_serializers.UnimplementedShortenerServiceServer
	App *App
}

func getUserIDFromMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	auth := md.Get("authorization")
	if len(auth) == 0 {
		return ""
	}

	return auth[0] // тут можешь распарсить токен если нужно
}

func (s *GRPCServer) ShortenURL(ctx context.Context, req *proto_serializers.URLShortenRequest) (*proto_serializers.URLShortenResponse, error) {
	userID := getUserIDFromMetadata(ctx)

	id, err := s.App.ShortenerService.Shorten(ctx, req.Url, userID)
	if err != nil {
		if errors.Is(err, service.ErrURLExists) {
			return nil, status.Error(codes.AlreadyExists, "url already exists")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resultURL, err := url.JoinPath(config.ResolveAddress, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto_serializers.URLShortenResponse{
		Result: resultURL,
	}, nil
}

func (s *GRPCServer) ExpandURL(ctx context.Context, req *proto_serializers.URLExpandRequest) (*proto_serializers.URLExpandResponse, error) {
	urlData, err := s.App.ShortenerService.GetURL(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if urlData.IsDeleted {
		return nil, status.Error(codes.FailedPrecondition, "url deleted")
	}

	return &proto_serializers.URLExpandResponse{
		Result: urlData.OriginalURL,
	}, nil
}

func (s *GRPCServer) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*proto_serializers.UserURLsResponse, error) {
	userID := getUserIDFromMetadata(ctx)

	records, err := s.App.ShortenerService.GetUserShortens(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(records) == 0 {
		return &proto_serializers.UserURLsResponse{}, nil
	}

	resp := &proto_serializers.UserURLsResponse{
		Url: make([]*proto_serializers.URLData, 0, len(records)),
	}

	for _, r := range records {
		shortURL, err := url.JoinPath(config.ResolveAddress, r.Short)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		resp.Url = append(resp.Url, &proto_serializers.URLData{
			ShortUrl:    shortURL,
			OriginalUrl: r.OriginalURL,
		})
	}

	return resp, nil
}
