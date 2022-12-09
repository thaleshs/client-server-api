package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type Cotacao struct {
	Bid string `json:"bid"`
}

func main() {
	ctxCotacao, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxCotacao, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic("app encerrado")
		// fmt.Fprintf(os.Stderr, "Erro ao criar requisição %v\n", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer requisição %v\n", err)
		panic("app encerrado")
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer ler respota %v\n", err)
		panic("app encerrado")
	}

	var cotacao Cotacao
	err = json.Unmarshal(body, &cotacao)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer parse da respota %v\n", err)
		panic("app encerrado")
	}

	f, err := os.Create("cotacao.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao criar o arquivo %v\n", err)
		panic("app encerrado")
	}
	defer f.Close()

	_, err = f.WriteString("Dólar: { " + cotacao.Bid + " }")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao gravar no arquivo %v\n", err)
		panic("app encerrado")
	}

	fmt.Println("Arquivo criado com sucesso")

}
