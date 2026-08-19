package main

import (
	"context"
	"log"

	"xm-companies-manager/internal/config"
	"xm-companies-manager/internal/database"
	"xm-companies-manager/internal/database/sqlc"
	"xm-companies-manager/internal/events"
	"xm-companies-manager/internal/handlers"
	"xm-companies-manager/internal/middleware"
	"xm-companies-manager/internal/repos"
	"xm-companies-manager/internal/routes"
	"xm-companies-manager/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	pool, err := database.NewPool(ctx, cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	producer, err := events.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic)
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	queries := sqlc.New(pool)
	companyRepo := repos.NewCompanyRepository(queries)
	companyService := services.NewCompanyService(companyRepo, producer)
	companyHandler := handlers.NewCompanyHandler(companyService)

	r := wire(companyHandler, cfg.JWT.Secret)

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func wire(companyHandler *handlers.CompanyHandler, jwtSecret string) *gin.Engine {
	r := gin.Default()

	companiesGrp := r.Group("/companies")
	companiesGrp.Use(middleware.RequireAuth(jwtSecret))
	routes.AddRoutes(companiesGrp, companyHandler)

	return r
}
