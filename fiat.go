package gopay

import (
	"fmt"
	"log"
	"time"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/account"
	"github.com/stripe/stripe-go/v81/accountlink"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/paymentintent"
	"github.com/stripe/stripe-go/v81/paymentmethod"
)

// Fiats represents a slice of Fiat payment services.
type Fiats []Fiat

// Fiat represents a single fiat payment service provider such as Stripe.
type Fiat struct {
	Name     string      `mapstructure:"name"`     // The name of the payment service provider (e.g., "STRIPE").
	ApiKey   string      `mapstructure:"apikey"`   // The API key used to authenticate requests to the payment service.
	Callback string      `mapstructure:"callback"` // The API key used to authenticate requests to the payment service.
	Service  FiatService `mapstructure:"service"`  // The specific fiat service type (e.g., STRIPE).
}

// Transfer represents information about a transfer (e.g., recipient, amount).
type Transfer struct {
	Amount      float64 // The amount to transfer.
	Destination string  // The destination account for the transfer.
}

// FiatTransactionInfo holds information about a fiat transaction.
type FiatTransactionInfo struct {
	TXID           string      `json:"tx_id"`        // Transaction ID from payment gateway.
	TotalAmount    int64       `json:"total_amount"` // Total transaction amount in minor units (e.g., cents).
	Currency       string      `json:"currency"`     // The currency used for the transaction.
	Meta           interface{} `json:"meta"`         // Metadata or additional information about the transaction.
	Date           time.Time   `json:"date"`         // The date the transaction was created.
	Confirmed      bool        `json:"confirmed"`    // Whether the payment has been confirmed.
	RequiresAction bool        `json:"requires_action"`
	ClientSecret   string      `json:"client_secret"`
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

// FiatPaymentConfirmParams contains parameters necessary for confirming a fiat transaction.
type FiatPaymentConfirmParams struct {
	ServiceName     string // The name of the service provider (e.g., "STRIPE").
	PaymentIntentID string
}

type FiatPaymentConfirmInfo struct {
	PaymentIntent *stripe.PaymentIntent `json:"payment_intent"`
	IsConfirmed   bool                  `json:"is_confirmed"`
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

// Pay attempts to pay the specified service using the provided parameters.
func (fiats Fiats) ConfirmPayment(params FiatPaymentConfirmParams) (*FiatPaymentConfirmInfo, error) {
	for _, f := range fiats {
		if params.ServiceName != f.Name {
			continue // Skip the service if it does not match the provided name.
		}
		switch f.Service {
		// TODO: add new confirm services here.
		default:
			// Default to Stripe if no specific service is added.
			return f.StripeConfirmPayment(params)
		}
	}
	return nil, fmt.Errorf("service %s could not found", params.ServiceName)
}

// StripePay handles a payment using the Stripe payment gateway.
func (f Fiat) StripePay(params FiatParams) (*FiatTransactionInfo, error) {
	// Set up the Stripe API key for authentication.
	// @FIXME: it may cause data race
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
		Confirm:       stripe.Bool(true),
		CaptureMethod: stripe.String(string(stripe.PaymentIntentCaptureMethodAutomatic)),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled:        stripe.Bool(true),
			AllowRedirects: stripe.String("never"), // Block redirect-based methods
		},
		PaymentMethodOptions: &stripe.PaymentIntentPaymentMethodOptionsParams{
			Card: &stripe.PaymentIntentPaymentMethodOptionsCardParams{
				RequestThreeDSecure: stripe.String("automatic"), // Let Stripe decide when 3DS is needed
			},
		},
		SetupFutureUsage: stripe.String(string(stripe.PaymentIntentSetupFutureUsageOffSession)),
	}

	// If there is a transfer, add related data to the payment intent.
	if params.Transfer != nil {
		intentParams.ConfirmationMethod = stripe.String(string(stripe.PaymentIntentConfirmationMethodAutomatic))
		intentParams.ReturnURL = stripe.String(f.Callback)
		intentParams.Confirm = stripe.Bool(true)
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
	log.Printf("payment intent: %v", result)
	// Create transaction info using the result from Stripe.
	info := &FiatTransactionInfo{
		TXID:        result.ID,
		TotalAmount: result.Amount,
		Date:        time.Now(),
		Currency:    string(result.Currency),
		Meta:        result,
	}

	if result.Status == stripe.PaymentIntentStatusRequiresAction {
		info.RequiresAction = true
		info.ClientSecret = result.ClientSecret
		return info, nil
	}

	if result.Status == stripe.PaymentIntentStatusRequiresConfirmation {
		confirmed, err := paymentintent.Confirm(result.ID, nil)
		if err != nil {
			return info, err
		}
		result = confirmed
	}

	if result.Status != stripe.PaymentIntentStatusSucceeded {
		return info, fmt.Errorf("Payment is not completed and is in the %s status", result.Status)
	}

	info.Confirmed = true
	return info, nil

	// // Confirm the payment intent using the selected payment method.
	// if _, err := paymentintent.Confirm(
	// 	result.ID,
	// 	&stripe.PaymentIntentConfirmParams{
	// 		PaymentMethod: stripe.String(method.ID),
	// 	},
	// ); err != nil {
	// 	return info, err // Return the transaction info and any errors during confirmation.
	// }
}

func (f Fiat) StripeConfirmPayment(params FiatPaymentConfirmParams) (*FiatPaymentConfirmInfo, error) {
	// Set up the Stripe API key for authentication.
	// @FIXME: it may cause data race
	stripe.Key = f.ApiKey

	intent, err := paymentintent.Get(params.PaymentIntentID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve payment intent: %v", err)
	}

	info := &FiatPaymentConfirmInfo{
		PaymentIntent: intent,
		IsConfirmed:   false,
	}

	if intent.Status == stripe.PaymentIntentStatusSucceeded {
		info.IsConfirmed = true
	}

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

func (f Fiat) AddCustomer(email string) (*stripe.Customer, error) {
	// @FIXME: it may cause data race
	stripe.Key = f.ApiKey

	c, err := customer.New(&stripe.CustomerParams{
		Email: stripe.String(email),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %v", err)
	}

	return c, nil
}

func (f Fiat) AttachPaymentMethod(customerID string, cardToken string) (*stripe.PaymentMethod, error) {
	// @FIXME: it may cause data race
	stripe.Key = f.ApiKey

	pm, err := paymentmethod.New(&stripe.PaymentMethodParams{
		Type: stripe.String("card"),
		Card: &stripe.PaymentMethodCardParams{
			Token: stripe.String(cardToken),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create payment method: %v", err)
	}
	// 3. Attach payment method to customer
	paymentmethod.Attach(pm.ID, &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	})

	_, err = customer.Update(customerID, &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(pm.ID),
		},
	})
	if err != nil {
		return pm, fmt.Errorf("attached payment method but failed to set as default: %w", err)
	}
	return pm, nil
}

func (f Fiat) FetchCards(customerID string) ([]*stripe.PaymentMethod, error) {
	// @FIXME: it may cause data race
	stripe.Key = f.ApiKey
	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(customerID),
		Type:     stripe.String("card"),
	}

	iter := paymentmethod.List(params)
	var cards []*stripe.PaymentMethod

	for iter.Next() {
		pm := iter.PaymentMethod()
		cards = append(cards, pm)
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return cards, nil
}

func (f Fiat) DeleteCard(paymentMethodID string) error {
	// @FIXME: it may cause data race
	stripe.Key = f.ApiKey

	if _, err := paymentmethod.Detach(paymentMethodID, nil); err != nil {
		return fmt.Errorf("failed to detach payment method: %v", err)
	}

	return nil
}

func (f Fiat) CreateAccount(country string) (*stripe.Account, error) {
	// @FIXME: it may cause data race
	stripe.Key = f.ApiKey

	acc, err := account.New(&stripe.AccountParams{
		Type:    stripe.String(string(stripe.AccountTypeExpress)),
		Country: stripe.String(country),
		Capabilities: &stripe.AccountCapabilitiesParams{
			CardPayments: &stripe.AccountCapabilitiesCardPaymentsParams{
				Requested: stripe.Bool(true),
			},
			Transfers: &stripe.AccountCapabilitiesTransfersParams{
				Requested: stripe.Bool(true),
			},
		},
		Settings: &stripe.AccountSettingsParams{
			Payouts: &stripe.AccountSettingsPayoutsParams{
				Schedule: &stripe.AccountSettingsPayoutsScheduleParams{
					Interval: stripe.String(string(stripe.AccountSettingsPayoutsScheduleIntervalManual)),
				},
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create account: %v", err)
	}

	return acc, nil
}

func (f Fiat) CreateAccountLink(account *stripe.Account, redirectURL string) (*stripe.AccountLink, error) {
	// @FIXME: it may cause data race
	stripe.Key = f.ApiKey

	accountLink, err := accountlink.New(&stripe.AccountLinkParams{
		Account:    stripe.String(account.ID),
		RefreshURL: stripe.String(redirectURL),
		ReturnURL:  stripe.String(redirectURL),
		Type:       stripe.String(string(stripe.AccountLinkTypeAccountOnboarding)),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create account link: %v", err)
	}

	return accountLink, nil
}

func (f Fiat) FetchAccount(accountID string) (*stripe.Account, error) {
	// @FIXME: it may cause data race
	stripe.Key = f.ApiKey

	acc, err := account.GetByID(accountID, nil)

	if err != nil {
		return nil, fmt.Errorf("failed to create account link: %v", err)
	}

	return acc, nil
}
