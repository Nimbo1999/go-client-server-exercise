package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Quotation struct {
	Bid string `json:"bid"`
}

const SERVER_URL = "http://localhost:8080/cotacao"

func main() {
	quotation, err := RequestQuotation()
	if err != nil {
		panic(err)
	}

	err = ReportQuotation(quotation)
	if err != nil {
		panic(err)
	}
}

func NewClient() *http.Client {
	return &http.Client{}
}

func RequestQuotation() (*Quotation, error) {
	client := NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", SERVER_URL, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	quotation := &Quotation{}
	err = json.Unmarshal(data, quotation)
	if err != nil {
		return nil, err
	}

	return quotation, nil
}

func ReportQuotation(quotation *Quotation) error {
	file, err := os.Create("cotacao.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	value, err := strconv.ParseFloat(quotation.Bid, 32)
	if err != nil {
		return err
	}

	data := []byte(fmt.Sprintf("Dólar: R$ %.2f\n", value))

	bytes, err := file.Write(data)
	if err != nil {
		return err
	}
	fmt.Printf("New value of USD Quotation was updated successfully!\nFile cotacao.txt generated with %d bytes.\n", bytes)
	return nil
}
