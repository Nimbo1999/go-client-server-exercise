package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const QUOTATION_SERVICE_URL = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

type Quotation struct {
	ID         int64  `json:"id"`
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

type ErrorResponse struct {
	Message string `json:"message"`
	Reason  string `json:"reason"`
	Status  int    `json:"status"`
}

type USDToBRL struct {
	USDBRL Quotation `json:"USDBRL"`
}

func main() {
	mux := http.NewServeMux()
	registerRoutes(mux)
	http.ListenAndServe(":8080", mux)
}

func NewSQLite() (*sql.DB, error) {
	const dsn = "./db/test.db?cache=shared&mode=memory"
	return sql.Open("sqlite3", dsn)
}

func registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/cotacao", CotacaoHandlerFunction)
}

func CotacaoHandlerFunction(w http.ResponseWriter, r *http.Request) {
	log.Println("Iniciando cotacao controller handler!")
	defer log.Println("Finalizando cotacao controller handler!")
	w.Header().Add("Content-Type", "application/json")

	usdToBrl, err := NewHttpClientRequest()
	if err != nil {
		log.Println(err)
		SendError(w, &ErrorResponse{
			Message: "There was an error while retrieving the current quotation.",
			Reason:  err.Error(),
			Status:  http.StatusInternalServerError,
		})
		return
	}

	err = SaveQuotation(usdToBrl)
	if err != nil {
		log.Println(err)
		SendError(w, &ErrorResponse{
			Message: "There was an error while persisting this quotation into the DB.",
			Reason:  err.Error(),
			Status:  http.StatusInternalServerError,
		})
		return
	}

	data, err := json.Marshal(usdToBrl.USDBRL)
	if err != nil {
		log.Println(err)
		SendError(w, &ErrorResponse{
			Message: "Unabled to marshal json data.",
			Reason:  err.Error(),
			Status:  http.StatusInternalServerError,
		})
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func NewHttpClientRequest() (*USDToBRL, error) {
	client := NewHttpClient()
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, "GET", QUOTATION_SERVICE_URL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	content := &USDToBRL{}
	err = json.Unmarshal(data, content)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func NewHttpClient() *http.Client {
	return &http.Client{}
}

func SendError(w http.ResponseWriter, err *ErrorResponse) {
	w.WriteHeader(err.Status)
	json.NewEncoder(w).Encode(err)
}

func SaveQuotation(usdToBrl *USDToBRL) error {
	db, err := NewSQLite()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	stmt, err := db.PrepareContext(ctx, "INSERT INTO quotation(code, codein, name, high, low, var_bid, pct_change, bid, ask, timestamp, create_date) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	quo := usdToBrl.USDBRL
	result, err := stmt.ExecContext(ctx, quo.Code, quo.Codein, quo.Name, quo.High, quo.Low, quo.VarBid, quo.PctChange, quo.Bid, quo.Ask, quo.Timestamp, quo.CreateDate)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	usdToBrl.USDBRL.ID = id
	rows, err := result.RowsAffected()

	if err != nil {
		return err
	}

	log.Println("Quotation added successfully!")
	log.Printf("Rows affected %d\n", rows)
	return nil
}
