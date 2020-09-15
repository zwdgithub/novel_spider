package util

import (
	"errors"
	"fmt"
	xhttp "github.com/zwdgithub/simple_http"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"
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
			time.Sleep(time.Second * 5)
			continue
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

func GetWithProxy(url, encoding string, p ...interface{}) (string, error) {
	for i := 0; i <= 3; i++ {
		h := xhttp.NewHttpUtil()
		h.Get("http://localhost:8092/get?url=" + url)
		//h.Get(url)
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
			time.Sleep(time.Second * 5)
			continue
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
			time.Sleep(time.Second * 1)
			continue
		}
		return string(content), nil
	}
	return "", errors.New(fmt.Sprintf("http get %s error ", url))
}

func LoadYaml(fileName string, dst interface{}) (interface{}, error) {
	t := reflect.TypeOf(dst)
	if t.Kind() == reflect.Ptr { //指针类型获取真正type需要调用Elem
		t = t.Elem()
	}
	newItem := reflect.New(t).Interface()
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(file, newItem)
	if err != nil {
		return nil, err
	}
	return newItem, nil
}

func ValidChapterName(chapterName string) bool {
	if strings.Contains(chapterName, "请假") {
		return false
	}
	reg := regexp.MustCompile("[\\d+一二三四五六七八九十百千第章节]")
	c := reg.FindString(chapterName)
	return len(c) > 0
}

func Trim(s string) string {
	s = strings.Trim(s, " ")
	s = strings.Trim(s, "\r")
	s = strings.Trim(s, "\n")
	s = strings.Trim(s, "\t")
	return s
}
