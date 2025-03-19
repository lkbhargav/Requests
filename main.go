package requests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	netUrl "net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
	Redirects  []string
	Response   map[string]interface{}
	StatusCode int
	Raw        []byte
}

// Request => Object required to make HTTP calls using Do method
type Request struct {
	Cookies              map[string]string
	CutsomRedirectMethod func(req *http.Request, via []*http.Request) error
	FormData             map[string]string
	Headers              map[string]string
	IsJSONResponse       bool
	JSONBody             map[string]interface{}
	Method               string
	QueryStrings         map[string]string
	ResponseStruct       interface{}
	RequestTimeout       time.Duration
	URL                  string
}

var structName string = "Structure"
var temporaryHolder map[string][]string
var counter int

func init() {
	temporaryHolder = make(map[string][]string)
	counter = 0
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
	var response map[string]interface{}

	if r.QueryStrings != nil {
		for k, v := range r.QueryStrings {
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

			for k, v := range r.FormData {
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
		for k, v := range r.Headers {
			req.Header.Set(k, v)
		}
	}

	if len(r.JSONBody) > 0 {
		req.Header.Set("Content-type", "application/json")
	}

	if r.Cookies != nil {
		for k, v := range r.Cookies {
			cookie := http.Cookie{Name: k, Value: v}
			req.AddCookie(&cookie)
		}
	}

	var timeout time.Duration = 10
	if r.RequestTimeout != time.Duration(0*time.Second) {
		timeout = r.RequestTimeout
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{Error: err}
	}

	if r.IsJSONResponse {
		err := json.Unmarshal(body, &response)
		if err != nil {
			return Response{Error: err}
		}
	} else {
		tmp := make(map[string]interface{})

		err := json.Unmarshal(body, &tmp)

		if err != nil {
			response = map[string]interface{}{"response": string(body)}
		} else {
			response = map[string]interface{}{"response": tmp}
		}
	}

	if r.ResponseStruct != nil {
		err := parse(body, r.ResponseStruct)
		if err != nil {
			return Response{Error: err}
		}
	}

	return Response{Response: response, Raw: body, StatusCode: resp.StatusCode, Error: nil, Redirects: redirects}
}

func parse(data []byte, obj interface{}) error {
	err := json.Unmarshal(data, obj)

	if err != nil {
		return err
	}

	return nil
}

// GenerateStructure => Takes a map[string]interface{} and prints out the structures
func GenerateStructure(inpMap map[string]interface{}) {
	counter++
	strCounter := strconv.Itoa(counter)

	tempArr := temporaryHolder[structName+strCounter]
	temporaryHolder[structName+strCounter] = append(tempArr, "type "+structName+"_"+strCounter+" struct {")

	for k, v := range inpMap {
		caser := cases.Title(language.English)
		key := caser.String(k)
		jsonTag := "`json:\"" + k + "\"`"
		if strings.HasPrefix(fmt.Sprintf("%v", reflect.TypeOf(v)), "map[string]") {
			tempArr = temporaryHolder[structName+strCounter]
			temporaryHolder[structName+strCounter] = append(tempArr, "  "+key+"  "+structName+"_"+strconv.Itoa(counter+1)+"  "+jsonTag)
			GenerateStructure(v.(map[string]interface{}))
		} else if strings.HasPrefix(fmt.Sprintf("%v", reflect.TypeOf(v)), "[]map[string]") {
			tempArr = temporaryHolder[structName+strCounter]
			temporaryHolder[structName+strCounter] = append(tempArr, "  "+key+"  []"+structName+"_"+strconv.Itoa(counter+1)+"  "+jsonTag)
			GenerateStructure(v.([]map[string]interface{})[0])
		} else {
			tempArr = temporaryHolder[structName+strCounter]
			temporaryHolder[structName+strCounter] = append(tempArr, "  "+key+"  "+fmt.Sprintf("%v", reflect.TypeOf(v))+"  "+jsonTag)
		}
	}

	tempArr = temporaryHolder[structName+strCounter]
	temporaryHolder[structName+strCounter] = append(tempArr, "}")

	for _, i := range temporaryHolder[structName+strCounter] {
		fmt.Println(i)
	}
}

// ConvertMap => Converts an interface{} (currently implemented only for map[string]string & defaulted to map[string]interface{}) to map[string]interface{}
func ConvertMap(req interface{}) (resp map[string]interface{}) {
	resp = make(map[string]interface{})

	switch fmt.Sprintf("%v", reflect.TypeOf(req)) {
	case "map[string]string":
		for k, v := range req.(map[string]string) {
			resp[k] = v
		}
	default:
		return req.(map[string]interface{})
	}

	return
}
