-- name: CreateCompany :one
INSERT INTO companies (
    id,
    name,
    description,
    amount_of_employees,
    registered,
    type
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;


-- name: GetCompany :one
SELECT *
FROM companies
WHERE id = $1;


-- name: PatchCompany :one
UPDATE companies
SET
    name = COALESCE(sqlc.narg('name')::text, name),
    description = COALESCE(sqlc.narg('description')::text, description),
    amount_of_employees = COALESCE(
        sqlc.narg('amount_of_employees')::integer,
        amount_of_employees
    ),
    registered = COALESCE(
        sqlc.narg('registered')::boolean,
        registered
    ),
    type = COALESCE(
        sqlc.narg('type')::text,
        type
    ),
    updated_at = NOW()
WHERE id = sqlc.arg('id')::uuid
RETURNING *;


-- name: DeleteCompany :execrows
DELETE FROM companies
WHERE id = $1;