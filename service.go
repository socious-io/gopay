package gopay

import "github.com/jmoiron/sqlx"

type Service struct {
	db     *sqlx.DB
	chains Chains
	fiats  Fiats
}

func Setup(db *sqlx.DB, fiats Fiats, chains Chains) (*Service, error) {
	p := new(Service)
	if err := runMigrate(db); err != nil {
		return nil, err
	}
	p.db = db
	p.chains = chains
	p.fiats = fiats
	return p, nil
}
