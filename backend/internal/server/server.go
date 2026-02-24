package server

import (
	_ "embed"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/maxbrt/steamzip-app/internal/handlers"
	"github.com/maxbrt/steamzip-app/internal/utils"
)

//go:embed docs/docs.html
var docsHTML []byte

//go:embed docs/api.json
var docsJSON []byte

type Server struct {
	Router *chi.Mux
}

func NewServer() *Server {
	s := &Server{}
	s.Router = chi.NewRouter()
	return s
}

func (s *Server) MountHandlers(h *handlers.Handler) {
	allowedOrigins := strings.Split(
		utils.GetEnv("CORS_ALLOWED_ORIGINS", "http://localhost:8788"),
		",",
	)

	s.Router.Use(cors.Handler(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         300,
	}))

	s.Router.Use(middleware.Heartbeat("/health"))
	s.Router.Use(middleware.Logger)
	s.Router.Use(middleware.Recoverer)

	// Global rate limit: 200 requests/min per IP
	s.Router.Use(httprate.LimitByIP(200, time.Minute))

	// Security headers
	s.Router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
			next.ServeHTTP(w, r)
		})
	})

	s.Router.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(docsHTML)
	})
	s.Router.Get("/docs/api.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(docsJSON)
	})

	// Stripe webhook must be outside the /api group to avoid any future auth middleware
	s.Router.Post("/api/webhooks/stripe", h.HandleStripeWebhook)

	s.Router.Route("/api", func(r chi.Router) {
		// Tighter rate limit on session creation to prevent DB spam
		r.With(httprate.LimitByIP(10, time.Minute)).Post("/sessions", h.HandleCreateSession)
		r.Get("/sessions/{sessionId}/status", h.HandleGetSession)
		r.Post("/sessions/{sessionId}/assets/upload-url", h.HandleGetUploadURL)
		r.Post("/sessions/{sessionId}/assets/confirm-upload", h.HandleConfirmUpload)
		r.Get("/sessions/{sessionId}/assets/master", h.HandleGetMasterImage)
		r.Post("/sessions/{sessionId}/assets/process", h.HandleProcessAsset)
		r.Post("/sessions/{sessionId}/checkout", h.HandleCreateCheckout)
	})
}
