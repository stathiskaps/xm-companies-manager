package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"xm-companies-manager/internal/database/sqlc"
	"xm-companies-manager/internal/handlers"
	"xm-companies-manager/internal/repos"
	"xm-companies-manager/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type publishedEvent struct {
	eventType string
	companyID uuid.UUID
}

type fakeProducer struct {
	events []publishedEvent
}

func (p *fakeProducer) PublishCompanyEvent(
	ctx context.Context,
	eventType string,
	companyID uuid.UUID,
) error {
	p.events = append(p.events, publishedEvent{
		eventType: eventType,
		companyID: companyID,
	})
	return nil
}

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx := context.Background()

	container, err := postgres.Run(
		ctx,
		"postgres:18.6",
		postgres.WithDatabase("xm_test"),
		postgres.WithUsername("xm"),
		postgres.WithPassword("123456"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}

	testcontainers.CleanupContainer(t, container)

	connString, err := container.ConnectionString(
		ctx,
		"sslmode=disable",
	)
	if err != nil {
		t.Fatalf("get connection string: %v", err)
	}

	db, err := sql.Open("pgx", connString)
	if err != nil {
		t.Fatalf("open migration db: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err := goose.Up(db, "../../sql/migrations"); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		t.Fatalf("create pgx pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping postgres: %v", err)
	}

	t.Cleanup(pool.Close)

	return pool
}

func setupTestApp(t *testing.T) (
	*gin.Engine,
	*sqlc.Queries,
	*fakeProducer,
	string,
) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	pool := setupTestDB(t)
	queries := sqlc.New(pool)

	companyRepo := repos.NewCompanyRepository(queries)

	producer := &fakeProducer{}

	companyService := services.NewCompanyService(
		companyRepo,
		producer,
	)

	companyHandler := handlers.NewCompanyHandler(companyService)

	const jwtSecret = "integration-test-secret"

	router := wire(companyHandler, jwtSecret)
	token := generateTestJWT(t, jwtSecret)

	return router, queries, producer, token
}

func generateTestJWT(t *testing.T, secret string) string {
	t.Helper()

	now := time.Now()

	claims := jwt.RegisteredClaims{
		Issuer:    "xm-companies-manager",
		Subject:   "integration-test-user",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign JWT: %v", err)
	}

	return signed
}

func createTestCompany(
	t *testing.T,
	router http.Handler,
	token string,
	name string,
) string {
	t.Helper()

	body := fmt.Sprintf(`{
		"name": %q,
		"description": "Integration test company",
		"amount_of_employees": 120,
		"registered": true,
		"type": "Corporations"
	}`, name)

	req := httptest.NewRequest(
		http.MethodPost,
		"/companies/",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusCreated, recorder.Code)

	var response struct {
		ID string `json:"ID"`
	}

	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.NotEmpty(t, response.ID)

	return response.ID
}

func TestGetCompany(t *testing.T) {
	router, _, _, token := setupTestApp(t)

	companyID := createTestCompany(
		t,
		router,
		token,
		"Get Test",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/companies/"+companyID,
		nil,
	)

	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response struct {
		ID                string `json:"ID"`
		Name              string `json:"Name"`
		Description       string `json:"Description"`
		AmountOfEmployees int32  `json:"AmountOfEmployees"`
		Registered        bool   `json:"Registered"`
		Type              string `json:"Type"`
	}

	require.NoError(
		t,
		json.Unmarshal(recorder.Body.Bytes(), &response),
	)

	assert.Equal(t, companyID, response.ID)
	assert.Equal(t, "Get Test", response.Name)
	assert.Equal(t, "Integration test company", response.Description)
	assert.Equal(t, int32(120), response.AmountOfEmployees)
	assert.True(t, response.Registered)
	assert.Equal(t, "Corporations", response.Type)
}

func TestGetCompanyNotFound(t *testing.T) {
	router, _, _, token := setupTestApp(t)

	nonExistentID := uuid.New().String()

	req := httptest.NewRequest(
		http.MethodGet,
		"/companies/"+nonExistentID,
		nil,
	)

	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var response struct {
		Error string `json:"error"`
	}

	require.NoError(
		t,
		json.Unmarshal(recorder.Body.Bytes(), &response),
	)

	assert.Equal(t, "company not found", response.Error)
}

func TestPatchCompany(t *testing.T) {
	router, queries, _, token := setupTestApp(t)

	companyID := createTestCompany(
		t,
		router,
		token,
		"Patch Test",
	)

	body := `{
		"description": "Updated description",
		"amount_of_employees": 250,
		"registered": false
	}`

	req := httptest.NewRequest(
		http.MethodPatch,
		"/companies/"+companyID,
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response struct {
		Name              string `json:"Name"`
		Description       string `json:"Description"`
		AmountOfEmployees int32  `json:"AmountOfEmployees"`
		Registered        bool   `json:"Registered"`
		Type              string `json:"Type"`
	}

	require.NoError(
		t,
		json.Unmarshal(recorder.Body.Bytes(), &response),
	)

	// Changed fields
	assert.Equal(t, "Updated description", response.Description)
	assert.Equal(t, int32(250), response.AmountOfEmployees)
	assert.False(t, response.Registered)

	// Untouched fields
	assert.Equal(t, "Patch Test", response.Name)
	assert.Equal(t, "Corporations", response.Type)

	// Verify persisted state directly in PostgreSQL
	id, err := uuid.Parse(companyID)
	require.NoError(t, err)

	company, err := queries.GetCompany(
		context.Background(),
		pgtype.UUID{
			Bytes: id,
			Valid: true,
		},
	)
	require.NoError(t, err)

	assert.Equal(t, int32(250), company.AmountOfEmployees)
	assert.False(t, company.Registered)

	assert.True(t, company.Description.Valid)
	assert.Equal(t, "Updated description", company.Description.String)
}

func TestDeleteCompany(t *testing.T) {
	router, queries, _, token := setupTestApp(t)

	companyID := createTestCompany(
		t,
		router,
		token,
		"Delete Test",
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/companies/"+companyID,
		nil,
	)

	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNoContent, recorder.Code)

	id, err := uuid.Parse(companyID)
	require.NoError(t, err)

	_, err = queries.GetCompany(
		context.Background(),
		pgtype.UUID{
			Bytes: id,
			Valid: true,
		},
	)

	assert.ErrorIs(t, err, pgx.ErrNoRows)
}

func TestCreateCompany(t *testing.T) {
	router, queries, producer, token := setupTestApp(t)

	body := `{
		"name": "XM Test",
		"description": "Integration test company",
		"amount_of_employees": 120,
		"registered": true,
		"type": "Corporations"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/companies/",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusCreated, recorder.Code)

	var response struct {
		ID                string `json:"ID"`
		Name              string `json:"Name"`
		Description       string `json:"Description"`
		AmountOfEmployees int32  `json:"AmountOfEmployees"`
		Registered        bool   `json:"Registered"`
		Type              string `json:"Type"`
	}

	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.ID)
	assert.Equal(t, "XM Test", response.Name)
	assert.Equal(t, "Integration test company", response.Description)
	assert.Equal(t, int32(120), response.AmountOfEmployees)
	assert.True(t, response.Registered)
	assert.Equal(t, "Corporations", response.Type)
	require.Len(t, producer.events, 1)

	assert.Equal(t, "company.created", producer.events[0].eventType)
	assert.Equal(t, response.ID, producer.events[0].companyID.String())

	id, err := uuid.Parse(response.ID)
	require.NoError(t, err)

	company, err := queries.GetCompany(
		context.Background(),
		pgtype.UUID{
			Bytes: id,
			Valid: true,
		},
	)
	require.NoError(t, err)

	assert.Equal(t, "XM Test", company.Name)
	assert.Equal(t, int32(120), company.AmountOfEmployees)
	assert.True(t, company.Registered)
	assert.Equal(t, "Corporations", company.Type)
}

func TestCreateCompanyDuplicateName(t *testing.T) {
	router, _, _, token := setupTestApp(t)

	companyName := "Duplicate Test"

	createTestCompany(
		t,
		router,
		token,
		companyName,
	)

	body := fmt.Sprintf(`{
		"name": %q,
		"description": "Another company with the same name",
		"amount_of_employees": 50,
		"registered": true,
		"type": "NonProfit"
	}`, companyName)

	req := httptest.NewRequest(
		http.MethodPost,
		"/companies/",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusConflict, recorder.Code)

	var response struct {
		Error string `json:"error"`
	}

	require.NoError(
		t,
		json.Unmarshal(recorder.Body.Bytes(), &response),
	)

	assert.Equal(t, "company name already exists", response.Error)
}
