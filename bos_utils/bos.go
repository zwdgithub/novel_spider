package bos_utils

import (
	"fmt"
	"github.com/baidubce/bce-sdk-go/services/bos"
	"github.com/baidubce/bce-sdk-go/util/log"
	"io/ioutil"
)

const chapterNameFmt = "/files/%d/%d/%d.txt"

type BosUtil struct {
	bos    *bos.Client
	bucket string
}

func NewBosClient() *bos.Client {
	// 用户的Access Key ID和Secret Access Key
	AK, SK := "", ""
	ENDPOINT := ""
	// 初始化一个BosClient
	bosClient, err := bos.NewClient(AK, SK, ENDPOINT)
	if err != nil {
		log.Fatal("bos init error ", err)
	}
	return bosClient
}

func (b *BosUtil) PutChapter(aid, cid int, content string) error {
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
