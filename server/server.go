package server

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"net/http"
)

type Server struct {
	r *chi.Mux
}

func NewServer() *Server {
	s := &Server{}
	s.r = chi.NewRouter()

	s.r.Use(middleware.Logger)

	s.r.Route("/address", func(r chi.Router) {
		r.Get("/{account}", s.GetTransactions)
	})

	return s
}

func (s *Server) GetTransactions(w http.ResponseWriter, r *http.Request)  {
	fmt.Println(r)

	addr := chi.URLParam(r, "account")
	fmt.Println(addr)
}

func (s *Server) Start() {
	http.ListenAndServe("0.0.0.0:8081", s.r)
}