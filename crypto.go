// Package gopay provides functions to interact with multiple cryptocurrency networks and retrieve transaction information.
package gopay

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/blockfrost/blockfrost-go"
)

// Chains represents a slice of Chain objects. Each Chain can represent a different blockchain network.
type Chains []Chain

// Chain represents a blockchain network, such as Ethereum (EVM) or Cardano. It includes network details like its name, explorer URL,
// contract address, associated tokens, type, and network mode.
type Chain struct {
	Name            string        `json:"name" mapstructure:"name"`                        // Name of the blockchain network
	Explorer        string        `json:"explorer" mapstructure:"explorer"`                // URL of the block explorer for the network
	ContractAddress string        `json:"contract_address" mapstructure:"contractaddress"` // Address of the contract associated with the network
	Tokens          []CryptoToken `json:"-" mapstructure:"tokens"`                         // List of tokens associated with the blockchain, hidden in JSON output
	Type            NetworkType   `json:"type" mapstructure:"type"`                        // Type of blockchain (e.g., EVM, Cardano)
	Mode            NetworkMode   `json:"mode" mapstructure:"mode"`                        // Network operation mode (e.g., mainnet, testnet)
	ApiKey          string        `json:"-" mapstructure:"apikey"`                         // API key for interacting with the blockchain explorer, hidden in JSON output
}

// CryptoToken represents a specific token on a blockchain. It includes the token's name, symbol, address, and the number of decimals it uses.
type CryptoToken struct {
	Name     string `json:"name" mapstructure:"name"`         // Name of the token (e.g., "Ethereum")
	Symbol   string `json:"symbol" mapstructure:"symbol"`     // Symbol of the token (e.g., "ETH")
	Address  string `json:"address" mapstructure:"address"`   // Blockchain address associated with the token
	Decimals int    `json:"decimals" mapstructure:"decimals"` // Number of decimal places the token supports
}

// CryptoTransactionInfo contains details about a transaction on the blockchain, such as transaction hash, amount,
// sender and recipient addresses, token details, confirmation status, and date.
type CryptoTransactionInfo struct {
	TxHash      string      `json:"txhash"`       // Transaction hash (unique identifier for the transaction)
	TotalAmount float64     `json:"total_amount"` // Total amount of tokens transferred in the transaction
	To          string      `json:"to"`           // Address of the recipient
	From        string      `json:"from"`         // Address of the sender
	Token       CryptoToken `json:"token"`        // Token associated with the transaction
	Date        time.Time   `json:"date"`         // Date and time of the transaction
	Confirmed   bool        `json:"confirmed"`    // Confirmation status of the transaction (e.g., confirmed or not)
	Message     string      `json:"message"`      // Optional message associated with the transaction
	Meta        interface{} `json:"meta"`         // Additional metadata associated with the transaction
}

// EvmTokenTransferResponse is the structure of the response received from an EVM-compatible blockchain explorer API.
// It contains details about a specific token transfer transaction.
type EvmTokenTransferResponse struct {
	BlockNumber       string `json:"blockNumber"`       // Block number where the transaction is included
	TimeStamp         string `json:"timeStamp"`         // Timestamp of the transaction
	Hash              string `json:"hash"`              // Unique hash for the transaction
	Nonce             string `json:"nonce"`             // Transaction nonce
	BlockHash         string `json:"blockHash"`         // Block hash in which the transaction is recorded
	From              string `json:"from"`              // Address of the sender
	ContractAddress   string `json:"contractAddress"`   // Contract address associated with the token transfer
	To                string `json:"to"`                // Address of the recipient
	Value             string `json:"value"`             // Amount transferred (in token's smallest unit)
	TokenName         string `json:"tokenName"`         // Token name (e.g., "ETH")
	TokenSymbol       string `json:"tokenSymbol"`       // Token symbol (e.g., "ETH")
	TokenDecimal      string `json:"tokenDecimal"`      // Token decimal precision
	TransactionIndex  string `json:"transactionIndex"`  // Index of the transaction in the block
	Gas               string `json:"gas"`               // Gas used for the transaction
	GasPrice          string `json:"gasPrice"`          // Gas price for the transaction
	GasUsed           string `json:"gasUsed"`           // Total gas used in the transaction
	CumulativeGasUsed string `json:"cumulativeGasUsed"` // Total gas used in the block up to the transaction
	Input             string `json:"input"`             // Input data (for contract calls)
	Confirmations     string `json:"confirmations"`     // Number of confirmations the transaction has received
}

type CardanoTokenTransferResponse struct {
	Info  blockfrost.TransactionContent `json:"info"`
	Utxos blockfrost.TransactionUTXOs   `json:"utxos"`
	Block blockfrost.Block              `json:"block"`
}

// CryptoParams holds parameters used to retrieve transaction information, such as the transaction hash and token address.
type CryptoParams struct {
	TxHash       string // The transaction hash (ID) for the blockchain transaction.
	TokenAddress string // The address of the token associated with the transaction.
}

// GetTXInfo retrieves the transaction information based on the transaction hash and token. It identifies the appropriate blockchain
// (EVM or Cardano) based on the chain configuration and calls the corresponding method to retrieve transaction details.
func (c Chain) GetTXInfo(txHash string, token CryptoToken) (*CryptoTransactionInfo, error) {
	switch c.Type {
	case EVM:
		return c.getEvmTXInfo(txHash, token)
	case CARDANO:
		return c.getCardanoTXInfo(txHash, token)
	default:
		return nil, fmt.Errorf("unknown crypto env")
	}
}

// ID returns the transaction hash as a string identifier for the CryptoTransactionInfo.
func (t CryptoTransactionInfo) ID() string {
	return t.TxHash
}

// getEvmTXInfo retrieves detailed transaction information from an Ethereum-like blockchain (EVM) using a block explorer API.
func (c Chain) getEvmTXInfo(txHash string, token CryptoToken) (*CryptoTransactionInfo, error) {

	var (
		maxRetries = 20          // Maximum number of retries
		retryDelay = time.Second // Delay between retries
		evmInfo    *EvmTokenTransferResponse
		resp       *http.Response
		err        error
		response   struct {
			Status  string
			Message string
			Result  []EvmTokenTransferResponse
		}
	)

	for retry := 0; retry < maxRetries; retry++ {
		url := fmt.Sprintf("%s?module=account&action=tokentx&address=%s&apikey=%s", c.Explorer, c.ContractAddress, c.ApiKey)
		resp, err = http.Get(url)
		if err != nil {
			fmt.Printf("Attempt %d: Error making HTTP request: %v\n", retry+1, err)
			time.Sleep(retryDelay)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("attempt %d: unexpected HTTP status: %s", retry+1, resp.Status)
			fmt.Printf("Attempt %d: Unexpected HTTP status: %s\n", retry+1, resp.Status)
			time.Sleep(retryDelay)
			continue
		}

		if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
			fmt.Printf("Attempt %d: Error decoding JSON: %v\n", retry+1, err)
			time.Sleep(retryDelay)
			continue
		}

		for _, res := range response.Result {
			if res.Hash == txHash {
				evmInfo = &res
			}
		}

		if evmInfo == nil {
			fmt.Printf("Attempt %d: transaction %s not found\n", retry+1, txHash)
			time.Sleep(retryDelay)
			continue
		}

		// If all data is fetched successfully, break the loop
		break
	}

	// If all retries fail, return the last error
	if err != nil || evmInfo == nil {
		return nil, fmt.Errorf("failed after %d retries: %v", maxRetries, err)
	}

	confirms, _ := strconv.Atoi(evmInfo.Confirmations)
	// Redo if blocks confirms are less that 10 blocks
	if confirms < 10 {
		time.Sleep(time.Second)
		return c.getEvmTXInfo(txHash, token)
	}

	return &CryptoTransactionInfo{
		TxHash:      txHash,
		TotalAmount: fromStrTokenValueToNumber(evmInfo.Value, evmInfo.TokenDecimal),
		Date:        fromStrTimestampToTime(evmInfo.TimeStamp),
		From:        evmInfo.From,
		To:          evmInfo.To,
		Meta:        evmInfo,
		Token:       token,
		Confirmed:   confirms >= 10,
	}, nil
}

// getCardanoTXInfo is a function for retrieving Cardano transaction information
func (c Chain) getCardanoTXInfo(txHash string, token CryptoToken) (*CryptoTransactionInfo, error) {
	api := blockfrost.NewAPIClient(
		blockfrost.APIClientOptions{
			Server:    c.Explorer,
			ProjectID: c.ApiKey,
		},
	)

	ctx := context.Background()
	maxRetries := 20          // Maximum number of retries
	retryDelay := time.Second // Delay between retries

	var (
		tx    blockfrost.TransactionContent
		utxos blockfrost.TransactionUTXOs
		block blockfrost.Block
		err   error
	)

	// Retry loop
	for retry := 0; retry < maxRetries; retry++ {
		// Fetch transaction details
		tx, err = api.Transaction(ctx, txHash)
		if err != nil {
			fmt.Printf("Attempt %d: Error fetching transaction: %v\n", retry+1, err)
			time.Sleep(retryDelay)
			continue
		}

		// Fetch transaction UTXOs
		utxos, err = api.TransactionUTXOs(ctx, txHash)
		if err != nil {
			fmt.Printf("Attempt %d: Error fetching transaction UTXOs: %v\n", retry+1, err)
			time.Sleep(retryDelay)
			continue
		}

		// Fetch block details
		block, err = api.Block(ctx, tx.Block)
		if err != nil {
			fmt.Printf("Attempt %d: Error fetching block: %v\n", retry+1, err)
			time.Sleep(retryDelay)
			continue
		}

		// If all data is fetched successfully, break the loop
		break
	}

	// If all retries fail, return the last error
	if err != nil {
		return nil, fmt.Errorf("failed after %d retries: %v", maxRetries, err)
	}
	var total float64

	for _, am := range utxos.Outputs[0].Amount {
		if matchAddress(token.Address, am.Unit) {
			total = fromStrTokenValueToNumber(am.Quantity, fmt.Sprintf("%d", token.Decimals))
		}
	}

	return &CryptoTransactionInfo{
		TxHash:      txHash,
		TotalAmount: total,
		Date:        time.Unix(int64(block.Time), 0),
		From:        utxos.Inputs[0].Address,
		To:          utxos.Outputs[0].Address,
		Meta:        CardanoTokenTransferResponse{tx, utxos, block},
		Token:       token,
		Confirmed:   true,
	}, nil
}

// TransactionInfo searches for a specific token and transaction hash, retrieves the appropriate chain, and returns transaction details.
func (chains Chains) TransactionInfo(params CryptoParams) (*CryptoTransactionInfo, error) {
	for _, c := range chains {
		for _, t := range c.Tokens {
			if strings.EqualFold(t.Address, params.TokenAddress) {
				return c.GetTXInfo(params.TxHash, t)
			}
		}
	}

	return nil, fmt.Errorf("token address %s not found", params.TokenAddress)
}
