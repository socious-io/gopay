package gopay

import (
	"fmt"
	"time"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/paymentintent"
	"github.com/stripe/stripe-go/v81/paymentmethod"
)

// Fiats represents a slice of Fiat payment services.
type Fiats []Fiat

// Fiat represents a single fiat payment service provider such as Stripe.
type Fiat struct {
	Name    string      // The name of the payment service provider (e.g., "STRIPE").
	ApiKey  string      // The API key used to authenticate requests to the payment service.
	Service FiatService // The specific fiat service type (e.g., STRIPE).
}

// Transfer represents information about a transfer (e.g., recipient, amount).
type Transfer struct {
	Amount      float64 // The amount to transfer.
	Destination string  // The destination account for the transfer.
}

// FiatTransactionInfo holds information about a fiat transaction.
type FiatTransactionInfo struct {
	TXID        string      `json:"tx_id"`        // Transaction ID from payment gateway.
	TotalAmount int64       `json:"total_amount"` // Total transaction amount in minor units (e.g., cents).
	Currency    string      `json:"currency"`     // The currency used for the transaction.
	Meta        interface{} `json:"meta"`         // Metadata or additional information about the transaction.
	Date        time.Time   `json:"date"`         // The date the transaction was created.
	Confirmed   bool        `json:"confirmed"`    // Whether the payment has been confirmed.
}

// FiatParams contains parameters necessary for initiating a fiat transaction.
type FiatParams struct {
	ServiceName string    // The name of the service provider (e.g., "STRIPE").
	Customer    string    // Customer ID on the payment provider's system.
	Description string    // A description of the payment.
	Amount      float64   // The amount to be paid.
	Currency    Currency  // The currency for the payment (e.g., USD, JPY).
	Transfer    *Transfer // Information about a transfer (optional).
}

// Pay attempts to pay the specified service using the provided parameters.
func (fiats Fiats) Pay(params FiatParams) (*FiatTransactionInfo, error) {
	for _, f := range fiats {
		if params.ServiceName != f.Name {
			continue // Skip the service if it does not match the provided name.
		}
		switch f.Service {
		// TODO: add new fiat services here.
		default:
			// Default to Stripe if no specific service is added.
			return f.StripePay(params)
		}
	}
	return nil, fmt.Errorf("service %s could not found", params.ServiceName)
}

// StripePay handles a payment using the Stripe payment gateway.
func (f Fiat) StripePay(params FiatParams) (*FiatTransactionInfo, error) {
	// Set up the Stripe API key for authentication.
	stripe.Key = f.ApiKey

	// List payment methods for the customer.
	list := paymentmethod.List(&stripe.PaymentMethodListParams{
		Customer: stripe.String(params.Customer),
		Type:     stripe.String("card"),
	})
	var method *stripe.PaymentMethod
	for list.Next() {
		method = list.PaymentMethod() // Get the first payment method.
	}
	if err := list.Err(); err != nil {
		return nil, err // Return any error encountered during payment method listing.
	}
	if method == nil {
		return nil, fmt.Errorf("card method %s could not be found", params.Customer)
	}

	// Create payment intent parameters.
	intentParams := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(stripeAmount(params.Amount, params.Currency)),
		Currency:      stripe.String(string(params.Currency)),
		Customer:      stripe.String(params.Customer),
		PaymentMethod: stripe.String(method.ID),
		Description:   stripe.String(params.Description),
	}

	// If there is a transfer, add related data to the payment intent.
	if params.Transfer != nil {
		intentParams.ConfirmationMethod = stripe.String("automatic")
		intentParams.ApplicationFeeAmount = stripe.Int64(int64(params.Amount - params.Transfer.Amount))
		intentParams.OnBehalfOf = stripe.String(params.Transfer.Destination)
		intentParams.TransferData = &stripe.PaymentIntentTransferDataParams{
			Destination: stripe.String(params.Transfer.Destination),
		}
	}

	// Create the payment intent in Stripe.
	result, err := paymentintent.New(intentParams)
	if err != nil {
		return nil, err // Return any error encountered while creating the payment intent.
	}

	// Create transaction info using the result from Stripe.
	info := &FiatTransactionInfo{
		TXID:        result.ID,
		TotalAmount: result.Amount,
		Date:        time.Now(),
		Currency:    string(result.Currency),
		Meta:        result,
	}

	// Confirm the payment intent using the selected payment method.
	if _, err := paymentintent.Confirm(
		result.ID,
		&stripe.PaymentIntentConfirmParams{
			PaymentMethod: stripe.String(method.ID),
		},
	); err != nil {
		return info, err // Return the transaction info and any errors during confirmation.
	}

	info.Confirmed = true // Mark the transaction as confirmed if successful.
	return info, nil
}

// stripeAmount converts a floating point amount to the appropriate integer amount for the selected currency.
func stripeAmount(amount float64, currency Currency) int64 {
	switch currency {
	case USD:
		// Convert USD amount to cents.
		return int64(amount * 100)
	case JPY:
		// JPY is typically in whole units, so no conversion necessary.
		return int64(amount)
	default:
		// Default case returns 0 if the currency is unrecognized.
		return 0
	}
}
