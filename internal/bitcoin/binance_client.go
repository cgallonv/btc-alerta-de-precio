// Package bitcoin provides functionality for interacting with cryptocurrency APIs.
package bitcoin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"
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
	httpClient    *resty.Client
	apiKey        string
	apiSecret     string
	tickerStorage *TickerStorage
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
func NewBinanceClient(apiKey, apiSecret string, baseURL string, tickerStorage *TickerStorage) *BinanceClient {
	client := resty.New()
	client.SetTimeout(10 * time.Second)
	client.SetHeader("X-MBX-APIKEY", apiKey)

	if baseURL == "" {
		baseURL = "https://api.binance.com" // Fallback to production if not specified
		log.Printf("⚠️ No Binance URL specified, using default: %s", baseURL)
	} else {
		log.Printf("🔧 Using Binance API: %s", baseURL)
	}
	client.SetBaseURL(baseURL)

	return &BinanceClient{
		httpClient:    client,
		apiKey:        apiKey,
		apiSecret:     apiSecret,
		tickerStorage: tickerStorage,
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
func (c *BinanceClient) GetAccountBalance(symbols []string) (*AccountBalance, error) {
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

	// Create a map for quick symbol lookup
	symbolMap := make(map[string]bool)
	if len(symbols) > 0 {
		for _, symbol := range symbols {
			symbolMap[strings.ToUpper(strings.TrimSpace(symbol))] = true
		}
		log.Printf("🔍 Filtering for symbols: %v", symbols)
	}

	// First filter and parse balances to avoid unnecessary API calls
	type parsedBalance struct {
		asset  string
		free   float64
		locked float64
		total  float64
	}
	var validBalances []parsedBalance

	// First pass: collect all balances that match our filter
	for _, balance := range response.Balances {
		// Skip if we have a symbol filter and this asset is not in it
		if len(symbolMap) > 0 && !symbolMap[balance.Asset] {
			continue
		}

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

		// Only include non-zero balances or specifically requested symbols
		if total > 0 || (len(symbolMap) > 0 && symbolMap[balance.Asset]) {
			validBalances = append(validBalances, parsedBalance{
				asset:  balance.Asset,
				free:   free,
				locked: locked,
				total:  total,
			})
		}
	}

	// Log the filtered balances before processing
	log.Printf("📋 Processing %d filtered assets: %v", len(validBalances), func() []string {
		var assets []string
		for _, b := range validBalances {
			assets = append(assets, b.asset)
		}
		return assets
	}())

	// Now process only the filtered balances
	for _, balance := range validBalances {
		// Skip price lookup for non-tradeable assets (like COP)
		var price, change24h float64
		if balance.asset != "COP" {
			// Get current price for the asset
			price, err = c.GetAssetPrice(balance.asset + "USDT")
			if err != nil {
				log.Printf("⚠️ Error getting price for %s: %v", balance.asset, err)
				price = 0
			}

			// Get 24h change
			change24h, err = c.Get24hChange(balance.asset + "USDT")
			if err != nil {
				log.Printf("⚠️ Error getting 24h change for %s: %v", balance.asset, err)
				change24h = 0
			}
		} else {
			log.Printf("ℹ️ Skipping price lookup for non-tradeable asset: %s", balance.asset)
		}

		valueUSD := balance.total * price
		accountBalance.TotalBalance += valueUSD
		accountBalance.AvailableBalance += balance.free * price

		accountBalance.Assets = append(accountBalance.Assets, AssetBalance{
			Symbol:    balance.asset,
			Free:      fmt.Sprintf("%.8f", balance.free),
			Locked:    fmt.Sprintf("%.8f", balance.locked),
			Total:     balance.total,
			ValueUSD:  valueUSD,
			Change24h: change24h,
		})

		log.Printf("📊 Asset %s: Free: %.8f, Locked: %.8f, Total: %.8f, Value: $%.2f, Change: %.2f%%",
			balance.asset, balance.free, balance.locked, balance.total, valueUSD, change24h)
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

	var response Ticker24hResponse
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

	// Store ticker data if storage is configured
	if c.tickerStorage != nil {
		if err := c.tickerStorage.StoreTicker24h("BTCUSDT", &response); err != nil {
			log.Printf("❌ Error storing ticker data: %v", err)
			// Don't return error here, continue with price update
		}
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

// GetHistoricalKlines fetches historical kline/candlestick data for a symbol.
// Interval can be: 1m,3m,5m,15m,30m,1h,2h,4h,6h,8h,12h,1d,3d,1w,1M
//
// Example usage:
//
//	klines, err := client.GetHistoricalKlines("BTCUSDT", "1m", time.Now().Add(-60*24*time.Hour), time.Now())
//	if err != nil {
//	    log.Printf("Error: %v", err)
//	    return
//	}
func (c *BinanceClient) GetHistoricalKlines(symbol, interval string, startTime, endTime time.Time) ([]Ticker24hResponse, error) {
	log.Printf("🔄 Fetching historical klines for %s from %s to %s", symbol, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))

	var klines [][]interface{}
	resp, err := c.httpClient.R().
		SetQueryParams(map[string]string{
			"symbol":    symbol,
			"interval":  interval,
			"startTime": fmt.Sprintf("%d", startTime.UnixMilli()),
			"endTime":   fmt.Sprintf("%d", endTime.UnixMilli()),
			"limit":     "1000", // Max limit per request
		}).
		SetResult(&klines).
		Get("/api/v3/klines")

	if err != nil {
		log.Printf("❌ Error fetching historical klines: %v", err)
		return nil, fmt.Errorf("error fetching historical klines: %w", err)
	}

	if resp.StatusCode() != 200 {
		binanceErr := NewBinanceError(resp.StatusCode(), resp.String())
		log.Printf("❌ Binance API error: %v", binanceErr)
		return nil, binanceErr
	}

	var tickers []Ticker24hResponse
	for _, k := range klines {
		// Convert kline data to Ticker24hResponse format
		openTime := k[0].(float64)
		closeTime := k[6].(float64)
		openPrice := k[1].(string)
		highPrice := k[2].(string)
		lowPrice := k[3].(string)
		closePrice := k[4].(string)
		volume := k[5].(string)
		quoteVolume := k[7].(string)
		trades := k[8].(float64)

		priceChange := fmt.Sprintf("%f", stringToFloat64(closePrice)-stringToFloat64(openPrice))
		priceChangePercent := fmt.Sprintf("%f", (stringToFloat64(closePrice)-stringToFloat64(openPrice))/stringToFloat64(openPrice)*100)

		ticker := Ticker24hResponse{
			Symbol:             symbol,
			OpenTime:           int64(openTime),
			CloseTime:          int64(closeTime),
			OpenPrice:          openPrice,
			HighPrice:          highPrice,
			LowPrice:           lowPrice,
			LastPrice:          closePrice,
			Volume:             volume,
			QuoteVolume:        quoteVolume,
			Count:              int64(trades),
			PriceChange:        priceChange,
			PriceChangePercent: priceChangePercent,
			WeightedAvgPrice:   closePrice, // Using close price as weighted avg
			PrevClosePrice:     openPrice,  // Using open price as prev close
		}
		tickers = append(tickers, ticker)
	}

	log.Printf("✅ Fetched %d historical klines", len(tickers))
	return tickers, nil
}

// Helper function to convert string to float64
func stringToFloat64(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
