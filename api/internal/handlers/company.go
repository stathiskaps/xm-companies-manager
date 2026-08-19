package handlers

import (
	"errors"
	"net/http"

	"xm-companies-manager/internal/repos"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CompanyHandler struct {
	repo *repos.CompanyRepository
}

func NewCompanyHandler(repo *repos.CompanyRepository) *CompanyHandler {
	return &CompanyHandler{
		repo: repo,
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

	company, err := h.repo.Get(c.Request.Context(), id)
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

	company, err := h.repo.Create(
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

func (h *CompanyHandler) UpdateCompany(c *gin.Context) {
	id, err := parseUUID(c.Param("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid company id",
		})
		return
	}

	var req PatchCompanyRequest

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

	company, err := h.repo.Patch(
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

	err = h.repo.Delete(c.Request.Context(), id)
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
