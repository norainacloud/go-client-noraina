package noraina

import (
	"net/http"
)

const (
	norainaDomain    = "https://nacp01.noraina.net/"
	loginRoute       = "api/login"
	instanceRoute    = "api/instance"
	certificateRoute = "api/certificate"
)

type NorainaApiClient struct {
	Client *http.Client
	Token  string
}

func New(email string, password string, httpClient *http.Client) (*NorainaApiClient, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	c := &NorainaApiClient{
		Client: httpClient,
	}

	token, err := c.GetAuthToken(email, password)
	if err != nil {
		return nil, err
	}
	c.Token = token

	return c, nil
}
