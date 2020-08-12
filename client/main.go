package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/pretty"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	initialBackoffMS  = 1000
	maxBackoffMS      = 32000
	backoffFactor     = 2
	rateLimitHttpCode = 429
	limit             = "1000"
)

type OktaClient struct {
	Domain      string
	Token       string
	httpClient  *http.Client
}

func NewClient(domain, token string) *OktaClient {
	return &OktaClient{
		Domain: domain,
		Token:  token,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (oktaClient *OktaClient) GetLogs(startTime string, endTime string, resultsChannel chan<- string) (int, error) {
	// Setup variables
	var events []string
	var tmpEventsRaw []interface{}
	count := 0
	afterLink := ""

	// Setup request
	params := url.Values{}
	params.Set("limit", limit)
	params.Set("since", startTime)
	params.Set("until", endTime)

	// Call request
	response, body, err := oktaClient.conductRequest("GET", "/api/v1/logs", params)

	// Handle HTTP error
	if err != nil {
		return -1, errors.New(fmt.Sprintf("Error conducting request: %v\n", err))
	}

	// Convert from JSON
	err = json.Unmarshal(body, &tmpEventsRaw)

	// Handle error
	if err != nil {
		return -1, errors.New(fmt.Sprintf("Error unmarshalling response body: %v\n", err))
	}

	// Convert to strings
	events, err = convertLogsToString(tmpEventsRaw)

	// Handle error
	if err != nil {
		return -1, errors.New(fmt.Sprintf("Error converting logs to strings: %v\n", err))
	}

	if len(events) == 0 {
		return 0, nil
	} else {
		count += len(events)
		for _, event := range events {
			// Ugly print the json into a single lined string
			resultsChannel <- string(pretty.Ugly([]byte(event)))
		}
	}

	// Get results offset
	afterLink = getResultsOffset(response)

	// Handle paged responses
	for afterLink != "" {
		// Clear variables
		tmpEventsRaw = nil
		events = nil

		// Set next link
		params.Set("next", afterLink)

		// Call request
		response, body, err = oktaClient.conductRequest("GET", "/api/v1/logs", params)

		// Handle error
		if err != nil {
			return -1, errors.New(fmt.Sprintf("Error conducting request: %v\n", err))
		}

		// Convert from JSON
		err = json.Unmarshal(body, &tmpEventsRaw)

		// Handle error
		if err != nil {
			return -1, errors.New(fmt.Sprintf("Error unmarshalling response body: %v\n", err))
		}

		// Convert to strings
		events, err = convertLogsToString(tmpEventsRaw)

		// Handle error
		if err != nil {
			return -1, errors.New(fmt.Sprintf("Error converting logs to strings: %v\n", err))
		}

		count += len(events)
		for _, event := range events {
			// Ugly print the json into a single lined string
			resultsChannel <- string(pretty.Ugly([]byte(event)))
		}

		afterLink = getResultsOffset(response)
	}

	return count, nil
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

	fmt.Printf("Calling URL: %s\n", urlObj.String())

	// Setup headers
	headers := make(map[string]string)
	headers["Accept"] = "application/json"
	headers["Authorization"] = fmt.Sprintf("SSWS %s", oktaClient.Token)
	headers["Content-Type"] = "application/json"

	// JSON marshal body
	var requestBody io.ReadCloser = nil
	if method == "POST" || method == "PUT" {
		// Marshal JSON
		bodyString, _ := json.Marshal(params)
		requestBody = ioutil.NopCloser(strings.NewReader(string(bodyString)))
	}

	response, body, err := oktaClient.makeRetryableHttpCall(method, urlObj, headers, requestBody)

	if err != nil {
		return nil, nil, err
	}

	return response, body, nil
}

func (oktaClient *OktaClient) makeRetryableHttpCall(
	method string,
	url url.URL,
	headers map[string]string,
	body io.ReadCloser,
) (*http.Response, []byte, error) {
	backoffMs := initialBackoffMS
	for {
		request, err := http.NewRequest(method, url.String(), nil)
		if err != nil {
			return nil, nil, err
		}

		if headers != nil {
			for k, v := range headers {
				request.Header.Set(k, v)
			}
		}
		if body != nil {
			request.Body = body
		}

		resp, err := oktaClient.httpClient.Do(request)
		var body []byte
		if err != nil {
			return resp, body, err
		}

		if backoffMs > maxBackoffMS || resp.StatusCode != rateLimitHttpCode {
			body, err = ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			return resp, body, err
		}

		time.Sleep(time.Millisecond * time.Duration(backoffMs))
		backoffMs *= backoffFactor
	}
}
