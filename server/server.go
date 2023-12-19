package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	ExchangeRateURL    = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	DatabasePath       = "./database.db"
	ServerErrorMessage = "A server error has occurred. We will resolve it as quickly as possible!"
)

type Quotation struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

func main() {
	mux := setupRoutes()

	log.Fatalf("Server exited: %v", http.ListenAndServe(":8080", mux))
}

func setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", handleQuotationRequest)

	return mux
}

func fetchDollarExchangeRate() (*Quotation, error) {
	type quotation struct {
		UsdToBrl Quotation `json:"USDBRL"`
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ExchangeRateURL, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var quotationResponse quotation

	err = json.NewDecoder(res.Body).Decode(&quotationResponse)
	if err != nil {
		return nil, err
	}

	return &quotationResponse.UsdToBrl, nil
}

func saveQuotationToDatabase(exchange *Quotation, db *sql.DB) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	sqlStmt, err := db.Prepare("INSERT INTO quotation(code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, create_date) VALUES(?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer sqlStmt.Close()

	_, err = sqlStmt.ExecContext(ctx, exchange.Code, exchange.Codein, exchange.Name, exchange.High, exchange.Low, exchange.VarBid, exchange.PctChange, exchange.Bid, exchange.Ask, exchange.Timestamp, exchange.CreateDate)
	if err != nil {
		return err
	}

	return nil
}

func handleQuotationRequest(responseWriter http.ResponseWriter, request *http.Request) {
	quotation, err := fetchDollarExchangeRate()
	if err != nil {
		log.Println(err.Error())
		http.Error(responseWriter, ServerErrorMessage, http.StatusInternalServerError)
		return
	}

	db, err := startDatabase()
	if err != nil {
		log.Println(err.Error())
		http.Error(responseWriter, ServerErrorMessage, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	err = saveQuotationToDatabase(quotation, db)
	if err != nil {
		log.Println(err.Error())
		http.Error(responseWriter, ServerErrorMessage, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(responseWriter).Encode(quotation)
}

func startDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", DatabasePath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS quotation 
        (id INTEGER PRIMARY KEY, code Text, codein TEXT, name TEXT, high TEXT, low TEXT, varBid TEXT,
        pctChange TEXT, bid TEXT, ask TEXT, timestamp TEXT, create_date TEXT)`,
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}
