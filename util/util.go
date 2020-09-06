package util

import (
	xhttp "github.com/zwdgithub/simple_http"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
)

func Get(url string, headers map[string]string, encoding string) (string, error) {
	h := xhttp.NewHttpUtil()
	h.Get(url).SetHeader(headers).Do()

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
		return "", nil
	}
	// fmt.Println(string(content))
	return string(content), nil
}
