package httptemplate

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"path"
	"regexp"
	"strings"
	"text/template"
)

var defaults = map[string]string{
	"https":             "false",
	"autoContentLength": "true",
}

var options = map[string]string{}

var validRequestMethods = map[string]bool{
	"GET":    true,
	"PUT":    true,
	"DELETE": true,
	"POST":   true,
}

type RequestCallback func(resp *http.Response, err error) error

func GetOptions(options map[string]string) map[string]string {
	m := map[string]string{}
	//set default settings
	for k, v := range defaults {
		m[k] = v
	}
	if options == nil {
		return m
	} else {
		//merge extra options
		for k, v := range options {
			m[k] = v
		}
		return m
	}
}

func getBasePath(basePath string) string {
	bp := path.Clean(basePath)
	if !strings.HasSuffix(basePath, "/") {
		bp = bp + "/"
	}
	return bp
}

func extractRequestParameters(httpText string) string {
	r := regexp.MustCompile(`\?([^\s]+)`)
	lines := strings.Split(httpText, "\n")
	var matches = r.FindStringSubmatch(lines[0])
	var parameters = ""
	if len(matches) > 0 {
		parameters = matches[1]
	}

	return parameters
}

func extractPostParameters(httpText string) string {
	r := regexp.MustCompile(`\n\n+(.*)`)
	var matches = r.FindStringSubmatch(httpText)
	var parameters = ""
	if len(matches) > 0 {
		parameters = matches[1]
	}
	return parameters
}

func extractHeaders(httpText string) (string, string, map[string]string, bool) {
	headers := map[string]string{}
	host := ""
	body := ""
	inBody := false
	headerRegex := regexp.MustCompile(`([-A-Za-z]+):\s+(.+)$`)
	lines := strings.Split(httpText, "\n")
	//pop of the first line
	lines = lines[1:]
	for _, line := range lines {

		match := headerRegex.MatchString(line)
		if match {
			header := headerRegex.FindStringSubmatch(line)
			if strings.ToLower(header[1]) == strings.ToLower("HOST") {
				host = header[2]
			} else {
				headers[header[1]] = header[2]
			}
		} else {

		}
	}
	return host, body, headers, inBody
}

func extractRequestMethod(httpText string) (string, error) {
	r := regexp.MustCompile(`^([A-Za-z]+)\s+(/[0-9/A-Za-z_n?=\-%&{{}}+\.]+)\s+HTTP.+$`)
	lines := strings.Split(httpText, "\n")
	var matches = r.FindStringSubmatch(lines[0])
	var method = matches[1]
	if !validRequestMethods[method] {
		return "", errors.New("Invalid request method in template, must be one of GET,PUT, POST or DELETE")
	}
	return method, nil
}

func addHeadersToRequest(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		req.Header.Add(k, v)
	}
}

func addOauth1AHeaderToRequest(req *http.Request, host string, requestParameters string, bodyParameters string, requestMethod string, options map[string]string) {
	req.Header.Add("Authorization", SignHeaderForOauth1A(host, requestParameters, bodyParameters, requestMethod, options))
}

func extractPath(httpText string) (string, error) {
	lines := strings.Split(httpText, "\n")
	r := regexp.MustCompile(`^([A-Za-z]+)\s+(/[0-9/A-Za-z_n?=\-%&{{}}+\.]+)\s+HTTP.+$`)
	var matches = r.FindStringSubmatch(lines[0])
	return matches[2], nil
}

func hasMergeVariables(httpText string) bool {
	matched, _ := regexp.MatchString("{{.+}}", httpText)
	return matched
}

func merge(httpText string, mergeValues map[string]string) (string, error) {

	tmpl, err := template.New("request").Parse(httpText)
	if err != nil {
		return "", errors.New("Error parsing http template")
	}
	stringWriter := new(bytes.Buffer)
	err = tmpl.Execute(stringWriter, mergeValues)
	if err != nil {
		return "", errors.New("Error merging variables in http template")
	}
	return stringWriter.String(), nil
}

func ProcessRequest(httpText string, mergeValues map[string]string, options map[string]string, callback RequestCallback, debug bool) (string, error) {
	requestOptions := GetOptions(options)
	//merge variables in path
	if hasMergeVariables(httpText) {
		httpText, _ = merge(httpText, mergeValues)
	}

	var method, err = extractRequestMethod(httpText)
	if err != nil {
		return "", err
	}

	var path, pathError = extractPath(httpText)
	if pathError != nil {
		return "", pathError
	}

	var host, body, headers, inBody = extractHeaders(httpText)

	client := &http.Client{}
	//add scheme if not available
	if !strings.HasPrefix(host, "http") {
		if requestOptions["https"] == "true" {
			host = "https://" + host
		} else {
			host = "http://" + host
		}
	}

	req, requestError := http.NewRequest(method, host+path, nil)
	if inBody {
		req, requestError = http.NewRequest(method, host+path, strings.NewReader(body))
	}
	if requestError != nil {
		return "", requestError
	}
	addHeadersToRequest(req, headers)

	requestParameters := extractRequestParameters(httpText)
	bodyParameters := extractPostParameters(httpText)

	_, ok := options["oauth1a_consumer_key"]
	if ok {
		addOauth1AHeaderToRequest(req, host+path, requestParameters, bodyParameters, method, options)
	}

	if debug {
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			fmt.Println("Error in request" + err.Error())
		}
		fmt.Println(" ************************* REQUEST ****************************")
		fmt.Println(string(dump))
	}

	resp, err := client.Do(req)
	if resp.StatusCode != 200 {
		if debug {
			dump, err := httputil.DumpResponse(resp, true)
			if err != nil {
				fmt.Println("Error in dumping response" + err.Error())
			}
			fmt.Println(" ************************* RESPONSE ****************************")
			fmt.Println(string(dump))
		} else {
			defer resp.Body.Close()
			contents, _ := ioutil.ReadAll(resp.Body)
			return "", errors.New("Error in request statuscode " + string(resp.StatusCode) + " " + string(contents))
		}
	}
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if callback != nil {
		return "", callback(resp, err)
	}
	contents, responseError := ioutil.ReadAll(resp.Body)
	if responseError != nil {
		return "", responseError
	}

	return string(contents), nil
}
