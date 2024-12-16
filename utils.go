package gopay

import (
	"math/big"
	"strconv"
	"strings"
	"time"
)

func fromStrTimestampToTime(valueStr string) time.Time {
	timestampInt, _ := strconv.ParseInt(valueStr, 10, 64)
	return time.Unix(timestampInt, 0)

}

func fromStrTokenValueToNumber(valueStr string, tokenDecimal string) float64 {
	tokenDecimal = strings.TrimSpace(tokenDecimal)

	// Convert the value to big.Int
	value := new(big.Int)
	_, success := value.SetString(valueStr, 10)
	if !success {
		return -1
	}

	// Convert tokenDecimal to an integer
	decimal := new(big.Int)
	_, success = decimal.SetString(tokenDecimal, 10)
	if !success {
		return -1
	}

	// Compute the factor (10^decimal)
	power := new(big.Int).Exp(big.NewInt(10), decimal, nil)

	// Perform the division and convert the result to a big.Float
	result := new(big.Float).Quo(new(big.Float).SetInt(value), new(big.Float).SetInt(power))

	// Convert to float64 and return the result
	floatResult, _ := result.Float64()
	return floatResult
}
