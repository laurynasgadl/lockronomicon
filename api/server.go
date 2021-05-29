package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/laurynasgadl/lockronomicon/pkg/locker"
)

type Server struct {
	router *mux.Router
	locker locker.Locker
}

func NewServer(locker locker.Locker) *Server {
	s := &Server{
		router: mux.NewRouter(),
		locker: locker,
	}
	routes(s)
	return s
}

func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.router)
}

func (s *Server) apiHandle(fn func(w http.ResponseWriter, r *http.Request) (int, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status, err := fn(w, r)

		if status != 0 {
			txt := http.StatusText(status)
			http.Error(w, fmt.Sprintf("%d %s", status, txt), status)
		}

		if status >= 400 || err != nil {
			log.Printf("%s: %v %v", r.URL.Path, status, err)
		}
	})
}
