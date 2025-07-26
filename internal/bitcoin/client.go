package bitcoin

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	httpClient *resty.Client
}

type CoinDeskResponse struct {
	BPI struct {
		USD struct {
			Rate      string  `json:"rate"`
			RateFloat float64 `json:"rate_float"`
		} `json:"USD"`
	} `json:"bpi"`
	Time struct {
		Updated string `json:"updated"`
	} `json:"time"`
}

// BinanceTicker24hrResponse estructura para la respuesta de ticker de 24 horas de Binance
type BinanceTicker24hrResponse struct {
	Symbol             string `json:"symbol"`
	PriceChange        string `json:"priceChange"`
	PriceChangePercent string `json:"priceChangePercent"`
	WeightedAvgPrice   string `json:"weightedAvgPrice"`
	PrevClosePrice     string `json:"prevClosePrice"`
	LastPrice          string `json:"lastPrice"`
	LastQty            string `json:"lastQty"`
	BidPrice           string `json:"bidPrice"`
	BidQty             string `json:"bidQty"`
	AskPrice           string `json:"askPrice"`
	AskQty             string `json:"askQty"`
	OpenPrice          string `json:"openPrice"`
	HighPrice          string `json:"highPrice"`
	LowPrice           string `json:"lowPrice"`
	Volume             string `json:"volume"`
	QuoteVolume        string `json:"quoteVolume"`
	OpenTime           int64  `json:"openTime"`
	CloseTime          int64  `json:"closeTime"`
	FirstId            int64  `json:"firstId"`
	LastId             int64  `json:"lastId"`
	Count              int64  `json:"count"`
}

type PriceData struct {
	Price              float64   `json:"price"`
	PriceChangePercent float64   `json:"price_change_percent"`
	Currency           string    `json:"currency"`
	Timestamp          time.Time `json:"timestamp"`
	Source             string    `json:"source"`
}

func NewClient() *Client {
	client := resty.New()
	client.SetTimeout(10 * time.Second)
	client.SetHeader("User-Agent", "BTC-Price-Alert/1.0")

	return &Client{
		httpClient: client,
	}
}

func (c *Client) GetCurrentPrice() (*PriceData, error) {
	// Intentar primero con Binance API (más confiable y actualizada)
	price, err := c.getBinancePrice()
	if err == nil {
		return price, nil
	}

	// Si falla, intentar con CoinDesk como respaldo primario
	price, err = c.getCoinDeskPrice()
	if err == nil {
		return price, nil
	}

	// Si falla, intentar con CoinGecko como respaldo secundario
	return c.getCoinGeckoPrice()
}

func (c *Client) getCoinDeskPrice() (*PriceData, error) {
	var response CoinDeskResponse

	resp, err := c.httpClient.R().
		SetResult(&response).
		Get("https://api.coindesk.com/v1/bpi/currentprice.json")

	if err != nil {
		return nil, fmt.Errorf("error fetching price from CoinDesk: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("CoinDesk API returned status %d", resp.StatusCode())
	}

	return &PriceData{
		Price:              response.BPI.USD.RateFloat,
		PriceChangePercent: 0.0, // CoinDesk no proporciona datos de cambio porcentual
		Currency:           "USD",
		Timestamp:          time.Now(),
		Source:             "CoinDesk",
	}, nil
}

func (c *Client) getCoinGeckoPrice() (*PriceData, error) {
	var response map[string]interface{}

	resp, err := c.httpClient.R().
		SetResult(&response).
		Get("https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd")

	if err != nil {
		return nil, fmt.Errorf("error fetching price from CoinGecko: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("CoinGecko API returned status %d", resp.StatusCode())
	}

	bitcoin, ok := response["bitcoin"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format from CoinGecko")
	}

	priceInterface, ok := bitcoin["usd"]
	if !ok {
		return nil, fmt.Errorf("USD price not found in CoinGecko response")
	}

	var price float64
	switch v := priceInterface.(type) {
	case float64:
		price = v
	case int:
		price = float64(v)
	case string:
		price, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing price from string: %w", err)
		}
	default:
		return nil, fmt.Errorf("unexpected price type: %T", priceInterface)
	}

	return &PriceData{
		Price:              price,
		PriceChangePercent: 0.0, // CoinGecko simple price no proporciona datos de cambio porcentual
		Currency:           "USD",
		Timestamp:          time.Now(),
		Source:             "CoinGecko",
	}, nil
}

func (c *Client) GetPriceHistory(days int) ([]PriceData, error) {
	var response map[string]interface{}

	resp, err := c.httpClient.R().
		SetResult(&response).
		Get(fmt.Sprintf("https://api.coingecko.com/api/v3/coins/bitcoin/market_chart?vs_currency=usd&days=%d", days))

	if err != nil {
		return nil, fmt.Errorf("error fetching price history: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("CoinGecko API returned status %d", resp.StatusCode())
	}

	prices, ok := response["prices"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format for price history")
	}

	var priceHistory []PriceData
	for _, p := range prices {
		pricePoint, ok := p.([]interface{})
		if !ok || len(pricePoint) < 2 {
			continue
		}

		timestamp := time.Unix(int64(pricePoint[0].(float64))/1000, 0)
		price := pricePoint[1].(float64)

		priceHistory = append(priceHistory, PriceData{
			Price:              price,
			PriceChangePercent: 0.0, // Históricamente no calculamos cambio porcentual
			Currency:           "USD",
			Timestamp:          timestamp,
			Source:             "CoinGecko",
		})
	}

	return priceHistory, nil
}

// getBinancePrice obtiene el precio y estadísticas de 24 horas de Binance
func (c *Client) getBinancePrice() (*PriceData, error) {
	var response BinanceTicker24hrResponse

	resp, err := c.httpClient.R().
		SetResult(&response).
		Get("https://api.binance.com/api/v3/ticker/24hr?symbol=BTCUSDT")

	if err != nil {
		return nil, fmt.Errorf("error fetching price from Binance: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("Binance API returned status %d", resp.StatusCode())
	}

	price, err := strconv.ParseFloat(response.LastPrice, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing price from Binance: %w", err)
	}

	priceChangePercent, err := strconv.ParseFloat(response.PriceChangePercent, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing price change percent from Binance: %w", err)
	}

	return &PriceData{
		Price:              price,
		PriceChangePercent: priceChangePercent,
		Currency:           "USD",
		Timestamp:          time.Now(),
		Source:             "Binance",
	}, nil
}

func (p *PriceData) FormatPrice() string {
	return fmt.Sprintf("$%.2f %s", p.Price, p.Currency)
}

func (p *PriceData) FormatPriceChange() string {
	if p.PriceChangePercent > 0 {
		return fmt.Sprintf("+%.2f%%", p.PriceChangePercent)
	} else if p.PriceChangePercent < 0 {
		return fmt.Sprintf("%.2f%%", p.PriceChangePercent)
	}
	return "0.00%"
}

func (p *PriceData) String() string {
	changeStr := ""
	if p.Source == "Binance" && p.PriceChangePercent != 0 {
		changeStr = fmt.Sprintf(" (%s)", p.FormatPriceChange())
	}

	return fmt.Sprintf("Bitcoin: %s%s (Source: %s, Updated: %s)",
		p.FormatPrice(),
		changeStr,
		p.Source,
		p.Timestamp.Format("2006-01-02 15:04:05"))
}
