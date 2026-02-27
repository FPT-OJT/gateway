package errors_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FPT-OJT/gateway/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteJSON_SetsStatusCode(t *testing.T) {
	rr := httptest.NewRecorder()
	errors.WriteJSON(rr, http.StatusNotFound, errors.ErrNotFound)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestWriteJSON_SetsContentTypeHeader(t *testing.T) {
	rr := httptest.NewRecorder()
	errors.WriteJSON(rr, http.StatusOK, errors.ErrNotFound)
	assert.Equal(t, "application/json; charset=utf-8", rr.Header().Get("Content-Type"))
}

func TestWriteJSON_EncodesBodyCorrectly(t *testing.T) {
	rr := httptest.NewRecorder()
	errors.WriteJSON(rr, http.StatusUnauthorized, errors.ErrUnauthorized)

	var resp errors.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "unauthorized", resp.Code)
	assert.NotEmpty(t, resp.Message)
}

func TestWriteJSON_WithDetail(t *testing.T) {
	rr := httptest.NewRecorder()
	detail := errors.ErrorResponse{
		Code:    "bad_gateway",
		Message: "Upstream unavailable",
		Detail:  "upstream error: timeout",
	}
	errors.WriteJSON(rr, http.StatusBadGateway, detail)

	var resp errors.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "bad_gateway", resp.Code)
	assert.NotNil(t, resp.Detail)
}

func TestWriteJSON_InternalError_500(t *testing.T) {
	rr := httptest.NewRecorder()
	errors.WriteJSON(rr, http.StatusInternalServerError, errors.ErrInternal)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	var resp errors.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "internal_error", resp.Code)
}

func TestPredefinedErrors_HaveExpectedCodes(t *testing.T) {
	assert.Equal(t, "not_found", errors.ErrNotFound.Code)
	assert.Equal(t, "unauthorized", errors.ErrUnauthorized.Code)
	assert.Equal(t, "bad_gateway", errors.ErrBadGateway.Code)
	assert.Equal(t, "internal_error", errors.ErrInternal.Code)
}

func TestPredefinedErrors_HaveNonEmptyMessages(t *testing.T) {
	for _, e := range []errors.ErrorResponse{
		errors.ErrNotFound,
		errors.ErrUnauthorized,
		errors.ErrBadGateway,
		errors.ErrInternal,
	} {
		assert.NotEmpty(t, e.Message, "error %q should have a message", e.Code)
	}
}
