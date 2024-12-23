package gopay

import "github.com/jmoiron/sqlx"

// The global config variable holds the configuration for the application.
var config = new(Config)

// Config represents the configuration structure for the payment service.
type Config struct {
	DB     *sqlx.DB // Database connection, initialized using the sqlx package.
	Chains Chains   // Chains represents the blockchain networks supported by the service.
	Fiats  Fiats    // Fiats represents the supported fiat services (e.g., Stripe).
	Prefix string   // Prefix is used for table name prefix or query prefix (database-related).
}

// Setup initializes the payment service with the provided configuration.
// It applies migrations, sets up the configuration, and returns any errors encountered.
func Setup(cfg Config) error {
	// Run migrations using the provided database and table prefix.
	if err := runMigrate(cfg.DB, cfg.Prefix); err != nil {
		return err // If migration fails, return the error.
	}

	// Set the global configuration to the provided config.
	config = &cfg
	return nil // Return nil to indicate successful setup.
}
