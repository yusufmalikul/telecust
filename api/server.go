package api

import (
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
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/conversations", GetConversations)
		r.Get("/conversations/{id}/messages", GetConversationMessages)
		r.Post("/conversations/{id}/takeover", TakeOverConversation)
		r.Post("/conversations/{id}/activate-bot", ActivateBot)
		r.Post("/conversations/{id}/send", SendMessage)
		r.Get("/knowledge-base", GetKnowledgeBase)
		r.Put("/knowledge-base", UpdateKnowledgeBase)
	})

	// Serve static files
	r.Get("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.FileServer(http.Dir("./web")).ServeHTTP(w, r)
	}))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting HTTP server on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
