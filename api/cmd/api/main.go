package main

import (
	"context"
	"log"

	"xm-companies-manager/internal/config"
	"xm-companies-manager/internal/database"
	"xm-companies-manager/internal/database/sqlc"
	"xm-companies-manager/internal/handlers"

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

	r := gin.Default()

	companies := r.Group("/companies")

	companies.POST("/", handlers.CreateCompany(queries))
	companies.PATCH("/:companyId", handlers.UpdateCompany(queries))
	companies.DELETE("/:companyId", handlers.DeleteCompany(queries))
	companies.GET("/:companyId", handlers.GetCompany(queries))

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
