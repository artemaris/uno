package models

//go:generate easyjson

//easyjson:json
type APIRequest struct {
	URL string `json:"url"`
}

//easyjson:json
type APIResponse struct {
	Result string `json:"result"`
}
