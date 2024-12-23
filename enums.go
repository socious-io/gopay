package gopay

import "fmt"

// TransactionType represents the type of transaction (Deposit or Payout).
type TransactionType string

// NetworkType defines the type of blockchain network (e.g., EVM or Cardano).
type NetworkType string

// NetworkMode specifies the mode of the blockchain network (e.g., Mainnet or Testnet).
type NetworkMode string

// Currency defines the type of currency (e.g., USD or JPY).
type Currency string

// PaymentStatus represents the current status of a payment.
type PaymentStatus string

// PaymentType represents the type of payment (Fiat or Crypto).
type PaymentType string

// FiatService defines the payment service used for Fiat transactions (e.g., STRIPE).
type FiatService string

// Constants for different payment types.
const (
	FIAT   PaymentType = "FIAT"   // Payment type for Fiat currency.
	CRYPTO PaymentType = "CRYPTO" // Payment type for Cryptocurrencies.
)

// Constants for transaction types.
const (
	DEPOSIT TransactionType = "DEPOSIT" // Type of transaction where funds are deposited.
	PAYOUT  TransactionType = "PAYOUT"  // Type of transaction where funds are paid out.
)

// Constants for network types.
const (
	EVM     NetworkType = "EVM"     // Ethereum Virtual Machine network type.
	CARDANO NetworkType = "CARDANO" // Cardano blockchain network type.
)

// Constants for network modes.
const (
	MAINNET NetworkMode = "MAINNET" // Mainnet mode of operation.
	TESTNET NetworkMode = "TESTNET" // Testnet mode of operation.
)

// Constants for currencies.
const (
	USD Currency = "USD" // US Dollar currency.
	JPY Currency = "JPY" // Japanese Yen currency.
)

// Constants for payment status.
const (
	INITIATED       PaymentStatus = "INITIATED"       // Payment has been initiated.
	PENDING_DEPOSIT PaymentStatus = "PENDING_DEPOSIT" // Payment is awaiting deposit.
	DEPOSITED       PaymentStatus = "DEPOSITED"       // Payment has been deposited.
	ON_HOLD         PaymentStatus = "ON_HOLD"         // Payment is on hold.
	PAID_OUT        PaymentStatus = "PAID_OUT"        // Payment has been paid out.
	CANCLED         PaymentStatus = "CANCELED"        // Payment has been canceled.
	REFUNDED        PaymentStatus = "REFUNDED"        // Payment has been refunded.
)

// Constants for fiat services.
const (
	STRIPE FiatService = "STRIPE" // Fiat service provider for Stripe.
)

// scanEnum is a helper function that converts an interface{} value to a string
// to support database scanning. It handles both byte slices and string values.
func scanEnum(value interface{}, target interface{}) error {
	switch v := value.(type) {
	case []byte:
		*target.(*string) = string(v) // Convert byte slice to string.
	case string:
		*target.(*string) = v // Assign string value.
	default:
		return fmt.Errorf("failed to scan type: %v", value) // Error on unsupported type.
	}
	return nil
}

// Scan method for the TransactionType type. It scans a value and stores it as a TransactionType.
func (c *TransactionType) Scan(value interface{}) error {
	return scanEnum(value, (*string)(c)) // Calls scanEnum to handle the scanning process.
}

// Scan method for the NetworkType type. It scans a value and stores it as a NetworkType.
func (c *NetworkType) Scan(value interface{}) error {
	return scanEnum(value, (*string)(c)) // Calls scanEnum to handle the scanning process.
}

// Scan method for the NetworkMode type. It scans a value and stores it as a NetworkMode.
func (c *NetworkMode) Scan(value interface{}) error {
	return scanEnum(value, (*string)(c)) // Calls scanEnum to handle the scanning process.
}

// Scan method for the Currency type. It scans a value and stores it as a Currency.
func (c *Currency) Scan(value interface{}) error {
	return scanEnum(value, (*string)(c)) // Calls scanEnum to handle the scanning process.
}

// Scan method for the PaymentStatus type. It scans a value and stores it as a PaymentStatus.
func (c *PaymentStatus) Scan(value interface{}) error {
	return scanEnum(value, (*string)(c)) // Calls scanEnum to handle the scanning process.
}

// Scan method for the FiatService type. It scans a value and stores it as a FiatService.
func (c *FiatService) Scan(value interface{}) error {
	return scanEnum(value, (*string)(c)) // Calls scanEnum to handle the scanning process.
}
