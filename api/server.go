package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func StartServer() {
	r := chi.NewRouter()

	// Middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public routes (no auth required)
	r.Post("/api/login", Login)
	r.Post("/api/logout", Logout)
	r.Get("/login.html", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/login.html")
	}))

	// Protected API routes (auth required)
	r.Route("/api", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				session, _ := store.Get(r, "auth-session")
				if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
					return
				}
				next.ServeHTTP(w, r)
			})
		})

		r.Get("/conversations", GetConversations)
		r.Get("/conversations/{id}/messages", GetConversationMessages)
		r.Post("/conversations/{id}/takeover", TakeOverConversation)
		r.Post("/conversations/{id}/activate-bot", ActivateBot)
		r.Post("/conversations/{id}/send", SendMessage)
		r.Get("/knowledge-base", GetKnowledgeBase)
		r.Put("/knowledge-base", UpdateKnowledgeBase)
	})

	// Protected static files (auth required)
	r.With(AuthMiddleware).Get("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.FileServer(http.Dir("./web")).ServeHTTP(w, r)
	}))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting HTTP server on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
