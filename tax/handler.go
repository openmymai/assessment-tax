package tax

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	store Storer
}

type Storer interface {
	GetAllowances() ([]Allowances, error)
	UpdateAllowance(allowance UpdateAllowance, id string) (ReturnAllowance, error)
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
	allowances, err := h.store.GetAllowances()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
	}
	for _, allowance := range allowances {
		if allowance.AllowanceType == "Personal" {
			personalAmount = allowance.Amount
		}
	}

	var donationAmount float64
	var kReceiptAmount float64
	for _, allowance := range t.Allowances {
		if allowance.AllowanceType == "donation" {
			donationAmount = allowance.Amount
		}
		if allowance.AllowanceType == "k-receipt" {
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

	if finalTax >= 0 {
		for _, tax := range taxLevel {
			if tax.Level == finalTaxLevel {
				tax.updateTax(finalTax)
				finalTaxLevelUpdate = append(finalTaxLevelUpdate, tax)
			} else {
				finalTaxLevelUpdate = append(finalTaxLevelUpdate, tax)
			}
		}
	}

	tax := Tax{
		Tax:      finalTax,
		TaxLevel: finalTaxLevelUpdate,
	}

	return c.JSON(http.StatusOK, tax)
}

func (h *Handler) SetPersonalDeductionHandler(c echo.Context) error {
	var a UpdateAllowance
	err := c.Bind(&a)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}
	id := "1"

	updateAllowance, err := h.store.UpdateAllowance(a, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, updateAllowance)
}
