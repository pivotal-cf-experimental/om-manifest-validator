package fetcher

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"context"
	"golang.org/x/oauth2"
)

func NewOAuthHTTPClient(host, username, password string) (*http.Client, error) {
	conf := &oauth2.Config{
		ClientID:     "opsman",
		ClientSecret: "",
		Endpoint: oauth2.Endpoint{
			TokenURL: fmt.Sprintf("%s/oauth/token", host),
		},
	}

	httpclient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	insecureContext := context.Background()
	insecureContext = context.WithValue(insecureContext, oauth2.HTTPClient, httpclient)

	token, err := conf.PasswordCredentialsToken(insecureContext, username, password)
	if err != nil {
		return nil, err
	}

	return conf.Client(insecureContext, token), nil
}
