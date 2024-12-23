# GOPay

GoPay is a flexible and modular payment processing system built using Go. It supports both fiat and cryptocurrency transactions. It provides an easy-to-use interface for handling payments, managing identities, and handling transaction statuses.

## Features

- **Payment Creation**: Allows for the creation of payments with details such as description, currency, amount, and type (Fiat or Crypto).
- **Transaction Management**: Supports the creation, verification, and cancellation of transactions.
- **Fiat Support**: Supports fiat currency payments with fiat service integration.
- **Crypto Support**: Supports cryptocurrency payments with crypto service integration.
- **Identity Management**: Links payments to identities, allocating amounts to specific roles or accounts.
- **Transaction Fees and Discounts**: Handles transaction fees and discounts, applying them to payment amounts.
- **Payment Modes**: Allows switching between fiat and crypto payment modes.

## Components

- **Payment**: Represents a payment transaction and its associated details, including identities, amount, and status.
- **Transaction**: Represents individual transactions tied to a payment, including details like transaction ID, fees, and status.
- **PaymentIdentity**: Represents an identity linked to a payment, with allocated amounts and roles.
- **TransactionType**: Enum representing different types of transactions (e.g., Deposit).
- **PaymentStatus**: Enum representing different statuses of a payment (e.g., Initiated, Confirmed).
- **PaymentType**: Enum representing the type of payment (e.g., Fiat, Crypto).

### Advantages of this Design:
1. **Seamless Audit Trail**: 
* `Transaction` captures real-money movement, with clear links to `Payment` for context. 
2. **Extensibility**: 
* Future features like fee structures, batching, or conditions are easily supported via `Transaction Meta`
3. **Flexibility in Roles**:
* Use PaymentIdentity.RoleName to model custom relationships (e.g., splitting fees, platform intermediaries).
4. **Custom Workflow**:
*Supports manual and programmatic payment workflows through linked transactions.`

## Setup and Installation

To get started with GoPay, follow these steps:

1. Clone the repository:
   ```bash
   git clone https://github.com//gopay.git
   ```

2. Navigate to the project directory:
   ```bash
   cd gopay
   ```

3. Install dependencies:
   ```bash
   go mod tidy
   ```

4. Set up your database connection in the `config` object (replace with your own credentials).

5. Run your Go application:
   ```bash
   go run main.go
   ```

## Usage

### Creating a Payment

```go
paymentParams := gopay.PaymentParams{
    Tag: "Payment #1",
    Description: "Test payment",
    Ref: "12345",
    Currency: gopay.USD,
    Meta: map[string]interface{}{"note": "This is a test payment"},
}

payment, err := gopay.New(paymentParams)
if err != nil {
    log.Fatal("Error creating payment: ", err)
}
```

### Adding an Identity to a Payment

```go
identityParams := gopay.IdentityParams{
    ID:       identityID,
    RoleName: "customer",
    Account:  "account1",
    Amount:   100.0,
    Meta:     nil,
}

identity, err := payment.AddIdentity(identityParams)
if err != nil {
    log.Fatal("Error adding identity: ", err)
}
```

### Verifying a Crypto Deposit

```go
err := payment.ConfirmDeposit("tx_hash_here")
if err != nil {
    log.Fatal("Error confirming deposit: ", err)
}
```

### Handling Fiat Payments

```go
err := payment.Deposit()
if err != nil {
    log.Fatal("Error processing deposit: ", err)
}
```


## License

GoPay is licensed under the GPL3 License. See the [LICENSE](https://github.com/socious-io/gopay/blob/main/LICENSE) file for more information.