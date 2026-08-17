package main

import (
	"xm-companies-manager/api/src/handlers"

	"github.com/gin-gonic/gin"
)

func main() {

	r := gin.Default()

	companies := r.Group("/companies")

	companies.POST("", handlers.GetCompany)
	companies.PATCH("/:companyId", handlers.UpdateCompany)
	companies.DELETE("/:companyId", handlers.DeleteCompany)
	companies.GET("/:companyId", handlers.GetCompany)

	r.Run()

}
