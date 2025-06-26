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

//easyjson:json
type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

//easyjson:json
type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

//easyjson:json
type BatchRequestList []BatchRequest

//easyjson:json
type BatchResponseList []BatchResponse

func UnmarshalBatchRequest(data []byte, v *[]BatchRequest) error {
	var list BatchRequestList
	if err := list.UnmarshalJSON(data); err != nil {
		return err
	}
	*v = list
	return nil
}

func MarshalBatchResponse(v []BatchResponse) ([]byte, error) {
	return BatchResponseList(v).MarshalJSON()
}

//easyjson:json
type UserURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	Deleted     bool   `json:"-"`
}
