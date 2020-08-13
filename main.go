package requests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	netUrl "net/url"
	"strings"
	"time"
)

// GET => Used for GET request
var GET string = "GET"

// POST => Used for POST request
var POST string = "POST"

// PUT => Used for PUT request
var PUT string = "PUT"

// DELETE => Used for DELETE request
var DELETE string = "DELETE"

// PATCH => Used for PATCH request
var PATCH string = "PATCH"

// Response => Object that is returned on successful HTTP request
type Response struct {
	Error      error
	Redirects  *[]string
	Response   *map[string]interface{}
	StatusCode int
}

// Request => Object required to make HTTP calls using Do method
type Request struct {
	Cookies              *map[string]string
	CutsomRedirectMethod func(req *http.Request, via []*http.Request) error
	FormData             *map[string]string
	Headers              *map[string]string
	IsJSONResponse       bool
	JSONBody             map[string]interface{}
	Method               string
	QueryStrings         *map[string]string
	ResponseStruct       interface{}
	RequestTimeout       *time.Duration
	URL                  string
}

// Do => Simplified version of net/http Do function
func (r Request) Do() Response {
	if r.URL == "" {
		return Response{Error: errors.New("url cannot be empty")}
	}

	if r.Method == "" {
		r.Method = GET
	}

	var qs string = "?"
	var redirects []string
	var response *map[string]interface{}

	if r.QueryStrings != nil {
		for k, v := range *r.QueryStrings {
			qs += fmt.Sprintf("%s=%s&", k, v)
		}
	}

	if qs == "?" {
		qs = ""
	} else {
		qs = string(qs[0 : len(qs)-1])
	}

	var req *http.Request

	var err error

	var jsonBodyBytes []byte

	if len(r.JSONBody) > 0 {
		jsonBodyBytes, err = json.Marshal(r.JSONBody)

		if err != nil {
			return Response{Error: err}
		}
	}

	switch r.Method {
	case POST:
		if r.FormData != nil {
			form := netUrl.Values{}

			for k, v := range *r.FormData {
				form.Add(k, v)
			}

			req, err = http.NewRequest("POST", fmt.Sprintf("%s%s", r.URL, qs), strings.NewReader(form.Encode()))
		} else {
			data := jsonBodyBytes
			req, err = http.NewRequest(POST, fmt.Sprintf("%s%s", r.URL, qs), bytes.NewBuffer(data))
		}
	case GET:
		req, err = http.NewRequest(GET, fmt.Sprintf("%s%s", r.URL, qs), nil)
	case PUT:
		data := jsonBodyBytes
		req, err = http.NewRequest(PUT, fmt.Sprintf("%s%s", r.URL, qs), bytes.NewBuffer(data))
	case DELETE:
		data := jsonBodyBytes
		req, err = http.NewRequest(DELETE, fmt.Sprintf("%s%s", r.URL, qs), bytes.NewBuffer(data))
	case PATCH:
		data := jsonBodyBytes
		req, err = http.NewRequest(PATCH, fmt.Sprintf("%s%s", r.URL, qs), bytes.NewBuffer(data))
	default:
		return Response{Error: errors.New("invalid method name")}
	}

	if err != nil {
		return Response{Error: err}
	}

	if r.Headers != nil {
		for k, v := range *r.Headers {
			req.Header.Set(k, v)
		}
	}

	if len(r.JSONBody) > 0 {
		req.Header.Set("Content-type", "application/json")
	}

	if r.Cookies != nil {
		for k, v := range *r.Cookies {
			cookie := http.Cookie{Name: k, Value: v}
			req.AddCookie(&cookie)
		}
	}

	var timeout time.Duration = 10
	if r.RequestTimeout != nil {
		timeout = *r.RequestTimeout
	}

	client := &http.Client{Timeout: time.Second * timeout}

	if r.CutsomRedirectMethod != nil {
		client.CheckRedirect = r.CutsomRedirectMethod
	} else {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			redirects = append(redirects, req.URL.String())
			return nil
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return Response{Error: err}
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Response{Error: err}
	}

	if r.IsJSONResponse {
		err := json.Unmarshal(body, &response)
		if err != nil {
			return Response{Error: err}
		}
	} else {
		response = &map[string]interface{}{"response": fmt.Sprintf("%q", body)}
	}

	if r.ResponseStruct != nil {
		err := parse(body, r.ResponseStruct)
		if err != nil {
			return Response{Error: err}
		}
	}

	return Response{Response: response, StatusCode: resp.StatusCode, Error: nil, Redirects: &redirects}
}

// Stringify => Takes a struct/map and stringfies it, so it could be used to pass into Request struct for JSONBody
func Stringify(obj interface{}) (string, error) {
	out, err := json.Marshal(&obj)

	if err != nil {
		return "", err
	}

	return string(out), nil
}

func parse(data []byte, obj interface{}) error {
	err := json.Unmarshal(data, obj)

	if err != nil {
		return err
	}

	return nil
}

// ConvertStringToStringPointer => Converts a string to a string pointer
func ConvertStringToStringPointer(str string) (strPointer *string) {
	strPointer = &str
	return strPointer
}

// ConvertStringArrToArrOfStringPointers => Converts an array of strings to an array of string pointers
func ConvertStringArrToArrOfStringPointers(arr []string) (pointerString []*string) {
	for _, v := range arr {
		pointerString = append(pointerString, ConvertStringToStringPointer(v))
	}
	return pointerString
}
