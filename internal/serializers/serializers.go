package serializers

//easyjson:json
type Request struct {
	URL string `json:"url"`
}

//easyjson:json
type Response struct {
	Result string `json:"result"`
}
