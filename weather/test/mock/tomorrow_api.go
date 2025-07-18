package mock

import (
	"log"

	"net/http"
	"net/http/httptest"
)

func NewTomorrowAPI() *httptest.Server {
	handler := http.NewServeMux()

	handler.HandleFunc("/weather/realtime", func(w http.ResponseWriter, r *http.Request) {
		city := r.URL.Query().Get("location")
		if city == CityDoesNotExist {
			errBody := `{"error": {"code": 400001, "message": "Not found.", "type": "error"}}`
			http.Error(w, errBody, http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		body := `{
			"data": {
				"values": {
					"temperature": 10000.0,
					"humidity": 100.0,
					"visibility": 12.7,
					"cloudCover": 0.1
				}
			}
		}`
		_, err := w.Write([]byte(body))
		if err != nil {
			log.Printf("tomorrow api: failed to write response body: %v", err)
		}
	})

	httpServer := httptest.NewServer(handler)
	return httpServer
}
