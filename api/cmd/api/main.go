package main

import (
	"context"
	"log"

	"xm-companies-manager/internal/config"
	"xm-companies-manager/internal/database"
	"xm-companies-manager/internal/database/sqlc"
	"xm-companies-manager/internal/handlers"
	"xm-companies-manager/internal/middleware"
	"xm-companies-manager/internal/repos"
	"xm-companies-manager/internal/routes"

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

	queries := sqlc.New(pool)
	companyRepo := repos.NewCompanyRepository(queries)

	r := wire(companyRepo, cfg.JWT.Secret)

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func wire(companyRepo *repos.CompanyRepository, jwtSecret string) *gin.Engine {
	r := gin.Default()

	companiesGrp := r.Group("/companies")
	companiesGrp.Use(middleware.RequireAuth(jwtSecret))

	companyHandler := handlers.NewCompanyHandler(companyRepo)

	routes.AddRoutes(companiesGrp, companyHandler)

	return r
}
