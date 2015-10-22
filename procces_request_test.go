package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// Test server , which overrides http.DefaultClient
func testServer(code int, body string) *httptest.Server {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, body)
	}))
	tr := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}
	httpClient := &http.Client{Transport: tr}

	http.DefaultClient = httpClient
	return server
}

func TestHasMergeVariables(t *testing.T) {
	testString := "I have a {{mergevariable}}"
	assert.True(t, hasMergeVariables(testString), "Expected to have merge variables")

	testString = "I don't have a mergevariables"
	assert.False(t, hasMergeVariables(testString), "Expected not to have merge variables")
}

func TestMerge(t *testing.T) {
	mergeValues := map[string]string{"key": "value"}
	testString := "I have a merge variable {{.key}}"
	result, _ := merge(testString, mergeValues)
	assert.Equal(t, "I have a merge variable value", result, "Expected variables to be merged")
}

func TestProcessingRequest(t *testing.T) {

	server := testServer(200, "{\"jsonkey\":\"jsonvalue\"}")
	defer server.Close()

	var mergeValues = map[string]string{
		"query": "test",
		"host":  server.URL,
	}

	httpText := `GET /search?hl=en&output=search&sclient=psy-ab&q={{.query}}&btnK= HTTP/1.1
				Host: {{.host}}
				User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_6_8) AppleWebKit/534.57.2 (KHTML, like Gecko) Version/5.1.7`

	str, err := ProcessRequest(httpText, mergeValues, nil, nil)
	assert.Nil(t, err, "Error object must be nil")
	assert.Equal(t, str, "{\"jsonkey\":\"jsonvalue\"}\n")
}

func TestExtractHeaders(t *testing.T) {
	httpText := `GET /search?hl=en&output=search&sclient=psy-ab&q={{query}}&btnK= HTTP/1.1
				Host: localhost:3000
				User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_6_8) AppleWebKit/534.57.2 (KHTML, like Gecko) Version/5.1.7`
	var host, _, headers, _ = extractHeaders(httpText)
	assert.NotNil(t, host, "Host cannot be nil")
	assert.NotEqual(t, "", host, "Host cannot be empty")
	assert.NotNil(t, headers, "Headers should not be nil")
	assert.Equal(t, 2, len(headers), "Invalid header length")

}

func TestExtractRequestMethod(t *testing.T) {
	httpText := `GET /search?hl=en&output=search&sclient=psy-ab&q={{query}}&btnK= HTTP/1.1
				Host: www.google.com
				User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_6_8) AppleWebKit/534.57.2 (KHTML, like Gecko) Version/5.1.7`
	method, err := extractRequestMethod(httpText)
	assert.Nil(t, err, "Error should be nil")
	assert.Equal(t, "GET", method, "Request method must match")

	//test with invalid request method
	httpText = `ILLEGAL /search?hl=en&output=search&sclient=psy-ab&q={{query}}&btnK= HTTP/1.1
				Host: www.google.com
				User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_6_8) AppleWebKit/534.57.2 (KHTML, like Gecko) Version/5.1.7`
	method, err = extractRequestMethod(httpText)
	assert.NotNil(t, err, "Error should not be nil")
	assert.Equal(t, "", method, "Request method must match")
}

func TestGetOptions(t *testing.T) {
	var options = map[string]string{
		"https": "true",
	}
	var result = GetOptions(options)
	assert.NotNil(t, result, "Not nil")
	assert.Equal(t, result["https"], "true", "Test equality of overiding value")

	//check default values
	result = GetOptions(nil)
	assert.NotNil(t, result, "Not nil")
	assert.Equal(t, result["https"], "false", "Test equality of default value")
}

func TestGetBasePath(t *testing.T) {
	var url = "www.nu.nl"
	var result = getBasePath(url)
	assert.Equal(t, result, "www.nu.nl/", "Test if trailing / is added")

	//test if / is not added when already ending with /
	url = "www.nu.nl"
	result = getBasePath(url)
	assert.Equal(t, result, "www.nu.nl/", "Test if trailing / is not added")
}

//Test helper methods
func HttpTemplateText() string {
	dat, err := ioutil.ReadFile("./test_request.http")
	if err != nil {
		fmt.Println("error")
	}
	return string(dat)
}
