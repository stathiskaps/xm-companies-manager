package main

type CompanyType string

const (
	CORPORATION         CompanyType = "Corporation"
	NON_PROFIT          CompanyType = "NonProfit"
	COOPERATIVE         CompanyType = "Cooperative"
	SOLE_PROPRIETORSHIP CompanyType = "SoleProprietorship"
)

type Company struct {
	ID                string      `json:"id"`
	Name              string      `json:"name"`
	Description       string      `json:"description,omitempty"`
	AmountOfEmployees int         `json:"amount_of_employees"`
	Registered        bool        `json:"registered"`
	Type              CompanyType `json:"company_type,omitempty"`
}
