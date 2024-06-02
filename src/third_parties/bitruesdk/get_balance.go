package bitruesdk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"

	"github.com/shopspring/decimal"
)

type BalanceGetBalance struct {
	Asset  string          `json:"asset"`
	Free   decimal.Decimal `json:"free"`
	Locked string          `json:"locked"`
}

// AccountInfo represents the account information
type ResponseGetBalance struct {
	MakerCommission  int                 `json:"makerCommission"`
	TakerCommission  int                 `json:"takerCommission"`
	BuyerCommission  int                 `json:"buyerCommission"`
	SellerCommission int                 `json:"sellerCommission"`
	UpdateTime       *int64              `json:"updateTime"`
	Balances         []BalanceGetBalance `json:"balances"`
	CanTrade         bool                `json:"canTrade"`
	CanWithdraw      bool                `json:"canWithdraw"`
	CanDeposit       bool                `json:"canDeposit"`
}

func GetBalance(apiKey, secretKey string) (*ResponseGetBalance, error) {
	url := "https://www.bitrue.com/api/v1/account"

	timestamp := time.Now().UnixMilli()
	mac := hmac.New(sha256.New, []byte(secretKey))
	data := "timestamp=" + strconv.FormatInt(timestamp, 10)
	mac.Write([]byte(data))
	signature := mac.Sum(nil)
	signatureStr := hex.EncodeToString(signature)

	req, err := http.NewRequest("GET", url+"?"+data+"&signature="+signatureStr, nil)
	client := http.Client{}
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-MBX-APIKEY", apiKey)
	req.Header.Add("X-MBX-SIGNATURE", signatureStr)
	req.Header.Add("X-MBX-TIMESTAMP", strconv.FormatInt(timestamp, 10))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(body))

	res := ResponseGetBalance{}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
