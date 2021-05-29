package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/laurynasgadl/lockronomicon/pkg/locker"
)

const lockerRootDir = "/tmp/test/locker"

func execServerTest(t *testing.T, fn func(server *Server)) {
	l, err := locker.NewFsLocker(lockerRootDir)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	defer os.RemoveAll(lockerRootDir)

	s := NewServer(l)

	fn(s)
}

func TestCreatesLock(t *testing.T) {
	execServerTest(t, func(server *Server) {
		body := strings.NewReader(`{"key":"test","ttl":300}`)
		req := httptest.NewRequest("POST", "/api/locks", body)
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		if w.Result().StatusCode != http.StatusOK {
			t.Errorf("expected status code %d, received %d", http.StatusOK, w.Result().StatusCode)
		}

		var resp LockResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		if err != nil {
			t.Errorf("could not decode response: %v", err)
		}

		if resp.Generation == 0 {
			t.Errorf("invalid generation returned")
		}
	})
}

func TestReturnsLockedStatus(t *testing.T) {
	execServerTest(t, func(server *Server) {
		_, err := server.locker.Lock("test", 300*time.Second)
		if err != nil {
			t.Errorf("unexpected error while locking: %v", err)
		}

		body := strings.NewReader(`{"key":"test","ttl":300}`)
		req := httptest.NewRequest("POST", "/api/locks", body)
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		if w.Result().StatusCode != http.StatusLocked {
			t.Errorf("expected status code %d, received %d", http.StatusOK, w.Result().StatusCode)
		}
	})
}

func TestOverridesExpiredLock(t *testing.T) {
	execServerTest(t, func(server *Server) {
		_, err := server.locker.Lock("test", 0)
		if err != nil {
			t.Errorf("unexpected error while locking: %v", err)
		}

		time.Sleep(1 * time.Second)

		body := strings.NewReader(`{"key":"test","ttl":300}`)
		req := httptest.NewRequest("POST", "/api/locks", body)
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		if w.Result().StatusCode != http.StatusOK {
			t.Errorf("expected status code %d, received %d", http.StatusOK, w.Result().StatusCode)
		}

		var resp LockResponse
		err = json.NewDecoder(w.Body).Decode(&resp)
		if err != nil {
			t.Errorf("could not decode response: %v", err)
		}

		if resp.Generation == 0 {
			t.Errorf("invalid generation returned")
		}
	})
}

func TestRefreshChangesLockGeneration(t *testing.T) {
	execServerTest(t, func(server *Server) {
		key := "test"
		gn, err := server.locker.Lock(key, 300*time.Second)
		if err != nil {
			t.Errorf("unexpected error while locking: %v", err)
		}

		time.Sleep(1 * time.Second)

		body := strings.NewReader(fmt.Sprintf(`{"generation":%d}`, gn))
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/locks/%s", key), body)
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		if w.Result().StatusCode != http.StatusOK {
			t.Errorf("expected status code %d, received %d", http.StatusOK, w.Result().StatusCode)
		}

		var resp LockResponse
		err = json.NewDecoder(w.Body).Decode(&resp)
		if err != nil {
			t.Errorf("could not decode response: %v", err)
		}

		if resp.Generation == 0 {
			t.Errorf("invalid generation returned")
		}

		if resp.Generation == gn {
			t.Errorf("old generation returned")
		}
	})
}

func TestRefreshFailsOnNonExistingLock(t *testing.T) {
	execServerTest(t, func(server *Server) {
		key := "test"
		gn, err := server.locker.Lock(key, 300*time.Second)
		if err != nil {
			t.Errorf("unexpected error while locking: %v", err)
		}

		err = server.locker.Release(key, gn)
		if err != nil {
			t.Errorf("unexpected error while releasing lock: %v", err)
		}

		time.Sleep(1 * time.Second)

		body := strings.NewReader(fmt.Sprintf(`{"generation":%d}`, gn))
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/locks/%s", key), body)
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		if w.Result().StatusCode != http.StatusNotFound {
			t.Errorf("expected status code %d, received %d", http.StatusOK, w.Result().StatusCode)
		}
	})
}

func TestRefreshFailsOnRegeneratedLock(t *testing.T) {
	execServerTest(t, func(server *Server) {
		key := "test"
		gn, err := server.locker.Lock(key, 300*time.Second)
		if err != nil {
			t.Errorf("unexpected error while locking: %v", err)
		}

		_, err = server.locker.Refresh(key, gn)
		if err != nil {
			t.Errorf("unexpected error while refreshing lock: %v", err)
		}

		time.Sleep(1 * time.Second)

		body := strings.NewReader(fmt.Sprintf(`{"generation":%d}`, gn))
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/locks/%s", key), body)
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		if w.Result().StatusCode != http.StatusPreconditionFailed {
			t.Errorf("expected status code %d, received %d", http.StatusOK, w.Result().StatusCode)
		}
	})
}

func TestReleaseRemovesLock(t *testing.T) {
	execServerTest(t, func(server *Server) {
		key := "test"
		gn, err := server.locker.Lock(key, 300*time.Second)
		if err != nil {
			t.Errorf("unexpected error while locking: %v", err)
		}

		body := strings.NewReader(fmt.Sprintf(`{"generation":%d}`, gn))
		req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/locks/%s", key), body)
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		if w.Result().StatusCode != http.StatusOK {
			t.Errorf("expected status code %d, received %d", http.StatusOK, w.Result().StatusCode)
		}

		_, _, err = server.locker.Expired(key)
		if !errors.Is(err, locker.ErrLockNotExist) {
			t.Errorf("expected error %v, received %v", locker.ErrLockNotExist, err)
		}
	})
}

func TestReleaseFailsOnNonExistingLock(t *testing.T) {
	execServerTest(t, func(server *Server) {
		key := "test"
		gn, err := server.locker.Lock(key, 300*time.Second)
		if err != nil {
			t.Errorf("unexpected error while locking: %v", err)
		}

		err = server.locker.Release(key, gn)
		if err != nil {
			t.Errorf("unexpected error while releasing lock: %v", err)
		}

		time.Sleep(1 * time.Second)

		body := strings.NewReader(fmt.Sprintf(`{"generation":%d}`, gn))
		req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/locks/%s", key), body)
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		if w.Result().StatusCode != http.StatusNotFound {
			t.Errorf("expected status code %d, received %d", http.StatusOK, w.Result().StatusCode)
		}
	})
}

func TestReleaseFailsOnRegeneratedLock(t *testing.T) {
	execServerTest(t, func(server *Server) {
		key := "test"
		gn, err := server.locker.Lock(key, 300*time.Second)
		if err != nil {
			t.Errorf("unexpected error while locking: %v", err)
		}

		_, err = server.locker.Refresh(key, gn)
		if err != nil {
			t.Errorf("unexpected error while refreshing lock: %v", err)
		}

		time.Sleep(1 * time.Second)

		body := strings.NewReader(fmt.Sprintf(`{"generation":%d}`, gn))
		req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/locks/%s", key), body)
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		if w.Result().StatusCode != http.StatusPreconditionFailed {
			t.Errorf("expected status code %d, received %d", http.StatusOK, w.Result().StatusCode)
		}
	})
}
