package requests

import (
	"fmt"
	"reflect"
	"testing"
)

const testURL = "https://user-accounts-staging.nocturnal.health"

func runThroughAllCommonConditions(resp Response, t *testing.T, expectedStatusCode int, expectedNumberOfRedirects int) {
	if resp.Error != nil {
		t.Errorf("Error wasn't expected. Message: %s", resp.Error.Error())
	}

	if len(resp.Redirects) != expectedNumberOfRedirects {
		t.Errorf("Redirects weren't expected. Redirects: %v", resp.Redirects)
	}

	if resp.StatusCode != expectedStatusCode {
		t.Errorf("Expected status code 200, but found %d", resp.StatusCode)
	}

	if resp.Response == nil {
		t.Errorf("Expected Response to be non nil, but found nil")
	}
}

func TestSimpleHTTPGETRequestWithoutMethod(t *testing.T) {
	resp := Request[*int]{
		IsJSONResponse: true,
		URL:            testURL,
	}.Do()

	runThroughAllCommonConditions(resp, t, 200, 0)
}

func TestSimpleGETRequestWithMethod(t *testing.T) {
	resp := Request[*int]{
		IsJSONResponse: true,
		URL:            testURL,
		Method:         GET,
	}.Do()

	runThroughAllCommonConditions(resp, t, 200, 0)
}

func TestGETRequestResponseString(t *testing.T) {
	resp := Request[*int]{
		URL: "http://www.google.com",
	}.Do()

	if resp.Error != nil {
		t.Errorf("Error wasn't expected. Message: %s", resp.Error.Error())
	}

	if fmt.Sprintf("%v", reflect.TypeOf(resp.Response["response"])) != "string" {
		t.Errorf("Expected the response to be of type string. Instead found: %v", reflect.TypeOf(resp.Response["response"]))
	}
}

func TestGETRequestResponseJSON(t *testing.T) {
	resp := Request[*int]{
		URL: testURL,
	}.Do()

	if resp.Error != nil {
		t.Errorf("Error wasn't expected. Message: %s", resp.Error.Error())
	}

	if fmt.Sprintf("%v", reflect.TypeOf(resp.Response["response"])) != "map[string]interface {}" {
		t.Errorf("Expected the response to be of type map[string]interface {}. Instead found: %v", reflect.TypeOf(resp.Response["response"]))
	}
}
