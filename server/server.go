package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sadysnaat/assignment/model"
	"net/http"
	"strconv"
)

type Server struct {
	r   *chi.Mux
	db  *sql.DB
	srv *http.Server
}

func NewServer(dbURL string) *Server {
	s := &Server{}
	s.r = chi.NewRouter()

	s.r.Use(middleware.Logger)

	s.r.Route("/address", func(r chi.Router) {
		r.Get("/{account}", s.GetTransactions)
	})

	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		panic(err)
	}
	s.db = db
	s.srv = &http.Server{}
	return s
}

func (s *Server) GetTransactions(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)
	limit := 10
	offset := 0
	order := "asc"
	sortBy := "amount"
	addr := chi.URLParam(r, "account")
	fmt.Println(addr)

	ls := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(ls)
	if err != nil {
		fmt.Println(err)
		limit = 10
	}

	ofs := r.URL.Query().Get("offset")
	offset, err = strconv.Atoi(ofs)
	if err != nil {
		fmt.Println(err)
		offset = 0
	}

	order = r.URL.Query().Get("order")
	if order == "" || (order != "asc" && order != "desc") {
		order = "asc"
	}

	sortBy = r.URL.Query().Get("sortBy")
	if sortBy == "" || (sortBy != "amount" && sortBy != "time") {
		sortBy = "amount"
	}

	fmt.Println(limit, offset, order, sortBy)

	t := new(model.Transaction)
	t = t.WithDB(s.db)

	txs, err := t.TransactionsForAccount(common.HexToAddress(addr), limit, offset, order, sortBy)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(txs)

	if len(txs) == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write(nil)
		return
	}

	b, err := json.Marshal(txs)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)

	return
}

func (s *Server) Start(apiHost, apiPort string) error {
	s.srv.Addr = fmt.Sprintf("%s:%s", apiHost, apiPort)
	s.srv.Handler = s.r
	go func() {
		err := s.srv.ListenAndServe()
		if err != nil {
			fmt.Println(err)
		}
	}()
	return nil
}
