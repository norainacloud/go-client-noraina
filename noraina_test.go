package noraina

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestCheck2xxResponse(t *testing.T) {
	r := &http.Response{
		StatusCode: http.StatusCreated,
	}

	err := CheckResponse(r)

	if err != nil {
		t.Errorf("CheckResponse for a 2xx Response should return nil", r.StatusCode)
	}
}
