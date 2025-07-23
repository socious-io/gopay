package gopay

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
)

// Payment represents a payment transaction and its associated details.
type Payment struct {
	ID                 uuid.UUID          `db:"id" json:"id"`
	Tag                string             `db:"tag" json:"tag"`
	Description        string             `db:"description" json:"description"`
	UniqueRef          string             `db:"unique_ref" json:"unique_ref"`
	TotalAmount        float64            `db:"total_amount" json:"total_amount"`
	Currency           Currency           `db:"currency" json:"currency"`
	FiatServiceName    *string            `db:"fiat_service_name" json:"fiat_service_name"`
	CryptoCurrency     *string            `db:"crypto_currency" json:"crypto_currency"`
	CryptoCurrencyRate *float64           `db:"crypto_currency_rate" json:"crypto_currency_rate"`
	Meta               types.JSONText     `db:"meta" json:"meta,omitempty"`
	Status             PaymentStatus      `db:"status" json:"status"`
	TransactionStatus  *TransactionStatus `db:"transaction_status" json:"transaction_status"`
	ClientSecret       *string            `db:"client_secret" json:"client_secret"`
	Type               PaymentType        `db:"type" json:"type"`
	CreatedAt          time.Time          `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time          `db:"updated_at" json:"updated_at"`

	Identities   []PaymentIdentity `db:"-" json:"identities"`
	Transactions []Transaction     `db:"-" json:"transactions"`
}

// PaymentIdentity represents a payment identity associated with a payment.
type PaymentIdentity struct {
	ID              uuid.UUID      `db:"id" json:"id"`
	PaymentID       uuid.UUID      `db:"payment_id" json:"payment_id"`
	IdentityID      uuid.UUID      `db:"identity_id" json:"identity_id"`
	Account         string         `db:"account" json:"account"`
	RoleName        string         `db:"role_name" json:"role_name"`
	AllocatedAmount float64        `db:"allocated_amount" json:"allocated_amount"`
	Meta            types.JSONText `db:"meta" json:"meta,omitempty"`
	CreatedAt       time.Time      `db:"created_at" json:"created_at"`
}

// PaymentParams holds the parameters to create a new payment.
type PaymentParams struct {
	Tag         string
	Description string
	Ref         string
	Currency    Currency
	TotalAmount float64
	Type        PaymentType
	Meta        interface{}
}

// IdentityParams holds the parameters for creating a new payment identity.
type IdentityParams struct {
	ID       uuid.UUID
	RoleName string
	Account  string
	Amount   float64
	Meta     interface{}
}

// Table returns the table name for the Payment model, using the config prefix if available.
func (Payment) Table() string {
	if config.Prefix == "" {
		return "payments"
	}
	return fmt.Sprintf("%s_payments", config.Prefix)
}

// Table returns the table name for the PaymentIdentity model, using the config prefix if available.
func (PaymentIdentity) Table() string {
	if config.Prefix == "" {
		return "payment_identities"
	}
	return fmt.Sprintf("%s_payment_identities", config.Prefix)
}

// AddIdentity adds a payment identity to a payment, associating an identity with a payment and allocating an amount.
func (p *Payment) AddIdentity(params IdentityParams) (*PaymentIdentity, error) {
	// Convert meta to JSONB
	metaJSON, err := json.Marshal(params.Meta)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal meta: %w", err)
	}

	// Prepare PaymentIdentity struct for scanning
	identity := new(PaymentIdentity)

	// SQL query with RETURNING *
	query := `
		INSERT INTO %s (payment_id, identity_id, role_name, allocated_amount, meta, account)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING *`
	query = fmt.Sprintf(query, identity.Table())
	// Execute query and scan the returned row into the struct
	if err := config.DB.QueryRowx(query, p.ID, params.ID, params.RoleName, params.Amount, metaJSON, params.Account).
		StructScan(identity); err != nil {
		return nil, err
	}
	p.Identities = append(p.Identities, *identity)

	return identity, nil
}

// SetToCryptoMode sets the payment to crypto mode, specifying the address and rate.
func (p *Payment) SetToCryptoMode(address string, rate float64) error {
	// SQL query with RETURNING *
	query := `
		UPDATE %s
		SET crypto_currency = $1, crypto_currency_rate = $2, type = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING *`
	query = fmt.Sprintf(query, p.Table())
	// Execute query and scan the returned row back into the Payment struct
	if err := config.DB.QueryRowx(query, address, rate, CRYPTO, p.ID).StructScan(p); err != nil {
		return fmt.Errorf("failed to set payment to crypto mode: %w", err)
	}

	return nil
}

// SetToFiatMode sets the payment to fiat mode, specifying the fiat service name.
func (p *Payment) SetToFiatMode(name string) error {
	// SQL query with RETURNING *
	query := `
		UPDATE %s
		SET fiat_service_name = $1, type = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING *`
	query = fmt.Sprintf(query, p.Table())
	// Execute query and scan the updated row back into the Payment struct
	if err := config.DB.QueryRowx(query, name, FIAT, p.ID).
		StructScan(p); err != nil {
		return fmt.Errorf("failed to set payment to fiat mode: %w", err)
	}

	return nil
}

// UpdateStatus updates the status to the desired one
func (p *Payment) Update() error {
	// SQL query with RETURNING *
	query := `
		UPDATE %s
		SET status = $1, meta=$2, transaction_status=COALESCE($3, transaction_status), client_secret = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING *`
	query = fmt.Sprintf(query, p.Table())
	// Execute query and scan the updated row back into the Payment struct
	if err := config.DB.QueryRowx(query, p.Status, p.Meta, p.TransactionStatus, p.ClientSecret, p.ID).
		StructScan(p); err != nil {
		return fmt.Errorf("failed to set payment status to %s: %w", p.Status, err)
	}

	return nil
}

// Deposit processes the fiat deposit for the payment, creating a corresponding transaction.
func (p *Payment) Deposit() error {
	// Only fiat payments can call this
	if p.Type != FIAT {
		return fmt.Errorf("only fiat payments can call this")
	}

	// Ensure that identities are assigned before processing the deposit
	if len(p.Identities) < 1 {
		return fmt.Errorf("you need to assign identity first")
	}

	// Create a new transaction for the deposit
	t := &Transaction{
		PaymentID:  p.ID,
		IdentityID: p.Identities[0].ID,
		Tag:        string(DEPOSIT),
		Amount:     p.TotalAmount,
		Type:       DEPOSIT,
	}

	// Set parameters for fiat service payment
	params := FiatParams{
		ServiceName: *p.FiatServiceName,
		Customer:    p.Identities[0].Account,
		Currency:    p.Currency,
		Description: p.Description,
		Amount:      p.TotalAmount,
	}

	// Handle transfer between identities if applicable
	if len(p.Identities) > 1 {
		params.Transfer = &Transfer{
			Amount:      p.Identities[1].AllocatedAmount,
			Destination: p.Identities[1].Account,
		}
		t.Fee = params.Amount - params.Transfer.Amount
	}

	// Create the transaction record in the database
	if err := t.Create(); err != nil {
		return err
	}

	// Perform the fiat payment service
	info, err := config.Fiats.Pay(params)
	if err != nil {
		t.Meta, _ = json.Marshal(map[string]interface{}{"info": info, "error": err.Error()})
		t.Cancel()
		return err
	}

	// Store info in the transaction and verify if successful
	t.Meta, _ = json.Marshal(map[string]interface{}{"info": info})
	if !info.Confirmed && !info.RequiresAction {
		return t.Cancel()
	}

	if info.RequiresAction {
		// t.Status =
		/* if err := t.ActionRequired(); err != nil {
			return err
		} */
		status := ACTION_REQUIRED
		p.TransactionStatus = &status
		p.ClientSecret = &info.ClientSecret
		p.Status = ON_HOLD

		return p.Update()
	}

	if err := t.Verify(); err != nil {
		return err
	}

	p.Status = DEPOSITED
	return p.Update()
}

func (p *Payment) ConfirmPayment(paymentIntentID string) error {
	// Only fiat payments can call this
	if p.Type != FIAT {
		return fmt.Errorf("only fiat payments can call this")
	}

	// Only fiat payments can call this
	if p.Status != ON_HOLD || p.TransactionStatus == nil || *p.TransactionStatus != ACTION_REQUIRED {
		return fmt.Errorf("only on-hold payments can be confirmed")
	}

	//Fetch last transaction
	if len(p.Transactions) < 1 {
		return fmt.Errorf("this payment has no transaction available")
	}
	t := p.Transactions[len(p.Transactions)-1]

	// Perform the fiat payment service
	info, err := config.Fiats.ConfirmPayment(FiatPaymentConfirmParams{
		ServiceName:     *p.FiatServiceName,
		PaymentIntentID: paymentIntentID,
	})
	if err != nil {
		t.Meta, _ = json.Marshal(map[string]interface{}{"info": info, "error": err.Error()})
		t.Cancel()
		return err
	}

	// Store info in the transaction and verify if successful
	t.Meta, _ = json.Marshal(map[string]interface{}{"info": info})
	if !info.IsConfirmed {
		return fmt.Errorf("payment with intent ID of %s is not confirmed yet", paymentIntentID)
	}

	if err := t.Verify(); err != nil {
		return err
	}

	transactionStatus := VERIFIED
	p.TransactionStatus = &transactionStatus
	p.Status = DEPOSITED
	return p.Update()
}

// ConfirmDeposit processes a crypto payment deposit confirmation.
// It checks if the payment type is CRYPTO, creates a corresponding transaction,
// retrieves the transaction info from the blockchain, and verifies the deposit.
// If the deposit is not confirmed, the transaction is canceled.
func (p *Payment) ConfirmDeposit(txID string, meta interface{}) error {
	// Only allow CRYPTO payment types to call this method
	if p.Type != CRYPTO {
		return fmt.Errorf("only crypto payments can call this")
	}

	// Create a new transaction with deposit details
	t := &Transaction{
		PaymentID:  p.ID,
		TXID:       txID,
		IdentityID: p.Identities[0].ID,
		Tag:        string(DEPOSIT),
		Amount:     p.TotalAmount,
		Type:       DEPOSIT,
	}

	// Create the transaction in the database
	if err := t.Create(); err != nil {
		return err
	}

	// Set up parameters for the blockchain transaction info query
	params := CryptoParams{
		TxHash:       txID,
		TokenAddress: *p.CryptoCurrency,
	}

	// Get the transaction info from the blockchain
	info, err := config.Chains.TransactionInfo(params)
	if err != nil {
		// If there is an error, store the info and cancel the transaction
		t.Meta, _ = json.Marshal(map[string]interface{}{"info": info, "meta": meta, "error": err.Error()})
		t.Cancel()
		return err
	}

	// Store the info and check if the transaction is confirmed
	t.Meta, _ = json.Marshal(map[string]interface{}{"info": info, "meta": meta})
	if !info.Confirmed {
		// If the transaction is not confirmed, cancel it
		return t.Cancel()
	}

	if info.TotalAmount < t.Amount {
		return fmt.Errorf("transaction amount mismatch: expected %f but got %f", t.Amount, info.TotalAmount)
	}

	// Verify the transaction if it's confirmed
	if err := t.Verify(); err != nil {
		return err
	}

	p.Status = DEPOSITED
	p.Meta, _ = json.Marshal(meta)
	return p.Update()
}

// Fetch retrieves a payment by ID, including its associated identities and transactions.
func Fetch(id uuid.UUID) (*Payment, error) {
	p := new(Payment)
	// Fetch the payment record from the database
	if err := config.DB.Get(p, fmt.Sprintf(`SELECT * FROM %s WHERE id=$1`, p.Table()), id); err != nil {
		return nil, err
	}

	// Fetch identities associated with the payment
	if err := config.DB.Select(&p.Identities, fmt.Sprintf(`SELECT * FROM %s WHERE payment_id=$1`, PaymentIdentity{}.Table()), id); err != nil {
		return nil, err
	}

	// Fetch transactions associated with the payment
	if err := config.DB.Select(&p.Transactions, fmt.Sprintf(`SELECT * FROM %s WHERE payment_id=$1`, Transaction{}.Table()), id); err != nil {
		return nil, err
	}

	return p, nil
}

// Fetch retrieves a payment by Unique Reference, including its associated identities and transactions.
func FetchByUniqueRef(uniqueRef string) (*Payment, error) {
	p := new(Payment)
	// Fetch the payment record from the database
	if err := config.DB.Get(p, fmt.Sprintf(`SELECT * FROM %s WHERE unique_ref=$1`, p.Table()), uniqueRef); err != nil {
		return nil, err
	}

	// Fetch identities associated with the payment
	if err := config.DB.Select(&p.Identities, fmt.Sprintf(`SELECT * FROM %s WHERE payment_id=$1`, PaymentIdentity{}.Table()), p.ID); err != nil {
		return nil, err
	}

	// Fetch transactions associated with the payment
	if err := config.DB.Select(&p.Transactions, fmt.Sprintf(`SELECT * FROM %s WHERE payment_id=$1`, Transaction{}.Table()), p.ID); err != nil {
		return nil, err
	}

	return p, nil
}

// New creates a new payment with the specified parameters.
func New(params PaymentParams) (*Payment, error) {
	// Convert meta to JSONB
	metaJSON, err := json.Marshal(params.Meta)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal meta: %w", err)
	}

	// Prepare the Payment struct for scanning
	payment := new(Payment)

	// SQL query with RETURNING *
	query := `
		INSERT INTO %s (tag, description, unique_ref, total_amount, currency, status, type, meta)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (unique_ref) DO UPDATE SET tag=$1 ,description=$2 ,total_amount=$4, currency=$5, status=$6, type=$7, meta=$8
		RETURNING *`

	// Execute query and scan the returned row into the struct
	query = fmt.Sprintf(query, payment.Table())
	if err := config.DB.QueryRowx(query, params.Tag, params.Description, params.Ref, params.TotalAmount, params.Currency, INITIATED, params.Type, metaJSON).
		StructScan(payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return payment, nil
}
