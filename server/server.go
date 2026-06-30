package server

import (
	"net/http"

	"github.com/stinkyfingers/poopjournal/auth"
	"github.com/stinkyfingers/poopjournal/handlers"
	"github.com/stinkyfingers/poopjournal/storage"
)

type Server struct {
	storage storage.Storage
}

func New(storage storage.Storage) (*Server, error) {
	return &Server{
		storage: storage,
	}, nil
}

func (s *Server) SetupRoutes() *http.ServeMux {
	err := auth.InitJWKS()
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()

	handler := handlers.NewHandler(s.storage)

	mux.HandleFunc("GET /userdata", auth.JWTMiddleware(http.HandlerFunc(handler.GetUserDataHandler)))

	mux.HandleFunc("POST /food", auth.JWTMiddleware(http.HandlerFunc(handler.AddFoodHandler)))
	mux.HandleFunc("PUT /food", auth.JWTMiddleware(http.HandlerFunc(handler.UpdateFoodHandler)))
	mux.HandleFunc("DELETE /food", auth.JWTMiddleware(http.HandlerFunc(handler.DeleteFoodHandler)))

	mux.HandleFunc("POST /poop", auth.JWTMiddleware(http.HandlerFunc(handler.AddPoopHandler)))
	mux.HandleFunc("PUT /poop", auth.JWTMiddleware(http.HandlerFunc(handler.UpdatePoopHandler)))
	mux.HandleFunc("DELETE /poop", auth.JWTMiddleware(http.HandlerFunc(handler.DeletePoopHandler)))

	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return mux
}

// CorsMiddleware wraps an http.Handler with CORS headers for all responses.
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
