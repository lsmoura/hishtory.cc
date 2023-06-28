package server

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRemoteAddrFromContext(t *testing.T) {
	retrieveIPAddr := func(addr *string) http.Handler {
		fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*addr = RemoteAddrFromContext(r.Context())
			w.WriteHeader(http.StatusNoContent)
		})

		return applyMiddlewares(fn, extractRemoteAddr)
	}

	t.Run("random value", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Real-Ip", "foo")
		w := httptest.NewRecorder()

		var foundAddress string

		retrieveIPAddr(&foundAddress).ServeHTTP(w, req)

		assert.Equal(t, "foo", foundAddress)
	})

	t.Run("no value", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		var foundAddress string

		retrieveIPAddr(&foundAddress).ServeHTTP(w, req)

		// 192.0.2.1:1234 is the default address assigned by httptest.NewRequest()
		assert.Equal(t, "192.0.2.1:1234", foundAddress)
	})
}

func TestPanicGuardMiddleware(t *testing.T) {
	panicHandler := func(w http.ResponseWriter, r *http.Request) {
		panic("foo")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	panicGuardMiddleware(http.HandlerFunc(panicHandler)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
