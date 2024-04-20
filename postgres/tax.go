package postgres

type TotalIncome struct {
	TotalIncome float64      `postgres:"total_income"`
	Wht         float64      `postgres:"wht"`
	Allowances  []Allowances `postgres:"allowances"`
}

type Allowances struct {
	AllowanceType string  `postgres:"allowance_type"`
	Amount        float64 `postgres:"amount"`
}

type Tax struct {
	Tax float64
}
