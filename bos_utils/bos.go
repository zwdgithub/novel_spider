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
)

const chapterNameFmt = "/files/article/txt/%d/%d/%d.txt"
const coverNameFmt = "/files/article/image/%d/%d/%ds.jpg"

type BosUtil struct {
	bos    *bos.Client
	bucket string
}

func NewBosClient() *BosUtil {
	// 用户的Access Key ID和Secret Access Key
	AK, SK := "", ""
	ENDPOINT := "hkg.bcebos.com"
	// 初始化一个BosClient
	bosClient, err := bos.NewClient(AK, SK, ENDPOINT)
	if err != nil {
		l.Fatal("bos init error ", err)
	}

	return &BosUtil{
		bos:    bosClient,
		bucket: "testxxfile",
	}
}

func (b *BosUtil) PutChapter(aid, cid int, content string) error {
	return nil
	enc := mahonia.NewEncoder("gbk")
	content = enc.ConvertString(content)
	objName := fmt.Sprintf(chapterNameFmt, aid/1000, aid, cid)
	r, err := b.bos.PutObjectFromString(b.bucket, objName, content, nil)
	err = nil
	fmt.Println(r)
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
