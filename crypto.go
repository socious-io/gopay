package gopay

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Chains []Chain

type Chain struct {
	Name            string        `json:"name"`
	Explorer        string        `json:"explorer"`
	ContractAddress string        `json:"contract_address"`
	Tokens          []CryptoToken `json:"-"`
	Type            NetworkType   `json:"type"`
	Mode            NetworkMode   `json:"mode"`
	ApiKey          string        `json:"-"`
}

type CryptoToken struct {
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Address  string `json:"address"`
	Decimals int    `json:"decimals"`
}

type CryptoTransactionInfo struct {
	TxHash      string      `json:"txhash"`
	TotalAmount float64     `json:"total_amount"`
	To          string      `json:"to"`
	From        string      `json:"from"`
	Token       CryptoToken `json:"token"`
	Date        time.Time   `json:"date"`
	Valid       bool        `json:"valid"`
	Message     string      `json:"message"`
	Meta        interface{} `json:"meta"`
}

type EvmTokenTransferResponse struct {
	BlockNumber       string `json:"blockNumber"`
	TimeStamp         string `json:"timeStamp"`
	Hash              string `json:"hash"`
	Nonce             string `json:"nonce"`
	BlockHash         string `json:"blockHash"`
	From              string `json:"from"`
	ContractAddress   string `json:"contractAddress"`
	To                string `json:"to"`
	Value             string `json:"value"`
	TokenName         string `json:"tokenName"`
	TokenSymbol       string `json:"tokenSymbol"`
	TokenDecimal      string `json:"tokenDecimal"`
	TransactionIndex  string `json:"transactionIndex"`
	Gas               string `json:"gas"`
	GasPrice          string `json:"gasPrice"`
	GasUsed           string `json:"gasUsed"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
	Input             string `json:"input"`
	Confirmations     string `json:"confirmations"`
}

func (c Chain) GetTXInfo(txHash string) (*CryptoTransactionInfo, error) {
	switch c.Type {
	case EVM:
		return c.getEvmTXInfo(txHash)
	case CARDANO:
		return c.getCardanoTXInfo(txHash)
	default:
		return nil, fmt.Errorf("unknown crypto env")
	}
}

func (t CryptoTransactionInfo) ID() string {
	return t.TxHash
}

func (c Chain) getEvmTXInfo(txHash string) (*CryptoTransactionInfo, error) {
	url := fmt.Sprintf("%s?module=account&action=tokentx&address=%s&apikey=%s", c.Explorer, c.ContractAddress, c.ApiKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	var response struct {
		Status  string
		Message string
		Result  []EvmTokenTransferResponse
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %v", err)
	}

	var evmInfo *EvmTokenTransferResponse

	for _, res := range response.Result {
		if res.Hash == txHash {
			evmInfo = &res
		}
	}

	if evmInfo == nil {
		return nil, fmt.Errorf("transaction %s not found", txHash)
	}

	info := &CryptoTransactionInfo{
		TxHash:      txHash,
		TotalAmount: fromStrTokenValueToNumber(evmInfo.Value, evmInfo.TokenDecimal),
		Date:        fromStrTimestampToTime(evmInfo.TimeStamp),
		From:        evmInfo.From,
		To:          evmInfo.To,
		Meta:        evmInfo,
	}

	for _, token := range c.Tokens {
		if strings.EqualFold(evmInfo.ContractAddress, token.Address) {
			info.Token = token
			info.Valid = true
			return info, nil
		}
	}

	info.Message = fmt.Sprintf("token <%s | %s> not match to configured tokens", evmInfo.ContractAddress, evmInfo.TokenName)
	return info, nil
}

func (Chain) getCardanoTXInfo(_ string) (*CryptoTransactionInfo, error) {
	return nil, fmt.Errorf("cardano transactions not implemented")
}

func (chains Chains) CryptoCryptoTransactionInfo(txHash, dest string) (*CryptoTransactionInfo, error) {

	for _, c := range chains {
		if strings.EqualFold(c.ContractAddress, dest) {
			return c.GetTXInfo(txHash)
		}
	}

	return nil, fmt.Errorf("contract address %s not found", dest)
}
