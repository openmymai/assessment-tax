package tax

import (
	"fmt"
	"log"
	"os"

	"github.com/gocarina/gocsv"
)

type TotalIncome struct {
	TotalIncome float64      `json:"totalIncome" csv:"totalIncome"`
	Wht         float64      `json:"wht" csv:"wht"`
	Allowances  []Allowances `json:"allowances" csv:"allowances"`
}

type Allowances struct {
	ID            int     `json:"id"`
	AllowanceType string  `json:"allowanceType" csv:"allowanceTypes"`
	Amount        float64 `json:"amount" csv:"amount"`
}

type Tax struct {
	Tax      float64    `json:"tax"`
	TaxLevel []TaxLevel `json:"taxLevel"`
}

type TaxLevel struct {
	Level string  `json:"level"`
	Tax   float64 `json:"tax"`
}

type UpdateAllowance struct {
	Amount float64 `json:"amount"`
}

type ReturnAllowance struct {
	PersonalDeduction float64 `json:"personalDeduction"`
}

type TotalIncomeCsv struct {
	TotalIncome float64 `csv:"totalIncome"`
	Wht         float64 `csv:"wht"`
	Donation    float64 `csv:"donation"`
}

type TaxUpload struct {
	Taxes []TaxCSV `json:"taxes"`
}

type TaxCSV struct {
	TotalIncome float64 `json:"totalIncome"`
	Tax         float64 `json:"tax"`
}

func calculateTotalIncome(income float64, personal float64, donation float64, kreceipt float64) float64 {
	var totalIncome float64

	if personal >= 0 && donation >= 0 && kreceipt >= 0 {
		totalIncome = income - personal - donation - kreceipt
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
		return "2,000,001 ขึ้นไป"
	} else {
		return "None"
	}
}

func calculateTax(income float64, wht float64, personal float64, donation float64, kreceipt float64) float64 {
	totalIncome := calculateTotalIncome(income, personal, donation, kreceipt)
	if wht < 0 {
		fmt.Println("Please correct wht")
	}
	switch determineTaxLevel(totalIncome) {
	case "0-150,000":
		return 0.0 - wht
	case "150,001-500,000":
		return computeTax(totalIncome, 150000.0, 0.10, 0.0) - wht
	case "500,001-1,000,000":
		return computeTax(totalIncome, 500000.0, 0.15, 50000.0) - wht
	case "1,000,001-2,000,000":
		return computeTax(totalIncome, 1000000.0, 0.20, 150000.0) - wht
	case "2,000,001 ขึ้นไป":
		return computeTax(totalIncome, 2000000.0, 0.35, 400000.0) - wht
	default:
		return 0.0
	}
}

func taxFromFile(filename string) []TotalIncomeCsv {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var totalIncomeCsv []TotalIncomeCsv
	if err := gocsv.UnmarshalFile(file, &totalIncomeCsv); err != nil {
		log.Fatal(err)
	}

	return totalIncomeCsv
}
