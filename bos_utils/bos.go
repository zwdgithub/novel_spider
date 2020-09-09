package bos_utils

import (
	"bytes"
	"fmt"
	"github.com/baidubce/bce-sdk-go/services/bos"
	"github.com/baidubce/bce-sdk-go/util/log"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
)

const chapterNameFmt = "/files/%d/%d/%d.txt"

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
		log.Fatal("bos init error ", err)
	}

	return &BosUtil{
		bos:    bosClient,
		bucket: "testxxfile",
	}
}

func (b *BosUtil) PutChapter(aid, cid int, content string) error {
	reader := transform.NewReader(bytes.NewReader([]byte(content)), simplifiedchinese.GBK.NewDecoder())
	bb, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	objName := fmt.Sprintf(chapterNameFmt, aid/1000, aid, cid)
	r, err := b.bos.PutObjectFromBytes(b.bucket, objName, bb, nil)
	fmt.Println(r)
	return err
}

func (b *BosUtil) PutCover(aid int) error {
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
