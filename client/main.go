package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"github.com/tidwall/pretty"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Constants for HTTP client
const (
	initialBackoffMS  = 1000
	maxBackoffMS      = 32000
	backoffFactor     = 2
	rateLimitHttpCode = 429
	limit             = "1000"
)

// Okta client struct
type OktaClient struct {
	Domain      string
	Token       string
	httpClient  *http.Client
}

// Create a new client with the okta domain and token and a http client with a 10 seconds timeout
func NewClient(domain, token string) *OktaClient {
	return &OktaClient{
		Domain: domain,
		Token:  token,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

// Get logs method with paged results logic
// Events are streamed into the results channel
func (oktaClient *OktaClient) GetLogs(startTime string, endTime string, resultsChannel chan<- string) (int, error) {
	// Setup variables
	count := 0
	afterLink := ""
	hasNext := true

	// Setup request
	params := url.Values{}
	params.Set("limit", limit)
	params.Set("since", startTime)
	params.Set("until", endTime)

	// Handle paged responses
	for hasNext {
		// Get logs
		events, newAfterLink, err := oktaClient.getLogsRequest(params, afterLink)

		// Handle error
		if err != nil {
			return -1, err
		}

		// Send events to channel
		for _, event := range events {
			// Ugly print the json into a single lined string
			resultsChannel <- string(pretty.Ugly([]byte(event)))
		}

		// Increment count
		count += len(events)

		// Set afterLink
		hasNext = newAfterLink != ""
		afterLink = newAfterLink
	}

	return count, nil
}

// Individual get logs request method
func (oktaClient *OktaClient) getLogsRequest(params url.Values, afterLink string) ([]string, string, error) {
	// Set variables
	var events []string
	var tmpEventsRaw []interface{}

	// Set next link
	if afterLink != "" {
		params.Set("after", afterLink)
	}

	// Call request
	response, body, err := oktaClient.conductRequest("GET", "/api/v1/logs", params)

	// Handle error
	if err != nil {
		return nil, "", errors.New(fmt.Sprintf("Error conducting request: %v\n", err))
	}

	// Convert from JSON
	err = json.Unmarshal(body, &tmpEventsRaw)

	// Handle error
	if err != nil {
		return nil, "", errors.New(fmt.Sprintf("Error unmarshalling response body: %v\n", err))
	}

	// Convert to strings
	events, err = convertLogsToString(tmpEventsRaw)

	// Handle error
	if err != nil {
		return nil, "", errors.New(fmt.Sprintf("Error converting logs to strings: %v\n", err))
	}

	// Get next page of results
	newAfterLink := getResultsOffset(response)

	return events, newAfterLink, nil
}

// Make an Okta API call.
// method is POST or GET
// uri is the URI of the Okta Rest call
// params HTTP query parameters to include in the call.
//
// Example: oktaClient.CallRequest("GET", "/auth/v2/check", nil)
func (oktaClient *OktaClient) conductRequest(method string, uri string, params url.Values) (*http.Response, []byte, error) {
	// Build the URL
	urlObj := url.URL{
		Scheme: "https",
		Host:   oktaClient.Domain,
		Path:   uri,
	}

	// Convert method to uppercase
	method = strings.ToUpper(method)

	// Encode params if GET request
	if method == "GET" {
		urlObj.RawQuery = params.Encode()
	}

	// Log for debugging
	if viper.GetBool("verbose") {
		fmt.Printf("Calling URL: %s\n", urlObj.String())
	}

	// Setup headers
	headers := make(map[string]string)
	headers["Accept"] = "application/json"
	headers["Authorization"] = fmt.Sprintf("SSWS %s", oktaClient.Token)
	headers["Content-Type"] = "application/json"

	// JSON marshal body if POST or PUT
	var requestBody io.ReadCloser = nil
	if method == "POST" || method == "PUT" {
		// Marshal JSON
		bodyString, _ := json.Marshal(params)
		requestBody = ioutil.NopCloser(strings.NewReader(string(bodyString)))
	}

	// Make a retryable HTTP call
	response, body, err := oktaClient.makeRetryableHttpCall(method, urlObj, headers, requestBody)

	// Handle error
	if err != nil {
		return nil, nil, err
	}

	return response, body, nil
}

// Make a retryable HTTP call. Supports APIs that return a 429 for too many requests
func (oktaClient *OktaClient) makeRetryableHttpCall(
	method string,
	url url.URL,
	headers map[string]string,
	body io.ReadCloser,
) (*http.Response, []byte, error) {
	backoffMs := initialBackoffMS
	for {
		// Setup new request
		request, err := http.NewRequest(method, url.String(), nil)

		// Handle error
		if err != nil {
			return nil, nil, err
		}

		// Setup headers
		if headers != nil {
			for k, v := range headers {
				request.Header.Set(k, v)
			}
		}

		// Setup body
		if body != nil {
			request.Body = body
		}

		// Conduct request
		resp, err := oktaClient.httpClient.Do(request)
		var body []byte

		// Handle error or failed response status code
		if err != nil || (resp.StatusCode != 200 && resp.StatusCode != rateLimitHttpCode) {
			if err == nil {
				return resp, body, errors.New(fmt.Sprintf("HTTP response code: %v\n", resp.Status))
			}
			return resp, body, err
		}

		// Handle rate limit code
		if backoffMs > maxBackoffMS || resp.StatusCode != rateLimitHttpCode {
			body, err = ioutil.ReadAll(resp.Body)
			_ = resp.Body.Close()
			return resp, body, err
		}

		time.Sleep(time.Millisecond * time.Duration(backoffMs))
		backoffMs *= backoffFactor
	}
}
