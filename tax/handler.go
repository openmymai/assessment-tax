package tax

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	store Storer
}

type Storer interface {
	GetAllowances() ([]Allowances, error)
	UpdatePersonalAllowance(allowance UpdateAllowance, id string) (ReturnAllowance, error)
	UpdateKreceiptAllowance(allowance UpdateAllowance, allowance_type string) (ReturnKreceipt, error)
}

func New(db Storer) *Handler {
	return &Handler{store: db}
}

type Err struct {
	Message string `json:"message"`
}

func (t *TaxLevel) updateTax(tax float64) {
	(*t).Tax = tax
}

func (h *Handler) TaxCalculationsHandler(c echo.Context) error {
	var t TotalIncome
	err := c.Bind(&t)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}

	if t.Wht < 0 || t.Wht > t.TotalIncome {
		return c.JSON(http.StatusBadRequest, Err{Message: "With Holding Tax should be > 0 or less than your income"})
	}

	var taxLevel = []TaxLevel{
		{"0-150,000", 0.0},
		{"150,001-500,000", 0.0},
		{"500,001-1,000,000", 0.0},
		{"1,000,001-2,000,000", 0.0},
		{"2,000,001 ขึ้นไป", 0.0},
	}

	var finalTaxLevelUpdate = []TaxLevel{}

	var personalAmount float64
	var kReceiptMax float64
	allowances, err := h.store.GetAllowances()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
	}
	for _, allowance := range allowances {
		if allowance.AllowanceType == "Personal" {
			personalAmount = allowance.Amount
		}
		if allowance.AllowanceType == "Kreceipt" {
			kReceiptMax = allowance.Amount
		}
	}

	var donationAmount float64
	var kReceiptAmount float64
	for _, allowance := range t.Allowances {
		if allowance.AllowanceType == "donation" {
			donationAmount = allowance.Amount
		}
		if allowance.AllowanceType == "k-receipt" {
			if allowance.Amount < 0 || allowance.Amount > kReceiptMax {
				Kresponse := fmt.Sprintf("K-Receipt Maximum should not more than %.2f", kReceiptMax)
				return c.JSON(http.StatusBadRequest, Err{Message: Kresponse})
			}
			kReceiptAmount = allowance.Amount
		}
	}

	if donationAmount < 0 || donationAmount > 100000 {
		return c.JSON(http.StatusBadRequest, Err{Message: "Donation should be > 0 or less than 100,000 THB"})
	}

	if kReceiptAmount < 0 || kReceiptAmount > 100000 {
		return c.JSON(http.StatusBadRequest, Err{Message: "k-receipt should be > 0 or less than 100,000 THB"})
	}

	finalTax := calculateTax(t.TotalIncome, t.Wht, personalAmount, donationAmount, kReceiptAmount)
	finalTaxLevel := determineTaxLevel(calculateTotalIncome(t.TotalIncome, personalAmount, donationAmount, kReceiptAmount))

	var taxRefund TaxRefund

	if finalTax >= 0 {
		for _, tax := range taxLevel {
			if tax.Level == finalTaxLevel {
				tax.updateTax(finalTax)
				finalTaxLevelUpdate = append(finalTaxLevelUpdate, tax)
			} else {
				finalTaxLevelUpdate = append(finalTaxLevelUpdate, tax)
			}
		}
		tax := Tax{
			Tax:      finalTax,
			TaxLevel: finalTaxLevelUpdate,
		}
		return c.JSON(http.StatusOK, tax)
	} else {
		taxRefund.TaxRefund = finalTax
		return c.JSON(http.StatusOK, taxRefund)
	}
}

func (h *Handler) SetPersonalDeductionHandler(c echo.Context) error {
	var a UpdateAllowance
	err := c.Bind(&a)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}
	allowance_type := "Personal"

	if a.Amount < 10000 || a.Amount > 100000 {
		return c.JSON(http.StatusBadRequest, Err{Message: "Please set Personal Allowance range 10,000 - 100,000 THB"})
	}

	updateAllowance, err := h.store.UpdatePersonalAllowance(a, allowance_type)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, updateAllowance)
}

func (h *Handler) SetKreceiptDeductionHandler(c echo.Context) error {
	var a UpdateAllowance
	err := c.Bind(&a)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}
	allowanceType := "Kreceipt"

	if a.Amount < 0 || a.Amount > 100000 {
		return c.JSON(http.StatusBadRequest, Err{Message: "K-Receipt Allowance Maximum setting is 100,000 THB"})
	}

	updateAllowance, err := h.store.UpdateKreceiptAllowance(a, allowanceType)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, updateAllowance)
}

func (h *Handler) TaxCalculationsCSVHandler(c echo.Context) error {
	tax := taxFromFile("taxes.csv")
	var finaltaxcsv []TaxCSV
	var taxCsv TaxCSV

	for _, t := range tax {
		if t.Wht < 0 || t.Wht > t.TotalIncome {
			return c.JSON(http.StatusBadRequest, Err{Message: "With Holding Tax should be > 0 or less than your income"})
		}

		var personalAmount float64
		allowances, err := h.store.GetAllowances()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
		}
		for _, allowance := range allowances {
			if allowance.AllowanceType == "Personal" {
				personalAmount = allowance.Amount
			}
		}

		donationAmount := t.Donation
		kReceiptAmount := 0.0

		if donationAmount < 0 || donationAmount > 100000 {
			return c.JSON(http.StatusBadRequest, Err{Message: "Donation should be > 0 or less than 100,000 THB"})
		}

		finalTax := calculateTax(t.TotalIncome, t.Wht, personalAmount, donationAmount, kReceiptAmount)

		if finalTax >= 0 {
			taxCsv = TaxCSV{
				TotalIncome: t.TotalIncome,
				Tax:         finalTax,
			}
		}

		finaltaxcsv = append(finaltaxcsv, taxCsv)

	}

	taxOutput := TaxUpload{
		Taxes: finaltaxcsv,
	}

	return c.JSON(http.StatusOK, taxOutput)
}
