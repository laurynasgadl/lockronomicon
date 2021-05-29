package api

import "net/http"

func routes(s *Server) {
	s.router.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"OK"}`))
	})

	api := s.router.PathPrefix("/api").Subrouter()
	api.Handle("/locks", s.apiHandle(s.handleLockCreate)).Methods("POST")
	api.Handle("/locks/{key:[\\w.-]+$}", s.apiHandle(s.handleLockRefresh)).Methods("PUT")
	api.Handle("/locks/{key:[\\w.-]+$}", s.apiHandle(s.handleLockRelease)).Methods("DELETE")
}
