package postgres

import (
	"log"

	"github.com/openmymai/assessment-tax/tax"
)

type Allowances struct {
	AllowanceType string  `postgres:"allowance_type"`
	Amount        float64 `postgres:"amount"`
}

func (p *Postgres) GetAllowances() ([]tax.Allowances, error) {
	rows, err := p.Db.Query("SELECT * FROM allowances")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allowances []tax.Allowances
	for rows.Next() {
		var a Allowances
		err := rows.Scan(&a.AllowanceType, &a.Amount)
		if err != nil {
			return nil, err
		}
		allowances = append(allowances, tax.Allowances{
			AllowanceType: a.AllowanceType,
			Amount:        a.Amount,
		})
	}
	return allowances, nil
}

func (p *Postgres) UpdatePersonalAllowance(a tax.UpdateAllowance, allowance_type string) (tax.ReturnAllowance, error) {
	var updateResult tax.ReturnAllowance

	row := p.Db.QueryRow("UPDATE allowances SET amount = $2 WHERE allowance_type = $1 RETURNING amount", allowance_type, a.Amount)
	err := row.Scan(&allowance_type)
	if err != nil {
		log.Fatal(err)
	}

	updateResult.PersonalDeduction = a.Amount

	return updateResult, nil
}

func (p *Postgres) UpdateKreceiptAllowance(a tax.UpdateAllowance, allowanceType string) (tax.ReturnKreceipt, error) {
	var updateResult tax.ReturnKreceipt

	row := p.Db.QueryRow("UPDATE allowances SET amount = $2 WHERE allowance_type = $1 RETURNING amount", allowanceType, a.Amount)
	err := row.Scan(&allowanceType)
	if err != nil {
		log.Fatal(err)
	}

	updateResult.Kreceipt = a.Amount

	return updateResult, nil
}
