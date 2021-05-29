package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"time"

	"github.com/gorilla/mux"
	"github.com/laurynasgadl/lockronomicon/pkg/locker"
)

type LockCreateRequest struct {
	Key string `json:"key"`
	Ttl int64  `json:"ttl"`
}

type LockRefreshRequest struct {
	Key        string `json:"key"`
	Generation int64  `json:"generation"`
}

type LockResponse struct {
	Generation int64 `json:"generation"`
}

func (s *Server) handleLockCreate(w http.ResponseWriter, r *http.Request) (int, error) {
	var body LockCreateRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return http.StatusBadRequest, err
	}

	if body.Key == "" {
		return http.StatusUnprocessableEntity, nil
	}

	matched, err := regexp.Match(`^[\w.-]+$`, []byte(body.Key))
	if !matched || err != nil {
		return http.StatusUnprocessableEntity, nil
	}

	gen, err := s.locker.Lock(body.Key, time.Second*time.Duration(body.Ttl))
	if err != nil {
		// check if the lock is expired and try to override if so
		if errors.Is(err, locker.ErrLockTaken) {
			gn, exp, e := s.locker.Expired(body.Key)
			if e != nil || !exp {
				return renderError(err)
			}

			e = s.locker.Release(body.Key, gn)
			if e != nil {
				return renderError(err)
			}

			gen, err = s.locker.Lock(body.Key, time.Second*time.Duration(body.Ttl))
			if err != nil {
				return renderError(err)
			}
		} else {
			return renderError(err)
		}
	}

	res := &LockResponse{
		Generation: gen,
	}

	return renderJSON(w, r, res)
}

func (s *Server) handleLockRefresh(w http.ResponseWriter, r *http.Request) (int, error) {
	var body LockRefreshRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return http.StatusBadRequest, err
	}

	if body.Generation < 1 {
		return http.StatusUnprocessableEntity, nil
	}

	vars := mux.Vars(r)
	gen, err := s.locker.Refresh(vars["key"], body.Generation)
	if err != nil {
		return renderError(err)
	}

	res := &LockResponse{
		Generation: gen,
	}

	return renderJSON(w, r, res)
}

func (s *Server) handleLockRelease(w http.ResponseWriter, r *http.Request) (int, error) {
	var body LockRefreshRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return http.StatusBadRequest, err
	}

	if body.Generation < 1 {
		return http.StatusUnprocessableEntity, nil
	}

	vars := mux.Vars(r)
	err = s.locker.Release(vars["key"], body.Generation)
	if err != nil {
		return renderError(err)
	}

	return http.StatusOK, nil
}
