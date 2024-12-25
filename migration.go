package gopay

import (
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

// Migration struct defines a migration version and its SQL query.
type Migration struct {
	Version   string    // Version represents the migration version.
	Query     string    // Query is the SQL query to be executed for this migration.
	AppliedAt time.Time // AppliedAt is the timestamp when the migration was applied.
}

// List of migrations for the payment package, including enum creation
var migrations = []Migration{
	// Migration 1: Create ENUM types for transaction-related data (like transaction type, payment status).
	{
		Version: "2024-01-01-create-enums",
		Query: `-- Create custom ENUM types if not already created
		CREATE TYPE gopay_transaction_type AS ENUM ('DEPOSIT', 'PAYOUT');
		CREATE TYPE gopay_network_type AS ENUM ('EVM', 'CARDANO');
		CREATE TYPE gopay_network_mode AS ENUM ('MAINNET', 'TESTNET');
		CREATE TYPE gopay_currency AS ENUM ('USD', 'JPY');
		CREATE TYPE gopay_payment_type AS ENUM ('FIAT', 'CRYPTO');
		CREATE TYPE gopay_payment_status AS ENUM (
			'INITIATED', 'PENDING_DEPOSIT', 'DEPOSITED', 'ON_HOLD',
			'PAID_OUT', 'CANCELED', 'REFUNDED'
		);`,
	},
	// Migration 2: Create the payments table to track payment details.
	{
		Version: "2024-01-02-create-payments-table",
		Query: fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s_payments (
			id UUID PRIMARY KEY,
			tag TEXT,
			description TEXT,
			unique_ref TEXT UNIQUE NOT NULL,
			total_amount DECIMAL(20, 6),
			currency %s,
			fiat_service_name VARCHAR(32),
			crypto_currency TEXT,
			crypto_currency_rate DECIMAL(20, 6),
			meta JSONB,
			status %s DEFAULT INITIATED,
			type %s,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`, "{prefix}", "gopay_currency", "gopay_payment_status", "gopay_payment_type"),
	},
	// Migration 3: Create a table for payment identities linking payments to users.
	{
		Version: "2024-01-03-create-payment_identities-table",
		Query: fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s_payment_identities (
			id UUID PRIMARY KEY,
			payment_id UUID REFERENCES %s_payments(id) ON DELETE CASCADE,
			identity_id UUID NOT NULL,
			role_name TEXT,
			account TEXT NOT NULL,
			allocated_amount DECIMAL(20, 6),
			meta JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`, "{prefix}", "{prefix}"),
	},
	// Migration 4: Create a transactions table to track payment transactions.
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
		);`, "{prefix}", "{prefix}", "gopay_transaction_type"),
	},
	// Migration 5: Create the payment_migrations table to track which migrations have been applied.
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
