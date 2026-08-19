package routes

import (
	"xm-companies-manager/internal/handlers"

	"github.com/gin-gonic/gin"
)

func AddRoutes(r *gin.RouterGroup, h *handlers.CompanyHandler) {
	r.POST("/", h.CreateCompany)
	r.GET("/:companyId", h.GetCompany)
	r.PATCH("/:companyId", h.UpdateCompany)
	r.DELETE("/:companyId", h.DeleteCompany)
}
