package dtos

type EncodeRequest struct {
	URL string `json:"url"`
}

type EncodeResponse struct {
	ShortURL string `json:"short_url"`
}
