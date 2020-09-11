package util

import (
	"errors"
	"fmt"
	xhttp "github.com/zwdgithub/simple_http"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"net/http"
)

type Headers map[string]string

func defaultCustomProxy(client *http.Client) *http.Client {
	client.Transport = &http.Transport{}
	return client
}

func Get(url, encoding string, p ...interface{}) (string, error) {
	for i := 0; i <= 3; i++ {
		h := xhttp.NewHttpUtil()
		h.Get(url)
		for _, item := range p {
			switch v := item.(type) {
			case Headers:
				h.SetHeader(v)
			case func(c *http.Client) *http.Client:
				h.CustomClient(v)
			}
		}
		h.Do()

		if h.Error() != nil {
			return "", h.Error()
		}
		response := h.Response()
		defer response.Body.Close()
		var reader io.Reader
		reader = response.Body
		if encoding == EncodingGBK {
			reader = transform.NewReader(response.Body, simplifiedchinese.GBK.NewDecoder())
		}
		content, err := ioutil.ReadAll(reader)
		if err != nil {
			continue
		}
		return string(content), nil
	}
	return "", errors.New(fmt.Sprintf("http get %s error ", url))
}
