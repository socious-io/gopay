package gopay

import "github.com/jmoiron/sqlx"

var config *Config

type Config struct {
	DB     *sqlx.DB
	Chains Chains
	Fiats  Fiats
	Prefix string
}

func Setup(cfg Config) error {
	if err := runMigrate(cfg.DB, cfg.Prefix); err != nil {
		return err
	}
	config = &cfg
	return nil
}
