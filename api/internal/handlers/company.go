package handlers

import (
	"errors"
	"net/http"

	"xm-companies-manager/internal/repos"
	"xm-companies-manager/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CompanyHandler struct {
	service *services.CompanyService
}

func NewCompanyHandler(service *services.CompanyService) *CompanyHandler {
	return &CompanyHandler{
		service: service,
	}
}

func (h *CompanyHandler) GetCompany(c *gin.Context) {
	id, err := parseUUID(c.Param("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid company id",
		})
		return
	}

	company, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "company not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get company",
		})
		return
	}

	c.JSON(http.StatusOK, company)
}

type CreateCompanyRequest struct {
	Name              string `json:"name" binding:"required,max=15"`
	Description       string `json:"description" binding:"max=3000"`
	AmountOfEmployees *int32 `json:"amount_of_employees" binding:"required,gte=0"`
	Registered        *bool  `json:"registered" binding:"required"`
	Type              string `json:"type" binding:"required"`
}

func (h *CompanyHandler) CreateCompany(c *gin.Context) {
	var req CreateCompanyRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if !validCompanyType(req.Type) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid company type",
		})
		return
	}

	company, err := h.service.Create(
		c.Request.Context(),
		repos.CreateCompanyParams{
			ID:                uuid.New(),
			Name:              req.Name,
			Description:       req.Description,
			AmountOfEmployees: *req.AmountOfEmployees,
			Registered:        *req.Registered,
			Type:              req.Type,
		},
	)

	if err != nil {
		if errors.Is(err, repos.ErrConflict) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "company name already exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create company",
		})
		return
	}

	c.JSON(http.StatusCreated, company)
}

type UpdateCompanyRequest struct {
	Name              *string `json:"name" binding:"omitempty,max=15"`
	Description       *string `json:"description" binding:"omitempty,max=3000"`
	AmountOfEmployees *int32  `json:"amount_of_employees" binding:"omitempty,gte=0"`
	Registered        *bool   `json:"registered"`
	Type              *string `json:"type"`
}

func (h *CompanyHandler) UpdateCompany(c *gin.Context) {
	id, err := parseUUID(c.Param("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid company id",
		})
		return
	}

	var req UpdateCompanyRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if req.Type != nil && !validCompanyType(*req.Type) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid company type",
		})
		return
	}

	company, err := h.service.Patch(
		c.Request.Context(),
		repos.PatchCompanyParams{
			ID:                id,
			Name:              req.Name,
			Description:       req.Description,
			AmountOfEmployees: req.AmountOfEmployees,
			Registered:        req.Registered,
			Type:              req.Type,
		},
	)

	if err != nil {
		switch {
		case errors.Is(err, repos.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "company not found",
			})

		case errors.Is(err, repos.ErrConflict):
			c.JSON(http.StatusConflict, gin.H{
				"error": "company name already exists",
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to update company",
			})
		}

		return
	}

	c.JSON(http.StatusOK, company)
}

func (h *CompanyHandler) DeleteCompany(c *gin.Context) {
	id, err := parseUUID(c.Param("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid company id",
		})
		return
	}

	err = h.service.Delete(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "company not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete company",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

func parseUUID(value string) (uuid.UUID, error) {
	return uuid.Parse(value)
}

func validCompanyType(t string) bool {
	switch t {
	case "Corporations",
		"NonProfit",
		"Cooperative",
		"Sole Proprietorship":
		return true
	default:
		return false
	}
}
