package requests

import "testing"

func runThroughAllCommonConditions(resp Response, t *testing.T, expectedStatusCode int, expectedNumberOfRedirects int) {
	if resp.Error != nil {
		t.Errorf("Error wasn't expected. Message: %s", resp.Error.Error())
	}

	if len(*resp.Redirects) != expectedNumberOfRedirects {
		t.Errorf("Redirects weren't expected. Redirects: %v", *resp.Redirects)
	}

	if resp.StatusCode != expectedStatusCode {
		t.Errorf("Expected status code 200, but found %d", resp.StatusCode)
	}

	if resp.Response == nil {
		t.Errorf("Expected Response to be non nil, but found nil")
	}
}

func TestSimpleHTTPGETRequestWithoutMethod(t *testing.T) {
	resp := Request{
		IsJSONResponse: true,
		URL:            "https://accounts.bgalytics.com",
	}.Do()

	runThroughAllCommonConditions(resp, t, 200, 0)
}

func TestSimpleGETRequestWithMethod(t *testing.T) {
	resp := Request{
		IsJSONResponse: true,
		URL:            "https://accounts.bgalytics.com",
		Method:         GET,
	}.Do()

	runThroughAllCommonConditions(resp, t, 200, 0)
}

// func TestSimplePOSTRequest(t *testing.T) {
// 	expectedResponse := "This is expected to be sent back as part of response body."

// 	resp := Request{
// 		URL:    "https://postman-echo.com/post",
// 		Method: POST,
// 	}.Do()

// 	runThroughAllCommonConditions(resp, t, 200, 0)

// 	mapResp := (*resp.Response)["response"].(string)

// 	if mapResp != expectedResponse {
// 		t.Errorf("Expected \"%s\", but found %s", expectedResponse, mapResp)
// 	}
// }
