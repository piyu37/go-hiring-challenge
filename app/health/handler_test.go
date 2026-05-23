package health

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func openTestDB(t *testing.T, setup func(sqlmock.Sqlmock)) *gorm.DB {
	t.Helper()

	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	if setup != nil {
		setup(mock)
	}

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		DisableAutomaticPing: true,
	})
	require.NoError(t, err)

	return db
}

func TestHandleLive(t *testing.T) {
	handler := NewHandler(nil)
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()

	handler.HandleLive(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Equal(t, "ok", body["status"])
}

func TestHandleReady(t *testing.T) {
	t.Run("returns ready when database is reachable", func(t *testing.T) {
		db := openTestDB(t, func(mock sqlmock.Sqlmock) {
			mock.ExpectPing()
		})

		handler := NewHandler(db)
		req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
		rec := httptest.NewRecorder()

		handler.HandleReady(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("returns service unavailable when database ping fails", func(t *testing.T) {
		db := openTestDB(t, func(mock sqlmock.Sqlmock) {
			mock.ExpectPing().WillReturnError(errors.New("db down"))
		})

		handler := NewHandler(db)
		req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
		rec := httptest.NewRecorder()

		handler.HandleReady(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	})
}
