package handlers

type CreateCompanyRequest struct {
	Name              string `json:"name" binding:"required,max=15"`
	Description       string `json:"description" binding:"max=3000"`
	AmountOfEmployees *int32 `json:"amount_of_employees" binding:"required,gte=0"`
	Registered        *bool  `json:"registered" binding:"required"`
	Type              string `json:"type" binding:"required"`
}

type PatchCompanyRequest struct {
	Name              *string `json:"name" binding:"omitempty,max=15"`
	Description       *string `json:"description" binding:"omitempty,max=3000"`
	AmountOfEmployees *int32  `json:"amount_of_employees" binding:"omitempty,gte=0"`
	Registered        *bool   `json:"registered"`
	Type              *string `json:"type"`
}
