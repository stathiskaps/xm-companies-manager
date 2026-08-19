package handlers

import (
	"errors"
	"net/http"

	"xm-companies-manager/internal/database/sqlc"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func GetCompany(queries *sqlc.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := parseUUID(c.Param("companyId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid company id",
			})
			return
		}

		company, err := queries.GetCompany(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
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
}

func CreateCompany(queries *sqlc.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		newID := uuid.New()

		company, err := queries.CreateCompany(
			c.Request.Context(),
			sqlc.CreateCompanyParams{
				ID: pgtype.UUID{
					Bytes: newID,
					Valid: true,
				},
				Name: req.Name,
				Description: pgtype.Text{
					String: req.Description,
					Valid:  req.Description != "",
				},
				AmountOfEmployees: *req.AmountOfEmployees,
				Registered:        *req.Registered,
				Type:              req.Type,
			},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to create company",
			})
			return
		}

		c.JSON(http.StatusCreated, company)
	}
}

func UpdateCompany(queries *sqlc.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		params := sqlc.PatchCompanyParams{
			ID: id,

			Name: pgtype.Text{
				String: valueOrZero(req.Name),
				Valid:  req.Name != nil,
			},

			Description: pgtype.Text{
				String: valueOrZero(req.Description),
				Valid:  req.Description != nil,
			},

			AmountOfEmployees: pgtype.Int4{
				Int32: valueOrZero(req.AmountOfEmployees),
				Valid: req.AmountOfEmployees != nil,
			},

			Registered: pgtype.Bool{
				Bool:  valueOrZero(req.Registered),
				Valid: req.Registered != nil,
			},

			Type: pgtype.Text{
				String: valueOrZero(req.Type),
				Valid:  req.Type != nil,
			},
		}

		company, err := queries.PatchCompany(
			c.Request.Context(),
			params,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "company not found",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to update company",
			})
			return
		}

		c.JSON(http.StatusOK, company)
	}
}

func DeleteCompany(queries *sqlc.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := parseUUID(c.Param("companyId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid company id",
			})
			return
		}

		rowsAffected, err := queries.DeleteCompany(
			c.Request.Context(),
			id,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to delete company",
			})
			return
		}

		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "company not found",
			})
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func parseUUID(value string) (pgtype.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return pgtype.UUID{}, err
	}

	return pgtype.UUID{
		Bytes: id,
		Valid: true,
	}, nil
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

func valueOrZero[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}

	return *v
}
