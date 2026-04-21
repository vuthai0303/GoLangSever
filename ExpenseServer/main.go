package main

import (
	"ExpenseServer/db"
	"ExpenseServer/graph"
	"ExpenseServer/middleware"
	"database/sql"
	"log"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func graphqlHandler(db *sql.DB) gin.HandlerFunc {
	h := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{DB: db}}))

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func playgroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL", "/query")

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or failed to load")
	}
	
	dbConn := db.InitDB()
	defer dbConn.Close()

	r := gin.Default()

	// Playground is public
	r.GET("/", playgroundHandler())

	// Graphql endpoint is protected by AuthMiddleware
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	protected.POST("/query", graphqlHandler(dbConn))

	log.Println("ExpenseServer running on :8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}
