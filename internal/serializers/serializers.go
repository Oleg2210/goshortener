package serializers

//easyjson:json
type Request struct {
	URL string `json:"url"`
}

//easyjson:json
type Response struct {
	Result string `json:"result"`
}

// --- Request DTO for batch ---
//
//easyjson:json
type BatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// --- Response DTO for batch ---
//
//easyjson:json
type BatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

//easyjson:json
type BatchRequestItemSlice []BatchRequestItem

//easyjson:json
type BatchResponseItemSlice []BatchResponseItem
