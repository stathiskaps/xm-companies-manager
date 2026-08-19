-- +goose Up

CREATE TABLE companies (
    id UUID PRIMARY KEY,
    name VARCHAR(15) NOT NULL UNIQUE,
    description VARCHAR(3000),
    amount_of_employees INTEGER NOT NULL CHECK (amount_of_employees >= 0),
    registered BOOLEAN NOT NULL,
    type VARCHAR(32) NOT NULL CHECK (
        type IN (
            'Corporations',
            'NonProfit',
            'Cooperative',
            'Sole Proprietorship'
        )
    ),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down

DROP TABLE companies;