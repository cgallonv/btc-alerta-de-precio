package bitcoin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

type BinanceAccountClient struct {
	httpClient *resty.Client
	apiKey     string
	apiSecret  string
}

type AccountBalance struct {
	TotalBalance     float64
	AvailableBalance float64
	Assets           []AssetBalance
	LastUpdated      time.Time
	// Additional fields from Binance API
	MakerCommission  int `json:"makerCommission"`
	TakerCommission  int `json:"takerCommission"`
	BuyerCommission  int `json:"buyerCommission"`
	SellerCommission int `json:"sellerCommission"`
	CommissionRates  struct {
		Maker  string `json:"maker"`
		Taker  string `json:"taker"`
		Buyer  string `json:"buyer"`
		Seller string `json:"seller"`
	} `json:"commissionRates"`
	CanTrade                   bool     `json:"canTrade"`
	CanWithdraw                bool     `json:"canWithdraw"`
	CanDeposit                 bool     `json:"canDeposit"`
	AccountType                string   `json:"accountType"`
	Permissions                []string `json:"permissions"`
	RequireSelfTradePrevention bool     `json:"requireSelfTradePrevention"`
	PreventSor                 bool     `json:"preventSor"`
	UpdateTime                 int64    `json:"updateTime"`
}

type AssetBalance struct {
	Symbol    string  `json:"symbol"`
	ValueUSD  float64 `json:"value_usd"`
	Change24h float64 `json:"change_24h"`
	Free      string  `json:"free"`   // Original amount available
	Locked    string  `json:"locked"` // Amount locked in orders
	Total     float64 `json:"total"`  // Calculated total (free + locked)
}

func NewBinanceAccountClient(apiKey, apiSecret string) *BinanceAccountClient {
	client := resty.New()
	client.SetTimeout(10 * time.Second)
	client.SetHeader("X-MBX-APIKEY", apiKey)
	client.SetBaseURL("https://api.binance.com")

	return &BinanceAccountClient{
		httpClient: client,
		apiKey:     apiKey,
		apiSecret:  apiSecret,
	}
}

func (c *BinanceAccountClient) GetAccountBalance() (*AccountBalance, error) {
	timestamp := time.Now().UnixMilli()
	queryString := fmt.Sprintf("timestamp=%d", timestamp)
	signature := c.generateSignature(queryString)

	log.Printf("Making request to Binance API with timestamp: %d", timestamp)
	log.Printf("Query string: %s", queryString)
	log.Printf("Signature length: %d", len(signature))

	var response struct {
		MakerCommission  int `json:"makerCommission"`
		TakerCommission  int `json:"takerCommission"`
		BuyerCommission  int `json:"buyerCommission"`
		SellerCommission int `json:"sellerCommission"`
		CommissionRates  struct {
			Maker  string `json:"maker"`
			Taker  string `json:"taker"`
			Buyer  string `json:"buyer"`
			Seller string `json:"seller"`
		} `json:"commissionRates"`
		CanTrade                   bool   `json:"canTrade"`
		CanWithdraw                bool   `json:"canWithdraw"`
		CanDeposit                 bool   `json:"canDeposit"`
		RequireSelfTradePrevention bool   `json:"requireSelfTradePrevention"`
		PreventSor                 bool   `json:"preventSor"`
		UpdateTime                 int64  `json:"updateTime"`
		AccountType                string `json:"accountType"`
		Balances                   []struct {
			Asset  string `json:"asset"`
			Free   string `json:"free"`
			Locked string `json:"locked"`
		} `json:"balances"`
		Permissions []string `json:"permissions"`
	}

	resp, err := c.httpClient.R().
		SetQueryParams(map[string]string{
			"timestamp": strconv.FormatInt(timestamp, 10),
			"signature": signature,
		}).
		SetResult(&response).
		Get("/api/v3/account")

	if err != nil {
		return nil, fmt.Errorf("error fetching account info: %w", err)
	}

	if resp.StatusCode() != 200 {
		log.Printf("Binance API error response: %s", resp.String())
		return nil, fmt.Errorf("Binance API returned status %d: %s", resp.StatusCode(), resp.String())
	}

	accountBalance := &AccountBalance{
		LastUpdated:                time.Now(),
		MakerCommission:            response.MakerCommission,
		TakerCommission:            response.TakerCommission,
		BuyerCommission:            response.BuyerCommission,
		SellerCommission:           response.SellerCommission,
		CommissionRates:            response.CommissionRates,
		CanTrade:                   response.CanTrade,
		CanWithdraw:                response.CanWithdraw,
		CanDeposit:                 response.CanDeposit,
		RequireSelfTradePrevention: response.RequireSelfTradePrevention,
		PreventSor:                 response.PreventSor,
		UpdateTime:                 response.UpdateTime,
		AccountType:                response.AccountType,
		Permissions:                response.Permissions,
	}

	for _, balance := range response.Balances {
		free, _ := strconv.ParseFloat(balance.Free, 64)
		locked, _ := strconv.ParseFloat(balance.Locked, 64)
		total := free + locked

		if total > 0 {
			// Get current price for the asset
			price, err := c.getAssetPrice(balance.Asset + "USDT")
			if err != nil {
				price = 0
			}

			valueUSD := total * price
			accountBalance.TotalBalance += valueUSD
			accountBalance.AvailableBalance += free * price

			// Get 24h change
			change24h, err := c.get24hChange(balance.Asset + "USDT")
			if err != nil {
				change24h = 0
			}

			accountBalance.Assets = append(accountBalance.Assets, AssetBalance{
				Symbol:    balance.Asset,
				Free:      balance.Free,
				Locked:    balance.Locked,
				Total:     total,
				ValueUSD:  valueUSD,
				Change24h: change24h,
			})
		}
	}

	return accountBalance, nil
}

func (c *BinanceAccountClient) generateSignature(queryString string) string {
	h := hmac.New(sha256.New, []byte(c.apiSecret))
	h.Write([]byte(queryString))
	return hex.EncodeToString(h.Sum(nil))
}

func (c *BinanceAccountClient) getAssetPrice(symbol string) (float64, error) {
	var response struct {
		Price string `json:"price"`
	}

	resp, err := c.httpClient.R().
		SetQueryParam("symbol", symbol).
		SetResult(&response).
		Get("/api/v3/ticker/price")

	if err != nil {
		return 0, err
	}

	if resp.StatusCode() != 200 {
		return 0, fmt.Errorf("API returned status %d", resp.StatusCode())
	}

	return strconv.ParseFloat(response.Price, 64)
}

func (c *BinanceAccountClient) get24hChange(symbol string) (float64, error) {
	var response struct {
		PriceChangePercent string `json:"priceChangePercent"`
	}

	resp, err := c.httpClient.R().
		SetQueryParam("symbol", symbol).
		SetResult(&response).
		Get("/api/v3/ticker/24hr")

	if err != nil {
		return 0, err
	}

	if resp.StatusCode() != 200 {
		return 0, fmt.Errorf("API returned status %d", resp.StatusCode())
	}

	return strconv.ParseFloat(response.PriceChangePercent, 64)
}
