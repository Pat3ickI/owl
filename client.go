package owl

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	netURL "net/url"
	"strings"
	"time"

	"golang.org/x/net/html/charset"
)

type Client struct {
	*http.Client
	Header         map[string]string
	Cookies        map[string]string
	RequestTimeout time.Duration
}

type Parameters struct {
	Header         map[string]string
	Cookies        map[string]string
	RequestTimeout time.Duration
	Timeout        time.Duration
	HttpClient     *http.Client
}

var DefaultParameters Parameters = Parameters{
	Header: map[string]string{
		"User-Agent":    "Owl Mozilla/5.0 Firefox/96.0",
		"Accept":        "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		"Cache-Control": "max-age=0",
	},
	RequestTimeout: 10 * time.Second,
	Timeout:        10 * time.Second,
}

func HttpClientWrapper(c *http.Client) *Client {
	return &Client{
		Client: c,
	}
}

func NewClient(para *Parameters) *Client {
	var (
		client Client
	)
	if para != nil {
		client.Header = DefaultParameters.Header
		client.Cookies = DefaultParameters.Cookies
		client.RequestTimeout = DefaultParameters.RequestTimeout
		client.Timeout = DefaultParameters.Timeout
	} else {
		client.Header = para.Header
		client.Cookies = para.Cookies
		client.RequestTimeout = para.RequestTimeout
	}
	if para.HttpClient != nil {
		client.Client = &http.Client{
			Timeout: client.Timeout,
		}
	}

	return &client
}
func (c *Client) Post(url string, contentType string, body interface{}) (io.Reader, error) {
	bodyReader, err := getBodyReader(body)
	if err != nil {
		return nil, err
	}
	c.Header = map[string]string{
		"Content-Type": contentType,
	}
	return buildRequest(c, url, "POST", bodyReader)

}

func (c *Client) Get(url string) (io.Reader, error) {
	return buildRequest(c, url, "GET", nil)
}

func buildRequest(c *Client, url string, method string, body io.Reader) (io.Reader, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.RequestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	setParameters(req, c)

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
}

func setParameters(req *http.Request, c *Client) {
	// For Headers
	for hname, hvalue := range c.Header {
		req.Header.Set(hname, hvalue)
	}
	//For Cookies
	for cname, cvalue := range c.Cookies {
		req.AddCookie(&http.Cookie{
			Name:  cname,
			Value: cvalue,
		})
	}
}

// getBodyReader serializes the body for a network request. See the test file for examples
func getBodyReader(rawBody interface{}) (io.Reader, error) {
	var bodyReader io.Reader

	if rawBody != nil {
		switch body := rawBody.(type) {
		case map[string]string:
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			bodyReader = bytes.NewBuffer(jsonBody)
		case netURL.Values:
			bodyReader = strings.NewReader(body.Encode())
		case []byte: //expects JSON format
			bodyReader = bytes.NewBuffer(body)
		case string: //expects JSON format
			bodyReader = strings.NewReader(body)
		default:
			return nil, errors.New("unable to determine the body type")
		}
	}

	return bodyReader, nil
}
