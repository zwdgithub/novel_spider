package bos_utils

import (
	"fmt"
	"github.com/axgle/mahonia"
	"github.com/baidubce/bce-sdk-go/services/bos"
	"github.com/baidubce/bce-sdk-go/services/bos/api"
	xhttp "github.com/zwdgithub/simple_http"
	"io/ioutil"
	l "log"
	"novel_spider/log"
	"novel_spider/util"
	"strings"
)

const chapterNameFmt = "/files/article/txt/%d/%d/%d.txt"
const coverNameFmt = "/files/article/image/%d/%d/%ds.jpg"
const opfNameFmt = "/files/article/image/%d/%d/index.opf"

type BosUtil struct {
	bos    *bos.Client
	bucket string
}

func NewBosClient(fileName string) *BosUtil {
	var conf *BosConf
	r, err := util.LoadYaml(fileName, conf)
	if err != nil {
		l.Fatal("bos conf load yaml error")
	}
	conf = r.(*BosConf)
	log.Info(conf)
	bosClient, err := bos.NewClient(conf.AK, conf.SK, conf.Endpoint)
	if err != nil {
		l.Fatal("bos init error ", err)
	}

	return &BosUtil{
		bos:    bosClient,
		bucket: conf.Bucket,
	}
}

type BosConf struct {
	AK       string `yaml:"ak"`
	SK       string `yaml:"sk"`
	Endpoint string `yaml:"endpoint"`
	Bucket   string `yaml:"bucket"`
}

func (b *BosUtil) PutChapter(aid, cid int, content string) error {
	enc := mahonia.NewEncoder("gbk")
	content = enc.ConvertString(content)
	content = strings.ReplaceAll(content, "\x1A", "  ")
	objName := fmt.Sprintf(chapterNameFmt, aid/1000, aid, cid)
	_, err := b.bos.PutObjectFromString(b.bucket, objName, content, nil)
	return err
}

func (b *BosUtil) PutOpf(aid int, content string) error {
	enc := mahonia.NewEncoder("gbk")
	content = enc.ConvertString(content)
	objName := fmt.Sprintf(opfNameFmt, aid/1000, aid)
	_, err := b.bos.PutObjectFromString(b.bucket, objName, content, nil)
	return err
}

func (b *BosUtil) PutCover(url string, aid int) error {
	h := xhttp.NewHttpUtil()
	h.Get(url).Do()
	if h.Error() != nil {
		return h.Error()
	}
	resp := h.Response()
	defer resp.Body.Close()
	bb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	objName := fmt.Sprintf(coverNameFmt, aid/1000, aid, aid)
	log.Infof("bos put cover %s", objName)
	_, err = b.bos.PutObjectFromBytes(b.bucket, objName, bb, &api.PutObjectArgs{})

	if err != nil {
		return err
	}
	return nil
}

func (b *BosUtil) GetChapter(aid, cid int) ([]byte, error) {
	r, err := b.bos.BasicGetObject(b.bucket, fmt.Sprintf(chapterNameFmt, aid/1000, aid, cid))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
