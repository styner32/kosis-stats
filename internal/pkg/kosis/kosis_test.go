package kosis

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
<<<<<<< HEAD
=======
	"net/url"
>>>>>>> ff700d9 (refactor: stream HTTP response body in KOSIS client)
	"testing"
)

func TestMakeRequest(t *testing.T) {
<<<<<<< HEAD
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
=======
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/success" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]*KosisSearchResponse{{OrgID: "101", OrgNm: "Test Org"}})
		} else if r.URL.Path == "/error" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(KosisSearchErrorResponse{Err: "400", ErrMsg: "Bad Request"})
		} else if r.URL.Path == "/invalid-json" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("invalid json"))
		}
	}))
	defer ts.Close()

	client := New("test-key")

	t.Run("success", func(t *testing.T) {
		var res []*KosisSearchResponse
		err := client.makeRequest(ts.URL+"/success", &res)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(res) != 1 || res[0].OrgNm != "Test Org" {
			t.Errorf("unexpected response: %v", res)
>>>>>>> ff700d9 (refactor: stream HTTP response body in KOSIS client)
		}
	})

	t.Run("error", func(t *testing.T) {
<<<<<<< HEAD
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
=======
		var res []*KosisSearchResponse
		err := client.makeRequest(ts.URL+"/error", &res)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		expectedErr := "400: Bad Request"
		if err.Error() != expectedErr {
			t.Errorf("expected error %q, got %q", expectedErr, err)
		}
	})

	t.Run("invalid-json", func(t *testing.T) {
		var res []*KosisSearchResponse
		err := client.makeRequest(ts.URL+"/invalid-json", &res)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestGet(t *testing.T) {
	client := New("test-key")

	client.client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		w := httptest.NewRecorder()
		// Path will be /openapi/test-path because baseURL is https://kosis.kr/openapi
		if req.URL.Path == "/openapi/test-path" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]MetaITM{{ObjID: "ITEM", ItmNM: "Test Item"}})
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
		return w.Result(), nil
	})

	t.Run("success", func(t *testing.T) {
		var data []MetaITM
		err := client.get("test-path", url.Values{}, &data)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(data) != 1 || data[0].ItmNM != "Test Item" {
			t.Errorf("unexpected response: %v", data)
		}
	})

	t.Run("error", func(t *testing.T) {
		client.client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			w := httptest.NewRecorder()
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
			return w.Result(), nil
		})

		var data []MetaITM
		err := client.get("test-path", url.Values{}, &data)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "kosis http 500: internal server error" {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
>>>>>>> ff700d9 (refactor: stream HTTP response body in KOSIS client)
