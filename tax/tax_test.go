//go:build unit
// +build unit

package tax

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTaxCalculation(t *testing.T) {
	t.Run("Story 1 As user, I want to calculate my tax", func(t *testing.T) {
		body := bytes.NewBufferString(`{
			"totalIncome": 500000.0,
			"wht": 0.0,
			"allowances": [
				{
					"allowanceType": "donation",
					"amount": 0.0
				}
			]
		}`)

		var tax Tax

		res := request(http.MethodPost, uri("tax/calculations"), body)
		err := res.Decode(&tax)

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Greater(t, tax.Tax, 0.0)
	})

	t.Run("Story 2 As user, I want to calculate my tax with WHT", func(t *testing.T) {
		body := bytes.NewBufferString(`{
			"totalIncome": 500000.0,
			"wht": 25000.0,
			"allowances": [
				{
					"allowanceType": "donation",
					"amount": 0.0
				}
			]
		}`)

		var tax Tax

		res := request(http.MethodPost, uri("tax/calculations"), body)
		err := res.Decode(&tax)

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Greater(t, tax.Tax, 0.0)
	})

	t.Run("Story 3 As user, I want to calculate my tax with Donation", func(t *testing.T) {
		body := bytes.NewBufferString(`{
			"totalIncome": 500000.0,
			"wht": 0.0,
			"allowances": [
				{
					"allowanceType": "donation",
					"amount": 100000.0
				}
			]
		}`)

		var tax Tax

		res := request(http.MethodPost, uri("tax/calculations"), body)
		err := res.Decode(&tax)

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Greater(t, tax.Tax, 0.0)
	})

	t.Run("Story 4 As user, I want to calculate my tax and return detail", func(t *testing.T) {
		body := bytes.NewBufferString(`{
			"totalIncome": 500000.0,
			"wht": 0.0,
			"allowances": [
				{
					"allowanceType": "donation",
					"amount": 100000.0
				}
			]
		}`)

		var tax Tax
		var taxAmount float64

		res := request(http.MethodPost, uri("tax/calculations"), body)
		err := res.Decode(&tax)

		for _, taxLevel := range tax.TaxLevel {
			if taxLevel.Level == "150,001-500,000" {
				taxAmount = taxLevel.Tax
			}
		}

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Greater(t, tax.Tax, 0.0)
		assert.Greater(t, taxAmount, 0.0)
	})

	t.Run("Story 5 As admin, I want to setting personal deduction", func(t *testing.T) {
		body := bytes.NewBufferString(`{
			"amount": 60000.0
		}`)

		var personal ReturnAllowance
		var personalDeduction float64

		res := adminrequest(http.MethodPost, uri("admin/deductions/personal"), body)
		err := res.Decode(&personal)

		personalDeduction = personal.PersonalDeduction

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Greater(t, personalDeduction, 0.0)
	})

	t.Run("Story 6 As user, I want to calculate my tax with csv", func(t *testing.T) {
		body := bytes.NewBufferString(`{}`)

		var tax TaxUpload

		res := request(http.MethodPost, uri("tax/calculations/upload-csv"), body)
		err := res.Decode(&tax)

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Greater(t, len(tax.Taxes), 0)
	})

	t.Run("Story 7 As user, I want to calculate my tax with tax level detail", func(t *testing.T) {
		body := bytes.NewBufferString(`{
			"totalIncome": 500000.0,
			"wht": 0.0,
			"allowances": [
				{
					"allowanceType": "k-receipt",
					"amount": 50000.0
				},
				{
					"allowanceType": "donation",
					"amount": 100000.0
				}
			]
		}`)

		var tax Tax
		var taxAmount float64

		res := request(http.MethodPost, uri("tax/calculations"), body)
		err := res.Decode(&tax)

		for _, taxLevel := range tax.TaxLevel {
			if taxLevel.Level == "150,001-500,000" {
				taxAmount = taxLevel.Tax
			}
		}

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Greater(t, tax.Tax, 0.0)
		assert.Greater(t, taxAmount, 0.0)
	})

	t.Run("Story 8 As admin, I want to setting k-receipt deduction", func(t *testing.T) {
		body := bytes.NewBufferString(`{
			"amount": 80000.0
		}`)

		var kreceipt ReturnKreceipt
		var kreceiptDeduction float64

		res := adminrequest(http.MethodPost, uri("admin/deductions/k-receipt"), body)
		err := res.Decode(&kreceipt)

		kreceiptDeduction = kreceipt.Kreceipt

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Greater(t, kreceiptDeduction, 0.0)
	})

	t.Run("Tax Refund", func(t *testing.T) {
		body := bytes.NewBufferString(`{
			"totalIncome": 500000.0,
			"wht": 30000.0,
			"allowances": [
				{
					"allowanceType": "donation",
					"amount": 10000.0
				}
			]
		}`)

		var tax TaxRefund

		res := request(http.MethodPost, uri("tax/calculations"), body)
		err := res.Decode(&tax)

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Less(t, tax.TaxRefund, 0.0)
	})
}

func uri(paths ...string) string {
	host := "http://172.20.10.2:8080"
	if paths == nil {
		return host
	}

	url := append([]string{host}, paths...)
	return strings.Join(url, "/")
}

func request(method, url string, body io.Reader) *Response {
	req, _ := http.NewRequest(method, url, body)
	// req.Header.Add("Authorization", os.Getenv("AUTH_TOKEN"))
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	return &Response{res, err}
}

func adminrequest(method, url string, body io.Reader) *Response {
	req, _ := http.NewRequest(method, url, body)
	req.Header.Add("Authorization", "Basic YWRtaW5UYXg6YWRtaW4h")
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	return &Response{res, err}
}

type Response struct {
	*http.Response
	err error
}

func (r *Response) Decode(v interface{}) error {
	if r.err != nil {
		return r.err
	}

	return json.NewDecoder(r.Body).Decode(v)
}
