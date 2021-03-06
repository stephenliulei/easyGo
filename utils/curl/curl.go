package curl

import (
	"bytes"
	"crypto/tls"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

// Curl http请求类
type Curl struct {
	Domain string
}

// Get http-get请求
func (c Curl) Get(uri string, params map[string]string) ([]byte, error) {
	var requestURL string
	if uri == "" {
		requestURL = c.Domain
	} else {
		requestURL = c.Domain + uri
	}
	if len(params) > 0 {
		parseURL, parseErr := url.Parse(requestURL)
		if parseErr != nil {
			return []byte{}, parseErr
		}
		q := parseURL.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		parseURL.RawQuery = q.Encode()
		requestURL = parseURL.String()
	}
	result, err := http.Get(requestURL)
	if err != nil {
		return []byte{}, err
	}
	resultByte, err1 := ioutil.ReadAll(result.Body)
	defer result.Body.Close()
	if err1 != nil {
		return []byte{}, err1
	}
	return resultByte, nil
}

// Post http的post请求
func (c Curl) Post(uri string, params url.Values) ([]byte, error) {
	requestURL := c.Domain + uri
	result, err := http.PostForm(requestURL, params)
	if err != nil {
		return []byte{}, err
	}
	defer result.Body.Close()
	resultByte, err1 := ioutil.ReadAll(result.Body)
	if err1 != nil {
		return []byte{}, err1
	}
	return resultByte, nil
}

// StreamPost http的post参数放流里请求
func (c Curl) StreamPost(uri string, params []byte, headers map[string]string) ([]byte, error) {
	requestURL := c.Domain + uri
	body := bytes.NewReader(params)
	request, err := http.NewRequest("POST", requestURL, body)
	if err != nil {
		return []byte{}, err
	}

	if len(headers) > 0 {
		for k, v := range headers {
			if k == "HOST" {
				request.Host = v
			} else {
				request.Header.Set(k, v)
			}
		}
	}

	var resp *http.Response
	resp, err = http.DefaultClient.Do(request)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	var resultByte []byte
	resultByte, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return resultByte, nil
}

// FlexibleClient 带header头的post请求
func (c Curl) FlexibleClient(uri string, method string, data url.Values, headers map[string]string) ([]byte, error) {
	requestURL := c.Domain + uri
	client := &http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	req, err := http.NewRequest(strings.ToUpper(method), requestURL, strings.NewReader(data.Encode()))
	if len(headers) > 0 {
		for k, v := range headers {
			if k == "HOST" {
				req.Host = v
			} else {
				req.Header.Set(k, v)
			}
		}
	}
	res, _ := client.Do(req)
	if err != nil {
		return []byte{}, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	return body, err
}

func (c Curl) PostFormFile(uri string, fields map[string]string, fname, fpath string) ([]byte, error) {

	requestURL := c.Domain + uri

	var err error
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if fields != nil {
		for k, v := range fields {
			if err = writer.WriteField(k, v); err != nil {
				return nil, err
			}
		}
	}

	if fname != "" && fpath != "" {
		file, err := os.Open(fpath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		part, err := writer.CreateFormFile(fname, path.Base(fpath))
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(part, file)
		if err != nil {
			return nil, err
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", requestURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
