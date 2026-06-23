package dtos

type DecodeRequest struct {
	ShortURL string `json:"short_url"`
}

type DecodeResponse struct {
	URL string `json:"url"`
}
