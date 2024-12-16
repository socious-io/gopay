package gopay

import (
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

// Migration struct defines a migration version and its SQL query.
type Migration struct {
	Version   string
	Query     string
	AppliedAt time.Time
}

// List of migrations for the payment package, including enum creation
var migrations = []Migration{
	{
		Version: "2024-01-01-create-enums",
		Query: `-- Create custom ENUM types if not already created
		CREATE TYPE IF NOT EXISTS transaction_type AS ENUM ('DEPOSIT', 'PAYOUT');
		CREATE TYPE IF NOT EXISTS network_type AS ENUM ('EVM', 'CARDANO');
		CREATE TYPE IF NOT EXISTS network_mode AS ENUM ('MAINNET', 'TESTNET');
		CREATE TYPE IF NOT EXISTS currency AS ENUM ('USD', 'JPY');
		CREATE TYPE IF NOT EXISTS payment_status AS ENUM (
			'INITIATED', 'PENDING_DEPOSIT', 'DEPOSITED', 'ON_HOLD',
			'PAID_OUT', 'CANCELED', 'REFUNDED'
		);`,
	},
	{
		Version: "2024-01-02-create-payments-table",
		Query: fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s_payments (
			id UUID PRIMARY KEY,
			tag TEXT,
			description TEXT,
			unique_ref TEXT,
			total_amount DECIMAL(20, 6),
			currency %s,
			crypto_currency VARCHAR(10),
			crypto_currency_rate DECIMAL(20, 6),
			meta JSONB,
			status %s,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`, "{prefix}", "currency", "payment_status"),
	},
	{
		Version: "2024-01-03-create-payment_identities-table",
		Query: fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s_payment_identities (
			id UUID PRIMARY KEY,
			payment_id UUID REFERENCES %s_payments(id) ON DELETE CASCADE,
			identity_id UUID NOT NULL,
			role_name TEXT,
			allocated_amount DECIMAL(20, 6),
			meta JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`, "{prefix}", "{prefix}"),
	},
	{
		Version: "2024-01-04-create-transactions-table",
		Query: fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s_transactions (
			id UUID PRIMARY KEY,
			payment_id UUID REFERENCES %s_payments(id) ON DELETE CASCADE,
			identity_id UUID NOT NULL,
			tx_id TEXT NOT NULL,
			tag TEXT,
			amount DECIMAL(20, 6),
			fee DECIMAL(20, 6),
			discount DECIMAL(20, 6),
			type %s,
			meta JSONB,
			canceled_at TIMESTAMP,
			verified_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`, "{prefix}", "{prefix}", "transaction_type"),
	},
	{
		Version: "2024-01-05-create-payment_migrations-table",
		Query: fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s_payment_migrations (
			version VARCHAR(50) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`, "{prefix}"),
	},
}

// runMigrate applies any pending migrations for the payment package.
func runMigrate(db *sqlx.DB, prefix string) error {
	// Ensure the migrations table exists
	err := createMigrationsTable(db, prefix)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Check applied migrations
	appliedVersions, err := getAppliedMigrations(db, prefix)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Apply pending migrations
	for _, migration := range migrations {
		if _, applied := appliedVersions[migration.Version]; !applied {
			log.Printf("Applying migration: %s", migration.Version)
			query := migration.Query
			query = replacePrefix(query, prefix) // Replace `{prefix}` with the actual prefix
			_, err := db.Exec(query)
			if err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
			}

			// Record migration as applied
			err = recordMigration(db, prefix, migration.Version)
			if err != nil {
				return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
			}
		}
	}

	return nil
}

// replacePrefix replaces `{prefix}` in migration queries with the actual table prefix.
func replacePrefix(query, prefix string) string {
	return fmt.Sprintf(query, prefix)
}

// createMigrationsTable ensures the `payment_migrations` table exists with dynamic prefix.
func createMigrationsTable(db *sqlx.DB, prefix string) error {
	query := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s_payment_migrations (
		version VARCHAR(50) PRIMARY KEY,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`, prefix)
	_, err := db.Exec(query)
	return err
}

// getAppliedMigrations retrieves all applied migration versions with dynamic prefix.
func getAppliedMigrations(db *sqlx.DB, prefix string) (map[string]struct{}, error) {
	query := fmt.Sprintf(`SELECT version FROM %s_payment_migrations`, prefix)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]struct{})
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = struct{}{}
	}

	return applied, nil
}

// recordMigration records a migration as applied in the `payment_migrations` table with dynamic prefix.
func recordMigration(db *sqlx.DB, prefix, version string) error {
	query := fmt.Sprintf(`INSERT INTO %s_payment_migrations (version) VALUES ($1)`, prefix)
	_, err := db.Exec(query, version)
	return err
}
