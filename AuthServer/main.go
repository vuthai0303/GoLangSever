package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	db_config "AuthServer/db"
	_ "AuthServer/docs"
	Handlers "AuthServer/handlers"
	middleware "AuthServer/middleware"

	"github.com/joho/godotenv"
)

// @title AuthServer API
// @version 1.0
// @description REST API for AuthServer
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or failed to load")
	}

	sqliteDb := db_config.InitDB()
	defer sqliteDb.Close()

	env := &Handlers.Env{DB: sqliteDb}

	r := chi.NewRouter()
	r.Use(chi_middleware.Logger)
	r.Use(chi_middleware.Recoverer)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/api/auth", func(r chi.Router) {
		// Public routes
		r.Post("/signup", env.Signup)
		r.Post("/signin", env.Signin)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)
			r.Get("/user", env.GetUserInfo)
			r.Put("/user", env.UpdateUserInfo)
			r.Delete("/user", env.DeleteUser)
			r.Post("/lock", env.LockUser)
			r.Post("/refresh", env.RefreshToken)
		})
	})

	log.Println("AuthServer running on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
