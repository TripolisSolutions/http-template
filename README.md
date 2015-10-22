# http-template
Create http requests based on http templates

Explanation and Usage
---------------------

http-template is a simple piece of code that accepts a string containing raw http. Here's an example file (a google search):

    GET /search?hl=en&output=search&sclient=psy-ab&q={{.query}}&btnK= HTTP/1.1
    Host: www.google.com
    User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_6_8) AppleWebKit/534.57.2 (KHTML, like Gecko) Version/5.1.7 Safari/534.57.2
    Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
    Referer: http://www.google.com/
    Accept-Language: en-us
    Accept-Encoding: gzip, deflate
    Connection: keep-alive


notice the **{{.query}}** on the first line. This is how we can modify the request with dynamic data. One good use for this is inserting session IDs into a request:

    ...
    Cookie: {{.session_id}}; path=/;
    ...


```js
str, err := ProcessRequest(httpText, mergeValues, nil)
```

Arguments
---------

### httpText
the first argument is a string representation of a raw http request

### mergeValues
the second argument is a map[string] string containing merge variables used in the template

### options
The third element is a map[string] string where you can define values to override the default settings for the following:
    - https: boolean (default: false) - if set to true request will be sent using https
    - autoContentLength: boolean (default: true) - when true, httpTemplate will fill in the 'Content-Length' header for you.

### callback 
the fourth argument is a callbacko of the following type:
```js
type RequestCallback func(resp *http.Response, err error) error
```
