package tax

import "fmt"

type TotalIncome struct {
	TotalIncome float64      `json:"totalIncome"`
	Wht         float64      `json:"wht"`
	Allowances  []Allowances `json:"allowances"`
}

type Allowances struct {
	AllowanceType string  `json:"allowanceType"`
	Amount        float64 `json:"amount"`
}

type Tax struct {
	Tax float64 `json:"tax"`
}

func calculateTotalIncome(income float64, donation float64, kreceipt float64) float64 {
	var totalIncome float64
	privateAllowance := 60000.0
	if donation >= 0 && kreceipt >= 0 {
		totalIncome = income - privateAllowance - donation - kreceipt
	}

	return totalIncome
}

func computeTax(totalIncome float64, prevlevel float64, taxrate float64, accumulate float64) float64 {
	return ((totalIncome - prevlevel) * taxrate) + accumulate
}

func determineTaxLevel(totalIncome float64) string {
	if totalIncome <= 0.0 {
		return "None"
	} else if totalIncome <= 150000.0 {
		return "0-150,000"
	} else if totalIncome > 150000 && totalIncome <= 500000 {
		return "150,001-500,000"
	} else if totalIncome > 500000 && totalIncome <= 1000000 {
		return "500,001-1,000,000"
	} else if totalIncome > 1000000 && totalIncome <= 2000000 {
		return "1,000,001-2,000,000"
	} else if totalIncome > 2000000 {
		return "2,000,000 ขึ้นไป"
	} else {
		return "None"
	}
}

func calculateTax(income float64, wht float64, donation float64, kreceipt float64) float64 {
	totalIncome := calculateTotalIncome(income, donation, kreceipt)
	if wht < 0 {
		fmt.Println("Please correct wht")
	}
	switch determineTaxLevel(totalIncome) {
	case "0-150,000":
		return 0.0
	case "150,001-500,000":
		return computeTax(totalIncome, 150000.0, 0.10, 0.0)
	case "500,001-1,000,000":
		return computeTax(totalIncome, 500000.0, 0.15, 50000.0)
	case "1,000,001-2,000,000":
		return computeTax(totalIncome, 1000000.0, 0.20, 150000.0)
	case "2,000,000 ขึ้นไป":
		return computeTax(totalIncome, 2000000.0, 0.35, 400000.0)
	default:
		return 0.0
	}
}
