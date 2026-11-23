package handler

type CreateURLResponse struct {
	Code        string `json:"code"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type CreateURLRequest struct {
	URL string `json:"url"`
}

type GetURLResponse struct {
	Code        string `json:"code"`
	OriginalURL string `json:"original_url"`
}
