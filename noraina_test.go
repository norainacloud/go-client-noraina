package noraina

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

var (
	mux    *http.ServeMux
	ctx    = context.TODO()
	client *Client
	server *httptest.Server
)

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	client = NewClient(nil)
	url, _ := url.Parse(server.URL)
	client.BaseURL = url
}

func teardown() {
	server.Close()
}

func TestNewClient(t *testing.T) {
	c := NewClient(nil)

	if c.client != http.DefaultClient {
		t.Errorf("NewClient should default to a http.DefaultClient")
	}

	if c.BaseURL.String() != "https://nacp01.noraina.net/" {
		t.Errorf("NewClient BaseURL should default to https://nacp01.noraina.net/")
	}

	if c.Token != "" {
		t.Errorf("NewClient should have an empty token")
	}
}

func TestNewRequest(t *testing.T) {
	c := NewClient(nil)
	inUrl, outUrl := "/api/certificates", defaultBaseURL+"api/certificates"
	inBody, outBody := &InstanceRequest{Name: "test", Services: []InstanceServiceRequest{}}, `{"name":"test","services":[]}`+"\n"

	req, _ := c.NewRequest(ctx, http.MethodGet, inUrl, inBody)

	if req.URL.String() != outUrl {
		t.Errorf("NewRequest(%v) URL = %v, expected %v", inUrl, req.URL, outUrl)
	}

	// test body was JSON encoded
	body, _ := ioutil.ReadAll(req.Body)
	if string(body) != outBody {
		t.Errorf("NewRequest(%v)Body = %v, expected %v", inBody, string(body), outBody)
	}

	if req.Header.Get("x-access-token") != "" {
		t.Error("Access token should be empty if not set")
	}

	c.Token = "mytoken"
	req2, _ := c.NewRequest(ctx, http.MethodGet, inUrl, inBody)
	if req2.Header.Get("x-access-token") != "mytoken" {
		t.Error("Access token should be populated if set in client")
	}
}

func TestDo(t *testing.T) {
	setup()
	defer teardown()

	type foo struct {
		A string
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if m := http.MethodGet; m != r.Method {
			t.Errorf("Request method = %v, expected %v", r.Method, m)
		}
		fmt.Fprint(w, `{"A":"a"}`)
	})

	req, _ := client.NewRequest(ctx, http.MethodGet, "/", nil)
	body := new(foo)
	err := client.Do(context.Background(), req, body)
	if err != nil {
		t.Fatalf("Do(): %v", err)
	}

	expected := &foo{"a"}
	if !reflect.DeepEqual(body, expected) {
		t.Errorf("Response body = %v, expected %v", body, expected)
	}
}

func TestDo_httpError(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad Request", 400)
	})

	req, _ := client.NewRequest(ctx, http.MethodGet, "/", nil)
	err := client.Do(context.Background(), req, nil)

	if err == nil {
		t.Error("Expected HTTP 400 error.")
	}
}

func TestCheck2xxResponses(t *testing.T) {
	validStatuses := []int{http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent, 298, 299}
	for _, s := range validStatuses {
		r := &http.Response{
			StatusCode: s,
		}
		err := CheckResponse(r)
		if err != nil {
			t.Errorf("CheckResponse for a %v Response should return nil", s)
		}
	}
}

func TestNon2xxEmptyResponses(t *testing.T) {
	r := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}

	err := CheckResponse(r).(*ErrorResponse)
	if err == nil {
		t.Fatal("A 500 empty response should error")
	}
}

func TestNon2xxNonJsonResponses(t *testing.T) {
	r := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       ioutil.NopCloser(strings.NewReader("Fatal error")),
	}

	err := CheckResponse(r)
	if err == nil {
		t.Fatal("A 500 non json response should error")
	}

	if !strings.Contains(err.Error(), "invalid character") {
		t.Fatalf("A 500 non json response should error with invalid character message, got %v", err.Error())
	}
}

func TestNon2xxJsonResponses(t *testing.T) {
	r := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       ioutil.NopCloser(strings.NewReader(`{"status":"ko"}`)),
	}

	err := CheckResponse(r)
	if err == nil {
		t.Fatal("A 400 json response should error")
	}

	expected := &ErrorResponse{
		StatusCode: http.StatusBadRequest,
		Message: map[string]string{
			"status": "ko",
		},
	}

	if !reflect.DeepEqual(err, expected) {
		t.Errorf("Error = %#v, expected %#v", err, expected)
	}
}

func TestErrorResponse_Error(t *testing.T) {
	err := ErrorResponse{Message: map[string]string{"status": "ko"}}
	if err.Error() == "" {
		t.Errorf("Expected non-empty ErrorResponse.Error()")
	}

	if err.Error() != "status: ko" {
		t.Errorf("Error Response serialization Error() method not returning as expected")
	}
}
