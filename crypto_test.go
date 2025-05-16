package gopay_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/socious-io/gopay"
)

// Mock response body for testing
func mockEtherscanResponseBody() *mockReadCloser {
	response := `{
		"status": "1",
		"message": "OK",
		"result": [{
			"hash": "0xTransactionHash",
			"from": "0xFromAddress",
			"to": "0xToAddress",
			"value": "1000000000000000000", 
			"tokenName": "ETH",
			"tokenSymbol": "ETH",
			"tokenDecimal": "18",
			"confirmations": "12"
		}]
	}`
	return &mockReadCloser{[]byte(response)}
}

// Test GetEvmTXInfo method
func TestGetTXInfo(t *testing.T) {
	// Setup
	chain := gopay.Chain{
		Name:            "Ethereum",
		Explorer:        "https://api.etherscan.io/api",
		ContractAddress: "0xYourContractAddress",
		ApiKey:          "YourAPIKey",
		Type:            gopay.EVM,
		Mode:            "mainnet",
	}

	txHash := "0xTransactionHash"
	token := gopay.CryptoToken{
		Name:     "Ether",
		Symbol:   "ETH",
		Address:  "0xTokenAddress",
		Decimals: 18,
	}

	// Create a mock response
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       mockEtherscanResponseBody(),
	}

	// Set up mock HTTP client
	mockClient := &MockHTTPClient{
		Response: mockResponse,
	}

	// Use the mock client in the test
	originalHTTPClient := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: mockClient}
	defer func() { http.DefaultClient = originalHTTPClient }() // Restore the original client after the test

	// Call GetTXInfo
	result, err := chain.GetTXInfo(txHash, token)

	// Validate results
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if result == nil {
		t.Errorf("Expected result, but got nil")
		return
	}
	if result.TxHash != txHash {
		t.Errorf("Expected txHash %s, but got %s", txHash, result.TxHash)
	}
	if result.TotalAmount != 1.0 {
		t.Errorf("Expected TotalAmount 1.0, but got %f", result.TotalAmount)
	}
	if result.From != "0xFromAddress" {
		t.Errorf("Expected From address 0xFromAddress, but got %s", result.From)
	}
	if result.To != "0xToAddress" {
		t.Errorf("Expected To address 0xToAddress, but got %s", result.To)
	}
	t.Log("crypto tests successfully done")
}

func TestCardanoTXInfo(t *testing.T) {
	// Setup
	chain := gopay.Chain{
		Name:            "Cardano",
		Explorer:        "https://cardano-mainnet.blockfrost.io/api/v0",
		ContractAddress: "",
		ApiKey:          "",
		Type:            gopay.CARDANO,
		Mode:            "mainnet",
	}

	txHash := "f1e8498b55c3a9689bafda15243d26514ed0730e4ea775f1655a4f04f84ccf1a"
	token := gopay.CryptoToken{
		Name:     "",
		Symbol:   "",
		Address:  "",
		Decimals: 6,
	}
	// Call GetTXInfo
	result, err := chain.GetTXInfo(txHash, token)

	if err != nil {
		t.Error(err)
	}
	b, _ := json.Marshal(result)
	fmt.Println(string(b), "----------------------")
}
