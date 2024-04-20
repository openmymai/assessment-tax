package postgres

import (
	"log"

	"github.com/openmymai/assessment-tax/tax"
)

type Allowances struct {
	ID            int     `postgres:"id"`
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
		err := rows.Scan(&a.ID, &a.AllowanceType, &a.Amount)
		if err != nil {
			return nil, err
		}
		allowances = append(allowances, tax.Allowances{
			ID:            a.ID,
			AllowanceType: a.AllowanceType,
			Amount:        a.Amount,
		})
	}
	return allowances, nil
}

func (p *Postgres) UpdateAllowance(a tax.UpdateAllowance, id string) (tax.ReturnAllowance, error) {
	var updateResult tax.ReturnAllowance

	row := p.Db.QueryRow("UPDATE allowances SET amount = $2 WHERE id = $1 RETURNING amount", id, a.Amount)
	err := row.Scan(&id)
	if err != nil {
		log.Fatal(err)
	}

	updateResult.PersonalDeduction = a.Amount

	return updateResult, nil
}
