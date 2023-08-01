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
	mux.HandleFunc("/", HomeHandlerFunction)
	mux.HandleFunc("/cotacao", CotacaoHandlerFunction)
}

func HomeHandlerFunction(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "http://localhost:8080/cotacao", http.StatusSeeOther)
}

func CotacaoHandlerFunction(w http.ResponseWriter, r *http.Request) {
	resp, err := NewHttpClientRequest()
	if err != nil {
		SendError(w, &ErrorResponse{
			Message: "Could not finish your request",
			Reason:  err.Error(),
			Status:  http.StatusRequestTimeout,
		})
		return
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)

	if err != nil {
		SendError(w, &ErrorResponse{
			Message: "Could no read data from request",
			Reason:  err.Error(),
			Status:  http.StatusInternalServerError,
		})
		log.Println(err.Error())
		return
	}
	var quotation USDToBRL
	err = json.Unmarshal(data, &quotation)
	if err != nil {
		log.Println("Unabled to Unmarshal JSON data")
		SendError(w, &ErrorResponse{
			Message: "Error while unmarshal JSON",
			Reason:  err.Error(),
			Status:  http.StatusInternalServerError,
		})
		return
	}

	db, err := NewSQLite()
	if err != nil {
		SendError(w, &ErrorResponse{
			Message: "Error while unmarshal JSON",
			Reason:  err.Error(),
			Status:  http.StatusInternalServerError,
		})
		log.Println("Unabled to create a connection with sqlite")
		return
	}
	defer db.Close()
	err = SaveQuotation(db, &quotation)
	if err != nil {
		SendError(w, &ErrorResponse{
			Message: "Could not execute db operation",
			Reason:  err.Error(),
			Status:  http.StatusInternalServerError,
		})
		log.Println(err.Error())
		log.Println("Unabled to execute db operation")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(quotation)
	if err != nil {
		log.Println("Unabled to encode JSON")
		SendError(w, &ErrorResponse{
			Message: "Error while encoding JSON",
			Reason:  err.Error(),
			Status:  http.StatusInternalServerError,
		})
		return
	}
}

func NewHttpClientRequest() (*http.Response, error) {
	client := NewHttpClient()
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, "GET", QUOTATION_SERVICE_URL, nil)
	if err != nil {
		return nil, err
	}
	return client.Do(request)
}

func NewHttpClient() *http.Client {
	return &http.Client{}
}

func SendError(w http.ResponseWriter, err *ErrorResponse) {
	w.WriteHeader(err.Status)
	json.NewEncoder(w).Encode(err)
}

func SaveQuotation(db *sql.DB, usdToBrl *USDToBRL) error {
	stmt, err := db.Prepare("INSERT INTO quotation(code, codein, name, high, low, var_bid, pct_change, bid, ask, timestamp, create_date) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	quo := usdToBrl.USDBRL
	// codein, name, high, low, varBid, pctChange, bid, ask, timestamp, create_date
	result, err := stmt.Exec(quo.Code, quo.Codein, quo.Name, quo.High, quo.Low, quo.VarBid, quo.PctChange, quo.Bid, quo.Ask, quo.Timestamp, quo.CreateDate)
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
	log.Printf("Rows affected %d\n", rows)
	return nil
}
