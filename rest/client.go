package rest

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	libregraph "github.com/owncloud/libre-graph-api-go"
	webdav "github.com/studio-b12/gowebdav"
)

type AuthCredentials struct {
	Username string
	Password string
}

type Client struct {
	Protocol string
	Host string
	Port string
	Auth AuthCredentials
	GraphAPICtx context.Context
	GraphAPI *libregraph.APIClient
	WebdavAPI *webdav.Client
}

type AuthTransport struct {
	BaseTransport http.RoundTripper
	Username      string
	Password      string
}

func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// All headers are resetting after redirect.
	// Set basic auth manually for webdav.
	req.SetBasicAuth(t.Username, t.Password)
	return t.BaseTransport.RoundTrip(req)
}

func NewClient(protocol string, host string, port string, username string, password string, insecure bool) *Client {
	configuration := &libregraph.Configuration{
		Servers: libregraph.ServerConfigurations{
			{
				URL: fmt.Sprintf("%s://%s:%s%s", protocol, host, port, "/graph"),
			},
		},
	}

	if insecure == true {
		configuration.HTTPClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
	}

	ctx := context.WithValue(context.Background(), libregraph.ContextBasicAuth, libregraph.BasicAuth{
		UserName: username,
		Password: password,
	})

	graphApiClient := libregraph.NewAPIClient(configuration)

	webdavApiClient := webdav.NewClient(fmt.Sprintf("%s://%s:%s%s", protocol, host, port, "/dav"), username, password)

	baseTransport := &http.Transport{}
	if insecure == true {
		baseTransport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	authTransport := &AuthTransport{
		BaseTransport: baseTransport,
		Username: username,
		Password: password,
	}
	webdavApiClient.SetTransport(authTransport)

	return &Client{
		Protocol: protocol,
		Host: host,
		Auth: AuthCredentials{
			Username: username,
			Password: password,
		},
		GraphAPICtx: ctx,
		GraphAPI: graphApiClient,
		WebdavAPI: webdavApiClient,
	}
}
