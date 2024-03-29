package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	postgres "github.com/Kiraco/usd_to_mxn_exchange_rate/data-collector/data"
	provider "github.com/Kiraco/usd_to_mxn_exchange_rate/data-collector/provider"
)

func getDiarioOficialProvider(db *sql.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	provider := provider.GetDiarioOficialFederacionProvider()
	postgres.InsertInto("diario_oficial", provider, db)
}

func getFixerProvider(db *sql.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	provider := provider.GetFixerProvider()
	postgres.InsertInto("fixer", provider, db)
}

func getBanxicoProvider(db *sql.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	provider := provider.GetBanxicoProvider()
	postgres.InsertInto("banxico", provider, db)
}

func UpdateRates(w http.ResponseWriter, req *http.Request) {
	db := postgres.ConnectDB()
	var wg sync.WaitGroup
	wg.Add(3)
	go getDiarioOficialProvider(db, &wg)
	go getFixerProvider(db, &wg)
	go getBanxicoProvider(db, &wg)
	wg.Wait()
	db.Close()
	time.Sleep(2 * time.Second)
	io.WriteString(w, "Rates updated...!\n")
}

func GetRates(w http.ResponseWriter, req *http.Request) {
	tables := []string{"banxico", "fixer", "diario_oficial"}
	db := postgres.ConnectDB()
	providers := []provider.Provider{}
	for _, table := range tables {
		tmpProvider := provider.Provider{}
		query := fmt.Sprintf("SELECT * FROM %s WHERE \"rate\" <> '' ORDER BY \"updatedAt\" DESC LIMIT 1", table)
		db.QueryRow(query).Scan(&tmpProvider.ID, &tmpProvider.Name, &tmpProvider.Rate, &tmpProvider.UpdatedAt)
		providers = append(providers, tmpProvider)
	}
	resp, err := json.Marshal(providers)
	if err != nil {
		log.Fatal(err)
	}
	db.Close()
	w.Write(resp)
}

func Ping(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "App running...!\n")
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", Ping)
	router.HandleFunc("/update", UpdateRates)
	router.HandleFunc("/rates", GetRates)
	log.Fatal(http.ListenAndServe(":3000", handlers.CORS(handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}), handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}), handlers.AllowedOrigins([]string{"*"}))(router)))
}
