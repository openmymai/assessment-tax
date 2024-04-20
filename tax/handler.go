package tax

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	store Storer
}

type Storer interface{}

func New(db Storer) *Handler {
	return &Handler{store: db}
}

type Err struct {
	Message string `json:"message"`
}

func (h *Handler) TaxCalculationsHandler(c echo.Context) error {
	var t TotalIncome
	err := c.Bind(&t)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}

	var donationAmount float64
	var kReceiptAmount float64
	for _, allowance := range t.Allowances {
		if allowance.AllowanceType == "donation" {
			donationAmount = allowance.Amount
		}
		if allowance.AllowanceType == "kreceipt" {
			kReceiptAmount = allowance.Amount
		}
	}

	if donationAmount < 0 || donationAmount > 100000 {
		return c.JSON(http.StatusBadRequest, Err{Message: "Donation should be > 0 or less than 100,000 THB"})
	}

	tax := Tax{
		Tax: calculateTax(t.TotalIncome, t.Wht, donationAmount, kReceiptAmount),
	}

	return c.JSON(http.StatusOK, tax)
}
