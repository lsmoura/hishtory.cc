package server

import (
	"bytes"
	"context"
	"encoding/json"
	"gihtub.com/lsmoura/hishtory_cc/internal/db"
	"gihtub.com/lsmoura/hishtory_cc/internal/model"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testDB(t *testing.T) *db.DB {
	t.Helper()

	database, err := db.NewWithSQLiteDSN(":memory:")
	require.NoError(t, err, "NewWithSQLiteDSN")

	t.Cleanup(func() {
		require.NoError(t, database.Close(), "db.Close")
	})

	require.NoError(t, database.Migrate())

	return database
}

func helperAPISubmit(t *testing.T, ctx context.Context, mux http.Handler, data []*model.EncHistoryEntry) *httptest.ResponseRecorder {
	t.Helper()

	body, err := json.Marshal(data)
	require.NoError(t, err, "json.Marshal")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/submit", bytes.NewReader(body))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req.WithContext(ctx))

	return w
}

func helperAPIRegister(t *testing.T, ctx context.Context, mux http.Handler, userID, deviceID string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/register?user_id="+userID+"&device_id="+deviceID, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req.WithContext(ctx))

	return w
}

func TestHandleRegister(t *testing.T) {
	database := testDB(t)

	server := New(database)
	mux := server.mux()

	t.Run("invalid request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/register", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("valid request", func(t *testing.T) {
		w := helperAPIRegister(t, context.Background(), mux, "foo", "bar")
		require.Equal(t, http.StatusNoContent, w.Code)

		devices, err := database.DevicesForUser(context.Background(), "foo")
		require.NoError(t, err, "database.DevicesForUser")
		require.Len(t, devices, 1)
	})
}

func TestHandleSubmit(t *testing.T) {
	database := testDB(t)

	server := New(database)
	ctx := context.Background()
	var mux http.Handler = server.mux()

	userID := "foo"
	deviceID := "bar"

	w := helperAPIRegister(t, ctx, mux, userID, deviceID)
	require.Equal(t, http.StatusNoContent, w.Code)

	t.Run("with invalid userID", func(t *testing.T) {
		data := []*model.EncHistoryEntry{
			{
				UserID:        "invalid",
				DeviceID:      deviceID,
				EncryptedData: []byte("data"),
			},
		}
		w := helperAPISubmit(t, ctx, mux, data)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("with invalid data format", func(t *testing.T) {
		body := `{"user_id":"invalid","device_id":"bar","enc_data":"data"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/submit", strings.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("valid request", func(t *testing.T) {
		data := []*model.EncHistoryEntry{
			{
				UserID:        userID,
				DeviceID:      deviceID,
				EncryptedData: []byte("data"),
			},
		}
		w := helperAPISubmit(t, ctx, mux, data)

		require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())
	})
}
