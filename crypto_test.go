package gopay_test

import (
	"fmt"
	"testing"

	"github.com/socious-io/gopay"
)

func TestGetTXInfo(t *testing.T) {
	token := gopay.CryptoToken{
		Address: "0x55d398326f99059ff775485246999027b3197955",
	}
	crypto := gopay.Chain{
		ContractAddress: "0x2Bdf475Bf5353cF52Aa4339A0FA308B4e1e22C3A",
		Explorer:        "https://api.bscscan.com/api",
		ApiKey:          "-----",
		Type:            gopay.EVM,
		Tokens:          []gopay.CryptoToken{token},
	}

	txHash := "0x11dd03d47456ef296f3cb4d6eea8e89f5975ae7cabd2e0b4f0f18d3e1d41914c"

	// Call the GetTXInfo function, which makes the real HTTP request
	info, err := crypto.GetTXInfo(txHash, token)
	fmt.Println(info, "----------------------------------@@")
	// Check for errors
	if err != nil {
		t.Errorf("Failed to get transaction info: %v", err)
	}
}
