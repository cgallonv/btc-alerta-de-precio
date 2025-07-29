// Package bitcoin provides functionality for interacting with cryptocurrency APIs.
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

// BinanceClient handles all Binance API operations including account information,
// price data, and trading functionality.
//
// Example usage:
//
//	client := NewBinanceClient(apiKey, apiSecret)
//	balance, err := client.GetAccountBalance()
//	if err != nil {
//	    log.Printf("Error: %v", err)
//	    return
//	}
//	fmt.Printf("Total Balance: $%.2f\n", balance.TotalBalance)
type BinanceClient struct {
	httpClient *resty.Client
	apiKey     string
	apiSecret  string
}

// AccountBalance represents account balance information from Binance API.
// It includes total balance, available balance, individual asset balances,
// and account status information.
//
// Example usage:
//
//	balance, err := client.GetAccountBalance()
//	if err != nil {
//	    return err
//	}
//	for _, asset := range balance.Assets {
//	    fmt.Printf("%s: %.8f ($%.2f)\n", asset.Symbol, asset.Total, asset.ValueUSD)
//	}
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

// AssetBalance represents balance information for a single asset.
// It includes both the original string values from Binance API and
// calculated numeric values for easier processing.
//
// Example usage:
//
//	asset := AssetBalance{
//	    Symbol: "BTC",
//	    Free: "0.12345678",
//	    Locked: "0.00000000",
//	    Total: 0.12345678,
//	    ValueUSD: 5000.00,
//	    Change24h: -2.5,
//	}
//	fmt.Printf("BTC Balance: %s (%.2f%%)\n", asset.Free, asset.Change24h)
type AssetBalance struct {
	Symbol    string  `json:"symbol"`
	ValueUSD  float64 `json:"value_usd"`
	Change24h float64 `json:"change_24h"`
	Free      string  `json:"free"`   // Original amount available
	Locked    string  `json:"locked"` // Amount locked in orders
	Total     float64 `json:"total"`  // Calculated total (free + locked)
}

// PriceData represents current price information for an asset.
// It includes the current price, 24-hour price change percentage,
// and metadata about the price source.
//
// Example usage:
//
//	price, err := client.GetCurrentPrice()
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("BTC: $%.2f (%+.2f%%)\n", price.Price, price.PriceChangePercent)
type PriceData struct {
	Price              float64   `json:"price"`
	PriceChangePercent float64   `json:"price_change_percent"`
	Currency           string    `json:"currency"`
	Timestamp          time.Time `json:"timestamp"`
	Source             string    `json:"source"`
}

// NewBinanceClient creates a new Binance API client with the provided API credentials.
// The client handles authentication and provides methods for accessing various Binance API endpoints.
//
// Example usage:
//
//	client := NewBinanceClient("your-api-key", "your-api-secret")
//	price, err := client.GetCurrentPrice()
//	if err != nil {
//	    log.Printf("Error: %v", err)
//	    return
//	}
func NewBinanceClient(apiKey, apiSecret string) *BinanceClient {
	client := resty.New()
	client.SetTimeout(10 * time.Second)
	client.SetHeader("X-MBX-APIKEY", apiKey)
	client.SetBaseURL("https://api.binance.com")

	return &BinanceClient{
		httpClient: client,
		apiKey:     apiKey,
		apiSecret:  apiSecret,
	}
}

// GetAccountBalance fetches account balance information from Binance API.
// It returns detailed information about the account including total balance,
// individual asset balances, and account status.
//
// Example usage:
//
//	balance, err := client.GetAccountBalance()
//	if err != nil {
//	    if binanceErr, ok := err.(*BinanceError); ok {
//	        log.Printf("Binance API error %d: %s", binanceErr.Code, binanceErr.Message)
//	    }
//	    return nil, err
//	}
//	fmt.Printf("Total Balance: $%.2f\n", balance.TotalBalance)
func (c *BinanceClient) GetAccountBalance() (*AccountBalance, error) {
	timestamp := time.Now().UnixMilli()
	queryString := fmt.Sprintf("timestamp=%d", timestamp)
	signature := c.generateSignature(queryString)

	log.Printf("🔄 Fetching account balance from Binance API")

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
		SetQueryParam("timestamp", fmt.Sprintf("%d", timestamp)).
		SetQueryParam("signature", signature).
		SetResult(&response).
		Get("/api/v3/account")

	if err != nil {
		log.Printf("❌ Error fetching account balance: %v", err)
		return nil, fmt.Errorf("error fetching account balance: %w", err)
	}

	if resp.StatusCode() != 200 {
		binanceErr := NewBinanceError(resp.StatusCode(), resp.String())
		log.Printf("❌ Binance API error: %v", binanceErr)
		return nil, binanceErr
	}

	log.Printf("✅ Account balance fetched successfully")

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
		free, err := strconv.ParseFloat(balance.Free, 64)
		if err != nil {
			log.Printf("⚠️ Error parsing free balance for %s: %v", balance.Asset, err)
			continue
		}

		locked, err := strconv.ParseFloat(balance.Locked, 64)
		if err != nil {
			log.Printf("⚠️ Error parsing locked balance for %s: %v", balance.Asset, err)
			continue
		}

		total := free + locked

		if total > 0 {
			// Get current price for the asset
			price, err := c.GetAssetPrice(balance.Asset + "USDT")
			if err != nil {
				log.Printf("⚠️ Error getting price for %s: %v", balance.Asset, err)
				price = 0
			}

			valueUSD := total * price
			accountBalance.TotalBalance += valueUSD
			accountBalance.AvailableBalance += free * price

			// Get 24h change
			change24h, err := c.Get24hChange(balance.Asset + "USDT")
			if err != nil {
				log.Printf("⚠️ Error getting 24h change for %s: %v", balance.Asset, err)
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

			log.Printf("📊 Asset %s: Free: %s, Locked: %s, Total: %.8f, Value: $%.2f, Change: %.2f%%",
				balance.Asset, balance.Free, balance.Locked, total, valueUSD, change24h)
		}
	}

	log.Printf("💰 Total balance: $%.2f, Available: $%.2f",
		accountBalance.TotalBalance, accountBalance.AvailableBalance)

	return accountBalance, nil
}

// GetCurrentPrice fetches current BTC price from Binance API.
// It returns detailed price information including the current price,
// 24-hour price change percentage, and metadata.
//
// Example usage:
//
//	price, err := client.GetCurrentPrice()
//	if err != nil {
//	    if binanceErr, ok := err.(*BinanceError); ok {
//	        log.Printf("Binance API error %d: %s", binanceErr.Code, binanceErr.Message)
//	    }
//	    return nil, err
//	}
//	fmt.Printf("BTC: $%.2f (%+.2f%%)\n", price.Price, price.PriceChangePercent)
func (c *BinanceClient) GetCurrentPrice() (*PriceData, error) {
	log.Printf("🔄 Fetching BTC price from Binance API")

	var response struct {
		Symbol             string `json:"symbol"`
		PriceChange        string `json:"priceChange"`
		PriceChangePercent string `json:"priceChangePercent"`
		LastPrice          string `json:"lastPrice"`
	}

	resp, err := c.httpClient.R().
		SetResult(&response).
		Get("/api/v3/ticker/24hr?symbol=BTCUSDT")

	if err != nil {
		log.Printf("❌ Error fetching price: %v", err)
		return nil, fmt.Errorf("error fetching price from Binance: %w", err)
	}

	if resp.StatusCode() != 200 {
		binanceErr := NewBinanceError(resp.StatusCode(), resp.String())
		log.Printf("❌ Binance API error: %v", binanceErr)
		return nil, binanceErr
	}

	price, err := strconv.ParseFloat(response.LastPrice, 64)
	if err != nil {
		log.Printf("❌ Error parsing price: %v", err)
		return nil, fmt.Errorf("error parsing price from Binance: %w", err)
	}

	priceChangePercent, err := strconv.ParseFloat(response.PriceChangePercent, 64)
	if err != nil {
		log.Printf("❌ Error parsing price change: %v", err)
		return nil, fmt.Errorf("error parsing price change percent from Binance: %w", err)
	}

	log.Printf("✅ BTC price fetched successfully: $%.2f (%+.2f%%)", price, priceChangePercent)

	return &PriceData{
		Price:              price,
		PriceChangePercent: priceChangePercent,
		Currency:           "USD",
		Timestamp:          time.Now(),
		Source:             "Binance",
	}, nil
}

// GetAssetPrice fetches current price for a symbol from Binance API.
// For USDT, it returns a fixed 1:1 USD value since USDT is a stablecoin.
//
// Example usage:
//
//	price, err := client.GetAssetPrice("BTCUSDT")
//	if err != nil {
//	    if binanceErr, ok := err.(*BinanceError); ok {
//	        log.Printf("Binance API error %d: %s", binanceErr.Code, binanceErr.Message)
//	    }
//	    return 0, err
//	}
//	fmt.Printf("BTC Price: $%.2f\n", price)
func (c *BinanceClient) GetAssetPrice(symbol string) (float64, error) {
	log.Printf("🔄 Fetching price for %s", symbol)

	// Special case for USDT
	if symbol == "USDTUSDT" {
		return 1.0, nil // USDT is always 1:1 with USD
	}

	var response struct {
		Price string `json:"price"`
	}

	resp, err := c.httpClient.R().
		SetResult(&response).
		Get("/api/v3/ticker/price?symbol=" + symbol)

	if err != nil {
		log.Printf("❌ Error fetching price: %v", err)
		return 0, fmt.Errorf("error fetching price from Binance: %w", err)
	}

	if resp.StatusCode() != 200 {
		binanceErr := NewBinanceError(resp.StatusCode(), resp.String())
		log.Printf("❌ Binance API error: %v", binanceErr)
		return 0, binanceErr
	}

	price, err := strconv.ParseFloat(response.Price, 64)
	if err != nil {
		log.Printf("❌ Error parsing price: %v", err)
		return 0, fmt.Errorf("error parsing price from Binance: %w", err)
	}

	log.Printf("✅ Price for %s: $%.2f", symbol, price)
	return price, nil
}

// Get24hChange fetches 24h price change percentage for a symbol from Binance API.
// For USDT, it returns 0% since USDT is a stablecoin.
//
// Example usage:
//
//	change, err := client.Get24hChange("BTCUSDT")
//	if err != nil {
//	    if binanceErr, ok := err.(*BinanceError); ok {
//	        log.Printf("Binance API error %d: %s", binanceErr.Code, binanceErr.Message)
//	    }
//	    return 0, err
//	}
//	fmt.Printf("BTC 24h Change: %+.2f%%\n", change)
func (c *BinanceClient) Get24hChange(symbol string) (float64, error) {
	log.Printf("🔄 Fetching 24h change for %s", symbol)

	// Special case for USDT
	if symbol == "USDTUSDT" {
		return 0.0, nil // USDT is stable, no change
	}

	var response struct {
		PriceChangePercent string `json:"priceChangePercent"`
	}

	resp, err := c.httpClient.R().
		SetResult(&response).
		Get("/api/v3/ticker/24hr?symbol=" + symbol)

	if err != nil {
		log.Printf("❌ Error fetching 24h change: %v", err)
		return 0, fmt.Errorf("error fetching 24h change from Binance: %w", err)
	}

	if resp.StatusCode() != 200 {
		binanceErr := NewBinanceError(resp.StatusCode(), resp.String())
		log.Printf("❌ Binance API error: %v", binanceErr)
		return 0, binanceErr
	}

	change, err := strconv.ParseFloat(response.PriceChangePercent, 64)
	if err != nil {
		log.Printf("❌ Error parsing 24h change: %v", err)
		return 0, fmt.Errorf("error parsing 24h change from Binance: %w", err)
	}

	log.Printf("✅ 24h change for %s: %.2f%%", symbol, change)
	return change, nil
}

// generateSignature creates an HMAC SHA256 signature for Binance API authentication.
func (c *BinanceClient) generateSignature(queryString string) string {
	mac := hmac.New(sha256.New, []byte(c.apiSecret))
	mac.Write([]byte(queryString))
	return hex.EncodeToString(mac.Sum(nil))
}

// FormatPrice returns a formatted string representation of the price.
func (p *PriceData) FormatPrice() string {
	return fmt.Sprintf("$%.2f", p.Price)
}

// FormatPriceChange returns a formatted string representation of the price change.
func (p *PriceData) FormatPriceChange() string {
	return fmt.Sprintf("%+.2f%%", p.PriceChangePercent)
}

// String returns a string representation of the price data.
func (p *PriceData) String() string {
	return fmt.Sprintf("BTC: %s (%s) [%s]",
		p.FormatPrice(),
		p.FormatPriceChange(),
		p.Source)
}
