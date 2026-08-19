package httpapi

import (
	"net/http/httptest"
	"testing"
)

func TestReadPaginationWithMax(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/api/spirits?limit=500&offset=3", nil)

	limit, offset, ok := readPaginationWithMax(recorder, request, 500)
	if !ok || limit != 500 || offset != 3 {
		t.Fatalf("unexpected pagination: limit=%d offset=%d ok=%v", limit, offset, ok)
	}
}

func TestReadPaginationKeepsDefaultMaximum(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/api/volunteers?limit=500", nil)

	_, _, ok := readPagination(recorder, request)
	if ok || recorder.Code != 400 {
		t.Fatalf("expected default pagination to reject limit 500, status=%d ok=%v", recorder.Code, ok)
	}
}
