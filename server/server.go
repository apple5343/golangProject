package server

import (
	"net/http"
	"server/config"
	"server/helpers"
	"server/server/db"

	"github.com/gorilla/mux"
)

type Server struct {
	db         *db.SqlDB
	router     *mux.Router
	config     *config.Config
	Calculator *helpers.Calculator
}

func Init(db *db.SqlDB, cfg *config.Config) (*Server, error) {
	router := mux.NewRouter()
	calculator, err := helpers.NewCalculator(cfg, *db)
	if err != nil {
		return &Server{}, err
	}
	s := &Server{db: db, config: cfg, router: router, Calculator: calculator}
	s.SetupRoutes()
	return s, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
