// test contains convenience types and functions used in BDD features
package test

import (
	"encoding/json"
	"fmt"
	"github.com/ddliu/go-httpclient"
	"io/ioutil"
	"log"
	"strings"
)

// Client is a convenience api so that we can run
// queries against an existing server and read
// responses from it. This is currently used in BDD scenarios
type Client struct {
	ServerUrl string
	http      *httpclient.HttpClient
	Resp      *httpclient.Response
	Json      map[string]interface{}
	Text      string
	Err       error
}

// NewClient returns a new HTTP client for the given
// server
func NewClient(serverUrl string) *Client {
	return &Client{
		http:      httpclient.NewHttpClient(),
		ServerUrl: serverUrl,
	}
}

// UrlFor builds a url for the given path
// using the client's internal server configuration.
func (c *Client) UrlFor(path string) string {
	return fmt.Sprintf("%s%s", c.ServerUrl, path)
}

// Get performs a GET request on the given path
// and updates its last response record
func (c *Client) Get(path string) {
	url := c.UrlFor(path)
	res, err := c.http.Get(url)
	c.Resp = res
	c.Err = err
	c.parseResponse()
}

// Get performs a DELETE request on the given path
// and updates its last response record
func (c *Client) Delete(path string) {
	url := c.UrlFor(path)
	res, err := c.http.Delete(url)
	c.Resp = res
	c.Err = err
	c.parseResponse()
}

// Post performs a POST request on the given path,
// with the given payload as json
func (c *Client) Post(path string, data string) {
	url := c.UrlFor(path)
	res, err := c.http.PostJson(url, data)
	c.Resp = res
	c.Err = err
	c.parseResponse()
}

// Post performs a PUT request on the given path,
// with the given payload as json
func (c *Client) Put(path string, data string) {
	url := c.UrlFor(path)
	res, err := c.http.PutJson(url, data)
	c.Resp = res
	c.Err = err
	c.parseResponse()
}

// parseResponse attempts to unmarshall the latest
// response to either generic map (json) or simple tesxt. This will
// initialize the Json and Textfields in the client's last response
// so that they can be consumed in tests in a convenient way
func (c *Client) parseResponse() {

	// If there was a connection issue, prevent
	// from accessing variables that might not even
	// be initialized
	if c.Err != nil {
		return
	}

	if c.Resp != nil && c.Resp.Body != nil {
		bytes, err := ioutil.ReadAll(c.Resp.Body)
		if err != nil {
			log.Printf("Could not read bytes from http response")
		} else {
			if c.HasJson() {
				c.maybeParseJson(bytes)
			} else if c.HasText() {
				c.maybeParseText(bytes)
			}
		}
	}
}

// maybeParseJson attempts to parse the response as json
func (c *Client) maybeParseJson(bytes []byte) {
	var anyJson map[string]interface{}
	err := json.Unmarshal(bytes, &anyJson)
	if err != nil {
		log.Printf("Could not unmarshall json: %v", string(bytes))
	} else {
		c.Json = anyJson
	}
}

// maybeParseText attempts to parse the response as json
func (c *Client) maybeParseText(bytes []byte) {
	c.Text = string(bytes)
}

// HasJson is a convenience function that returns true
// if the client has a json content-type in its last response
func (c *Client) HasJson() bool {
	return c.Resp != nil && strings.Contains(c.Resp.Header.Get("content-type"), "application/json")
}

// HasText is a convenience function that returns true
// if the client has a text/plain content-type in its last response
func (c *Client) HasText() bool {
	return c.Resp != nil && strings.Contains(c.Resp.Header.Get("content-type"), "text/plain")
}
