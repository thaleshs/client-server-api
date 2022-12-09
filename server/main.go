package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Cotacao struct {
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

type CotacaoResponse struct {
	Bid string `json:"bid"`
}

var db *sql.DB

func main() {
	dbsqlite, err := sql.Open("sqlite3", "cotacoes.db")
	if err != nil {
		log.Println(err)
	}
	db = dbsqlite

	fmt.Printf("rodando na porta: 8080")

	http.HandleFunc("/cotacao", CotacaoHandler)
	http.ListenAndServe(":8080", nil)
}

func CotacaoHandler(w http.ResponseWriter, r *http.Request) {
	ctxCotacao, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer cancel()

	cotacao, err := BuscaCotacao(ctxCotacao)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := CotacaoResponse{
		Bid: cotacao.Bid,
	}

	json.NewEncoder(w).Encode(response)

	ctxDb, cancelDb := context.WithTimeout(context.Background(), 10*time.Microsecond)
	defer cancelDb()

	err = InsertOneCotacao(ctxDb, db, cotacao)
	if err != nil {
		log.Println(err)
	}
}

func BuscaCotacao(ctx context.Context) (*Cotacao, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]any
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	js, err := json.Marshal(data["USDBRL"])
	if err != nil {
		return nil, err
	}

	var cotacao Cotacao
	err = json.Unmarshal(js, &cotacao)
	if err != nil {
		return nil, err
	}

	return &cotacao, nil
}

func InsertOneCotacao(ctx context.Context, db *sql.DB, cotacao *Cotacao) error {
	stmt, err := db.PrepareContext(ctx, "insert into cotacoes(code, code_in, name, high, low, var_bid, pct_change, bid, ask, timestamp, create_date) values (?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(cotacao.Code, cotacao.Codein, cotacao.Name,
		cotacao.High, cotacao.Low, cotacao.VarBid, cotacao.PctChange,
		cotacao.Bid, cotacao.Ask, cotacao.Timestamp, cotacao.CreateDate)
	if err != nil {
		return err
	}
	return nil
}
