package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	//"net/http/httputil"
	"bytes"
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

func readTemplate(template string) {

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
			}
			headers[header[0]] = header[1]
		} else {

		}
	}
	return host, body, headers, inBody
}

func createRequest() bool {
	return true
}

func extractHost(httpText string) string {
	return ""
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

func HttpTemplateText() string {
	dat, err := ioutil.ReadFile("./test_request.http")
	if err != nil {
		fmt.Println("error")
	}
	return string(dat)
}

func addHeadersToRequest(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		req.Header.Add(k, v)
	}
}

func extractPath(httpText string) (string, error) {
	lines := strings.Split(httpText, "\n")
	r := regexp.MustCompile(`^([A-Za-z]+)\s+(/[0-9/A-Za-z_n?=\-%&{{}}+\.]+)\s+HTTP.+$`)
	var matches = r.FindStringSubmatch(lines[0])
	return matches[2], nil
}

func ProcessRequest(httpText string, mergeValues map[string]string, callback RequestCallback) (string, error) {

	//merge variables in path
	matched, _ := regexp.MatchString("{{.+}}", httpText)
	if matched {
		tmpl, err := template.New("request").Parse(httpText)
		if err != nil {
			return "", errors.New("Error parsing http template")
		}
		stringWriter := new(bytes.Buffer)
		err = tmpl.Execute(stringWriter, mergeValues)
		if err != nil {
			return "", errors.New("Error merging variables in http template")
		}
		httpText = stringWriter.String()
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
		if options["https"] == "true" {
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
	// var requestDump, requestDumpError = httputil.DumpRequestOut(req, false)
	// if requestDumpError != nil {
	// 	return "", requestDumpError
	// }
	// fmt.Println(string(requestDump))
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
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

// module.exports = function(basePath) {
// 	//make sure we have the trailing '/'
// 	basePath += (basePath[basePath.length-1]==='/' ? '' : '/');

// 	var defaults = defaults({Https: false, AutoContentLength: true})

// 	var get_options = function(options) {
// 		if (options) {
// 			return _.extend({}, defaults, options); // mix options into defaults, copy to new object
// 		} else {
// 			return _.extend({}, defaults); // no options passed, copy defaults
// 		}
// 	};

// 	return function(template, data, options, callback) {

// 		if (_.isFunction(options) && !callback) {
// 			callback = options;
// 			options = null;
// 		}

// 		options = get_options(options);

// 		// load template from file system
// 		var http_text = _.template(
// 			fs.readFileSync(basePath + template + '.http', 'utf-8').toString(), data, {
// 				interpolate: /\{\{(.+?)\}\}/g
// 			});

// 		//parse first line
// 		var lines = http_text.split(/\n/);
// 		var matches = lines[0].match(/^([A-Za-z]+)\s+([\/0-9A-Za-z_&?=\-%+\.]+)\s+HTTP.+$/);
// 		var method = matches[1];
// 		var path = matches[2];
// 		lines.shift();

// 		// read all the headers
// 		var host, body = '', headers = {}, inBody = false;
// 		_.each(lines, function(line) {

// 			if (line.length > 0) {

// 				var matches = line.match(/([-A-Za-z]+):\s+(.+)$/);
// 				if (matches) {

// 					var k = matches[1];
// 					var v = matches[2];

// 					if (k.toUpperCase() === "HOST") {
// 						host = v;
// 					}
// 					headers[k] = v;

// 				} else {
// 					inBody = true;
// 					body += line+'\n';
// 				}

// 			} else {
// 				if (inBody) {
// 					body += '\n';
// 				}
// 				return;
// 			}
// 		});

// 		// if enabled, use whatever the resulting body length is
// 		if (options.autoContentLength) {
// 			headers['Content-Length'] = body.length;
// 		}

// 		// send request
// 		var proto = options.https ? https : http;
// 		var req = proto.request({
// 			host: host,
// 			port: options.https ? '443' : '80',
// 			path: path,
// 			method: method,
// 			headers: headers },
// 			function(res) {
// 				var buffers = [];

// 				// set up pipe for response data, gzip or plain based on respone header
// 				var stream = res.headers['content-encoding'] && res.headers['content-encoding'].match(/gzip/) ?
// 					res.pipe(zlib.createGunzip()) : res;

// 				stream.on('data', function(chunk) {
// 					buffers.push(chunk);
// 				}).on('end', function() {
// 					callback(res, Buffer.concat(buffers).toString());
// 				});
// 		});
// 		req.write(body);
// 		req.end();
//   	};
// };
