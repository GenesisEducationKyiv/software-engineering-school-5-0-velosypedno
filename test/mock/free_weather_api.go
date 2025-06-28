package mock

import (
	"log"

	"net/http"
	"net/http/httptest"
)

const CityDoesNotExist = "InvalidCity"

func NewFreeWeatherAPI() *httptest.Server {
	handler := http.NewServeMux()

	handler.HandleFunc("/current.json", func(w http.ResponseWriter, r *http.Request) {
		city := r.URL.Query().Get("q")
		if city == CityDoesNotExist {
			http.Error(w, `{"error": {"code": 1006, "message": "No matching location found."}}`,
				http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		body := []byte(`{"current": {"temp_c": 20.0, "humidity": 80.0, "condition": {"text": "Sunny"}}}`)
		_, err := w.Write(body)
		if err != nil {
			log.Printf("free weather api: failed to write response body: %v", err)
		}
	})

	httpServer := httptest.NewServer(handler)
	return httpServer
}
