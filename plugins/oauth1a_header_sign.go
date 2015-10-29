package plugins

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/mrjones/oauth"
	"math/rand"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type FormValue map[string]string

type Credentials struct {
	OauthConsumerKey     string
	OauthNonce           string
	OauthSignature       string
	OauthSignatureMethod string
	OauthTimestamp       int64
	OauthToken           string
	OauthVersion         string

	OauthConsumerSecret string
	OauthTokenSecret    string
}

func NewCredentials(consumer_key, token, consumer_secret, token_secret string) *Credentials {
	c := Credentials{
		OauthConsumerKey:     consumer_key,
		OauthToken:           token,
		OauthVersion:         "1.0",
		OauthSignatureMethod: "HMAC-SHA1",
		OauthNonce:           GenerateNonce(),
		OauthTimestamp:       time.Now().UTC().Unix(),

		OauthConsumerSecret: consumer_secret,
		OauthTokenSecret:    token_secret,
	}
	return &c
}

func GenerateNonce() string {
	var bytes [32]byte

	var b int64
	for i := 0; i != 32; i++ {
		b = rand.Int63n(62)
		switch {
		case b < 10:
			b += 48
		case b < 36:
			b += 55
		default:
			b += 61
		}
		bytes[i] = byte(b)
	}

	return string(bytes[:32])
}

func GenerateParameterString(str_url string, form_value FormValue, credentials *Credentials) string {
	// p_url, _ := url.Parse(str_url)
	v := map[string]string{}

	v["oauth_version"] = credentials.OauthVersion
	v["oauth_consumer_key"] = credentials.OauthConsumerKey
	v["oauth_nonce"] = credentials.OauthNonce
	v["oauth_signature_method"] = credentials.OauthSignatureMethod
	v["oauth_timestamp"] = string(strconv.FormatInt(credentials.OauthTimestamp, 10))
	v["oauth_token"] = credentials.OauthToken

	for key, val := range form_value {
		v[strings.TrimSpace(string(key))] = strings.TrimSpace(val)
	}

	var form_keys sort.StringSlice = make([]string, len(v))
	i := 0
	for key := range v {
		form_keys[i] = key
		i++
	}

	form_keys.Sort()

	var buffer bytes.Buffer
	for _, key := range form_keys {

		if buffer.Len() > 0 {
			buffer.WriteString("&")
		}

		buffer.WriteString(url.QueryEscape(key))
		buffer.WriteString("=")
		buffer.WriteString(strings.Replace(url.QueryEscape(v[string(key)]), "+", "%20", -1))
		//buffer.WriteString(url.QueryEscape(v[string(key)]))

	}
	result := buffer.String()

	return result
}

func GenerateSignatureBaseString(method, host, parameter_string string) string {
	var buffer bytes.Buffer

	buffer.WriteString(strings.ToUpper(method))
	buffer.WriteString("&")

	p_url, _ := url.Parse(host)
	var uri bytes.Buffer

	uri.WriteString(p_url.Scheme)
	uri.WriteString("://")
	uri.WriteString(p_url.Host)
	uri.WriteString(p_url.Path)

	buffer.WriteString(url.QueryEscape(uri.String()))
	buffer.WriteString("&")
	buffer.WriteString(url.QueryEscape(parameter_string))
	result := buffer.String()

	return result
}

func GenerateSigningKey(c *Credentials) string {
	var buffer bytes.Buffer

	buffer.WriteString(url.QueryEscape(c.OauthConsumerSecret))
	buffer.WriteString("&")
	buffer.WriteString(url.QueryEscape(c.OauthTokenSecret))

	result := buffer.String()

	return result
}

func GenerateSignature(signature_base_string, signing_key string) string {
	sha := sha1.New
	h := hmac.New(sha, []byte(signing_key))
	h.Write([]byte(signature_base_string))

	encoder := base64.StdEncoding
	result := encoder.EncodeToString(h.Sum(nil))

	return result
}

/*
/* Specific method for twitter header signing
*/
func SignHeaderForOauth1A(host string, requestParameters string, bodyParameters string, requestMethod string, options map[string]string) string {
	//NOTE: remove this line or turn off Debug if you don't
	//want to see what the headers look like
	//consumer.Debug(true)

	//Roll your own AccessToken struct
	accessToken := &oauth.AccessToken{Token: options["oauth1a_access_token"], Secret: options["oauth1a_access_token_secret"]}

	credentials := NewCredentials(options["oauth1a_consumer_key"], accessToken.Token, options["oauth1a_consumer_secret"], options["oauth1a_access_token_secret"])

	form_value := make(FormValue)

	m1, _ := url.ParseQuery(requestParameters)
	m2, _ := url.ParseQuery(bodyParameters)

	for k, v := range m1 {
		form_value[k] = v[0]
	}
	for k, v := range m2 {
		form_value[k] = v[0]
	}

	parameter_string := GenerateParameterString(host, form_value, credentials)
	signature_base_string := GenerateSignatureBaseString(requestMethod, host, parameter_string)
	signing_key := GenerateSigningKey(credentials)
	signature := GenerateSignature(signature_base_string, signing_key)

	credentials.OauthSignature = signature

	auth_header := fmt.Sprintf(`OAuth oauth_consumer_key="%s", oauth_nonce="%s", oauth_signature="%s", oauth_signature_method="%s", oauth_timestamp="%d", oauth_token="%s", oauth_version="%s"`,
		credentials.OauthConsumerKey,
		credentials.OauthNonce,
		url.QueryEscape(credentials.OauthSignature),
		credentials.OauthSignatureMethod,
		credentials.OauthTimestamp,
		accessToken.Token,
		credentials.OauthVersion)

	return auth_header
}
