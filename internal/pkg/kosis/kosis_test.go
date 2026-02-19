package kosis

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMakeRequest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expectedResponse := []KosisSearchResponse{
			{OrgID: "101", TblID: "DT_1BPA001", MtAtitle: "Test Title"},
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedResponse)
		}))
		defer server.Close()

		client := New("test-key")
		var actualResponse []KosisSearchResponse
		err := client.makeRequest(server.URL, &actualResponse)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(actualResponse) != 1 || actualResponse[0].MtAtitle != "Test Title" {
			t.Errorf("expected %+v, got %+v", expectedResponse, actualResponse)
		}
	})

	t.Run("error", func(t *testing.T) {
		expectedError := KosisSearchErrorResponse{
			Err:    "20",
			ErrMsg: "필수요청변수값이 누락되었습니다.",
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(expectedError)
		}))
		defer server.Close()

		client := New("test-key")
		var actualResponse []KosisSearchResponse
		err := client.makeRequest(server.URL, &actualResponse)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		expectedErrStr := "20: 필수요청변수값이 누락되었습니다."
		if err.Error() != expectedErrStr {
			t.Errorf("expected error %q, got %q", expectedErrStr, err.Error())
		}
	})
}
