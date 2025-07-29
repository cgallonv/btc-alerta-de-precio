// Package bitcoin provides functionality for interacting with cryptocurrency APIs.
package bitcoin

import (
	"encoding/json"
	"fmt"
)

// BinanceError represents an error from the Binance API.
// It includes the error code, message, and HTTP status code.
//
// Example usage:
//
//	if err != nil {
//	    if binanceErr, ok := err.(*BinanceError); ok {
//	        switch binanceErr.Code {
//	        case ErrInvalidAPIKey:
//	            log.Printf("Invalid API key: %s", binanceErr.Message)
//	        case ErrRateLimitExceeded:
//	            log.Printf("Rate limit exceeded: %s", binanceErr.Message)
//	        default:
//	            log.Printf("Binance error: %s", binanceErr)
//	        }
//	    }
//	    return err
//	}
type BinanceError struct {
	Code    int    `json:"code"` // Binance error code
	Message string `json:"msg"`  // Error message from Binance
	Status  int    // HTTP status code
}

// Error implements the error interface.
// It returns a formatted error message including the error code and message.
func (e *BinanceError) Error() string {
	return fmt.Sprintf("Binance API error (code %d): %s", e.Code, e.Message)
}

// NewBinanceError creates a new BinanceError from a response.
// It parses the response body to extract the error code and message.
// If parsing fails, it uses the raw body as the error message.
//
// Example usage:
//
//	resp, err := http.Get("https://api.binance.com/api/v3/ticker/price")
//	if resp.StatusCode != 200 {
//	    return NewBinanceError(resp.StatusCode, resp.Body)
//	}
func NewBinanceError(status int, body string) *BinanceError {
	binanceErr := &BinanceError{Status: status}

	// Try to parse error response
	if body != "" {
		var response struct {
			Code    int    `json:"code"`
			Message string `json:"msg"`
		}
		if err := json.Unmarshal([]byte(body), &response); err == nil {
			binanceErr.Code = response.Code
			binanceErr.Message = response.Message
		} else {
			binanceErr.Message = body
		}
	}

	return binanceErr
}

// Common Binance error codes.
// These constants represent standard error codes returned by the Binance API.
// They can be used to handle specific error cases in your application.
//
// Example usage:
//
//	if err != nil {
//	    if binanceErr, ok := err.(*BinanceError); ok {
//	        switch binanceErr.Code {
//	        case ErrInvalidAPIKey:
//	            log.Printf("Please check your API key configuration")
//	        case ErrRateLimitExceeded:
//	            log.Printf("Please reduce request frequency")
//	        }
//	    }
//	}
const (
	// ErrInvalidAPIKey indicates invalid API key, IP, or insufficient permissions
	ErrInvalidAPIKey = -2015

	// ErrInvalidSignature indicates invalid request signature
	ErrInvalidSignature = -1022

	// ErrInvalidTimestamp indicates request timestamp is too far from server time
	ErrInvalidTimestamp = -1021

	// ErrRateLimitExceeded indicates too many requests within the time window
	ErrRateLimitExceeded = -1003

	// ErrIPRateLimitExceeded indicates IP-based rate limit exceeded
	ErrIPRateLimitExceeded = -1004

	// ErrInvalidSymbol indicates the requested trading symbol is invalid
	ErrInvalidSymbol = -1121

	// ErrSystemError indicates an unknown internal error
	ErrSystemError = -1000

	// ErrMalformedRequest indicates invalid request format
	ErrMalformedRequest = -1002

	// ErrUnauthorized indicates invalid API key format
	ErrUnauthorized = -2014

	// ErrTooManyRequests indicates request rate limit exceeded
	ErrTooManyRequests = -1429

	// ErrServiceUnavailable indicates Binance service is temporarily unavailable
	ErrServiceUnavailable = -1016

	// ErrUnexpectedResponse indicates unexpected response format from server
	ErrUnexpectedResponse = -1006

	// ErrTimeoutError indicates request timeout
	ErrTimeoutError = -1007

	// ErrInsufficientPermission indicates API key lacks required permissions
	ErrInsufficientPermission = -2010
)
