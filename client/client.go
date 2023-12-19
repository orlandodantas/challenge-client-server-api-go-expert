package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	ExchangeRateURL = "http://localhost:8080/cotacao"
	Timeout         = 300 * time.Millisecond
	FileName        = "context.txt"
)

type ExchangeRate struct {
	Bid string `json:"bid"`
}

func main() {
	exchangeRate, err := fetchCurrentExchangeRate()
	if err != nil {
		log.Fatal(err)
	}

	err = saveRateToFile(exchangeRate)
	if err != nil {
		log.Fatal(err)
	}
}

func fetchCurrentExchangeRate() (*ExchangeRate, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ExchangeRateURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("unexpected status code: %d, body: %s", res.StatusCode, body)
	}

	var exchangeRate ExchangeRate

	err = json.NewDecoder(res.Body).Decode(&exchangeRate)
	if err != nil {
		return nil, err
	}

	return &exchangeRate, nil
}

func saveRateToFile(exchangeRate *ExchangeRate) error {
	err := os.WriteFile(FileName, []byte("Dólar: "+exchangeRate.Bid), 0644)
	if err != nil {
		return err
	}

	return nil
}
