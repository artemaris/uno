package models

//go:generate easyjson

// APIRequest представляет запрос на сокращение URL через API
//easyjson:json
type APIRequest struct {
	URL string `json:"url"` // Оригинальный URL для сокращения
}

// APIResponse представляет ответ API с сокращенным URL
//easyjson:json
type APIResponse struct {
	Result string `json:"result"` // Сокращенный URL
}

// BatchRequest представляет запрос на пакетное сокращение URL
//easyjson:json
type BatchRequest struct {
	CorrelationID string `json:"correlation_id"` // Идентификатор корреляции для связи запроса и ответа
	OriginalURL   string `json:"original_url"`   // Оригинальный URL для сокращения
}

// BatchResponse представляет ответ на пакетное сокращение URL
//easyjson:json
type BatchResponse struct {
	CorrelationID string `json:"correlation_id"` // Идентификатор корреляции из запроса
	ShortURL      string `json:"short_url"`      // Сокращенный URL
}

// BatchRequestList представляет список запросов на пакетное сокращение
//easyjson:json
type BatchRequestList []BatchRequest

// BatchResponseList представляет список ответов на пакетное сокращение
//easyjson:json
type BatchResponseList []BatchResponse

// UnmarshalBatchRequest десериализует JSON данные в список запросов на пакетное сокращение
func UnmarshalBatchRequest(data []byte, v *[]BatchRequest) error {
	var list BatchRequestList
	if err := list.UnmarshalJSON(data); err != nil {
		return err
	}
	*v = list
	return nil
}

// MarshalBatchResponse сериализует список ответов на пакетное сокращение в JSON
func MarshalBatchResponse(v []BatchResponse) ([]byte, error) {
	return BatchResponseList(v).MarshalJSON()
}

// UserURL представляет URL пользователя с информацией о статусе удаления
//easyjson:json
type UserURL struct {
	ShortURL    string `json:"short_url"`    // Сокращенный URL
	OriginalURL string `json:"original_url"` // Оригинальный URL
	Deleted     bool   `json:"deleted"`      // Флаг удаления URL
}
