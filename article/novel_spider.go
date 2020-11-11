package article

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/antlabs/strsim"
	"math"
	"novel_spider/db"
	"novel_spider/log"
	"novel_spider/model"
	"novel_spider/redis"
	"novel_spider/util"
	"runtime"
	"strings"
	"time"
)

var (
	contentShortError = errors.New("content length too short")
	chapterNotMatch   = errors.New("no chapter need to update ")
	shortContent      = "看最快更新无错小说，请记住 https://www.ihxs.la！章节内容正在手打中，请稍等片刻，内容更新后，请重新刷新页面，即可获取最新更新！"
)

type NovelWebsites interface {
	ArticleInfo(content string) (*Article, error)
	ChapterList(content string) ([]NewChapter, error)
	ChapterContent(chapterUrl string) (string, error)
	Consumer() (string, error)
	ConsumerMany() (string, error)
	HasNext(url string) *NewChapter
	NewList() ([]string, error)
}

type NovelSpider struct {
	ws      NovelWebsites
	wsInfo  *NovelWebsite
	service *db.ArticleService
	redis   *redis.RedisUtil
}

type NewArticle struct {
	Url            string
	NewChapterName string
	MaxChapterNum  int
	Host           string
}

type NewChapter struct {
	Url         string
	ChapterName string
}

func NewNovelSpider(ws NovelWebsites, wsInfo *NovelWebsite, service *db.ArticleService, redis *redis.RedisUtil) *NovelSpider {
	return &NovelSpider{
		ws:      ws,
		wsInfo:  wsInfo,
		service: service,
		redis:   redis,
	}
}

func (s *NovelSpider) CanParse(articleName, author string) (bool, error) {
	return s.redis.CanParse(articleName, author)
}

func (s *NovelSpider) ParseEnd(articleName, author string) {
	s.redis.ParseEnd(articleName, author)
}

func (s *NovelSpider) Consumer(many bool) {
	c := make(chan int, s.wsInfo.Concurrent)
	for {
		if s.redis.Pause(s.wsInfo.Host) {
			log.Infof("%s, spider stop", s.wsInfo.Host)
			break
		}
		var content string
		var err error
		if len(c) < s.wsInfo.Concurrent {
			if many {
				content, err = s.ws.ConsumerMany()
			} else {
				content, err = s.ws.Consumer()
			}
			if err != nil {
				time.Sleep(time.Second * 5)
				continue
			}
			var obj NewArticle
			err = json.Unmarshal([]byte(content), &obj)
			if err != nil {
				log.Errorf("consumer Unmarshal error:%v, value: %s", err, content)
				continue
			}
			c <- 1
			obj.MaxChapterNum = 100
			if many {
				obj.MaxChapterNum = 100000
			}
			go s.Process(obj, c)
		}
		time.Sleep(time.Second / 2)
	}
	for len(c) > 0 {

	}
	log.Infof("%s, stop success", s.wsInfo.Host)
}

func (s *NovelSpider) Process(obj NewArticle, c chan int) {
	defer func() {
		<-c
		if err := recover(); err != nil {
			log.Errorf("process %s, err: %v", obj.Url, err)
			stack := make([]byte, 1024*8)
			stack = stack[:runtime.Stack(stack, false)]
			log.Error(string(stack))
		}
		log.Infof("process %s, end", obj.Url)
	}()
	log.Infof("process %s, start", obj.Url)
	content := ""
	var err error
	if s.wsInfo.Proxy {
		content, err = util.GetWithProxy(obj.Url, s.wsInfo.Encoding, s.wsInfo.Headers)
	} else {
		content, err = util.Get(obj.Url, s.wsInfo.Encoding, s.wsInfo.Headers)
	}
	if err != nil {
		log.Infof("process %s, http get error: %v", obj.Url, err)
		return
	}
	article, err := s.ws.ArticleInfo(content)
	if err != nil || article == nil || article.ArticleName == "" || article.Author == "" {
		log.Infof("process %s, parse article info error, msg: %v", obj.Url, err)
		return
	}
	canParse, err := s.CanParse(article.ArticleName, article.Author)
	if err != nil {
		log.Infof("process %s, can not parse now, error: %v", obj.Url, err)
		return
	}
	if !canParse {
		log.Infof("process %s, can not parse now,", obj.Url)
		return
	}
	defer s.ParseEnd(article.ArticleName, article.Author)
	obj.NewChapterName = strings.TrimSpace(obj.NewChapterName)
	local, err := s.service.LocalArticleInfo(article.ArticleName, article.Author)
	if err != nil {
		log.Infof("process %s, get local info error: %v ", obj.Url, err)
		return
	}
	if local.Articleid == 0 {
		newArticle := &model.JieqiArticle{
			Articlename: article.ArticleName,
			Author:      article.Author,
			Intro:       article.Intro,
			Sortid:      article.SortId,
		}
		if newArticle.Intro == "" {
			newArticle.Intro = article.ArticleName
		}
		err := s.service.AddArticle(newArticle)
		_ = s.wsInfo.BosUtil.PutCover(article.ImgUrl, newArticle.Articleid)
		if err != nil {
			log.Infof("process %s, add new article error %v", obj.Url, err)
			return
		}
		local = newArticle
		log.Infof("process %s, add new article, %s, %s", obj.Url, local.Articlename, local.Author)
		go s.service.GenOpf(local.Articleid)
	}

	allChapters, err := s.ws.ChapterList(content)
	if err != nil || len(allChapters) == 0 {
		log.Infof("process %s, parse chapter list error: %v", obj.Url, err)
		return
	}

	for i, v := range allChapters {
		allChapters[i].ChapterName = strings.TrimSpace(v.ChapterName)
	}
	targetLast := obj.NewChapterName
	if targetLast == "" {
		targetLast = allChapters[len(allChapters)-1].ChapterName
	}

	article.LastChapter = targetLast
	if article.LastChapter == local.Lastchapter {
		log.Infof("process %s, need not update", obj.Url)
		return
	}

	order := local.Chapters
	newChapters := make([]NewChapter, 0)
	match := false
	local.Lastchapter = util.Trim(local.Lastchapter)

	if local.Chapters == 0 { // 新书
		match = true
		newChapters = allChapters
	}

	if !match {
		for i := len(allChapters) - 1; i >= 0; i-- { // 从后向前匹配章节名
			if allChapters[i].ChapterName == local.Lastchapter {
				match = true
				for j := i + 1; j < len(allChapters); j++ {
					newChapters = append(newChapters, allChapters[j])
				}
				break
			}
		}
	}

	if match && len(newChapters) == 0 {
		log.Infof("process %s, new chapters none, info: name:%s, author:%s, last:%s", obj.Url, article.ArticleName, article.Author, article.LastChapter)
		if obj.NewChapterName != "" && obj.NewChapterName != allChapters[len(allChapters)-1].ChapterName {
			if s.service.ChapterNameExists(local.Articleid, obj.NewChapterName) {
				return
			}
			log.Infof("process %s need retry, new: %s, old:%s", obj.Url, obj.NewChapterName, allChapters[len(allChapters)-1].ChapterName)
			s.retry(s.wsInfo.Host, obj.Url, obj.NewChapterName)
		}
		return
	}
	if !match {
		log.Infof("process %s, try to match last chapter", obj.Url)
		num := 5
		lastList := s.service.LastChapterList(local.Articleid, num)
		localCache := make([]string, 0)
		for _, v := range lastList {
			content, err := s.service.GetLocalContent(v.Articleid, v.Chapterid)
			content = strings.ReplaceAll(content, "\r", "")
			content = strings.ReplaceAll(content, "\n", "")
			if err != nil {
				log.Infof("process %s, get local content error: %v", obj.Url, err)
				return
			}
			if len(content) < 500 {
				if util.ValidChapterName(v.Chaptername) {
					localCache = make([]string, 0)
					goto matchLabel
				}
				continue
			}
			localCache = append(localCache, content)
		}

		for i := len(allChapters) - 1; i >= 0; i-- {
			content, err := s.ws.ChapterContent(allChapters[i].Url)
			if err != nil {
				log.Infof("process %s, try to match chapter get content error: %v", obj.Url, err)
				return
			}
			content = strings.ReplaceAll(content, "\r", "")
			content = strings.ReplaceAll(content, "\n", "")
			if len(content) < 500 {
				goto matchLabel
			}
			for _, c := range localCache {
				score := strsim.Compare(content, c, strsim.DiceCoefficient())
				if score >= 0.75 {
					match = true
					for j := i + 1; j < len(allChapters); j++ {
						newChapters = append(newChapters, allChapters[j])
					}
					log.Infof("process %s, try to match chapter success, new chapter len is %d", obj.Url, len(newChapters))
					if len(newChapters) == 0 {
						return
					}
					goto matchLabel
				}
			}
			if len(allChapters)-i >= 5 {
				break
			}
		}

		log.Infof("process %s, try to match all chapter", obj.Url)
		index, err := s.tryFindNewChapter(obj, allChapters, local)
		if err != nil {
			goto matchLabel
		}
		match = true
		for i := index + 1; i < len(allChapters); i++ {
			newChapters = append(newChapters, allChapters[i])
		}

		log.Infof("process %s, try to match chapter success, new chapter len is %d", obj.Url, len(newChapters))
		if len(newChapters) == 0 {
			return
		}
	}

matchLabel:
	if !match {
		matchLog := &model.NovelNotMatchLog{
			Host:             s.wsInfo.Host,
			Url:              obj.Url,
			ArticleName:      local.Articlename,
			Author:           local.Author,
			NewLastChapter:   allChapters[len(allChapters)-1].ChapterName,
			LocalArticleId:   local.Articleid,
			LocalLastChapter: local.Lastchapter,
		}
		s.service.SaveNotMachLog(matchLog)
		log.Infof("process %s, no chapter match, info: %s, %s, %s, %s", obj.Url, local.Articlename, local.Author, allChapters[len(allChapters)-1].ChapterName, local.Lastchapter)
		return
	}

	log.Infof("process %s, need crawl chapter %d", obj.Url, len(newChapters))
	if len(newChapters) == 0 {
		log.Infof("process %s, new chapters none, info: name:%s, author:%s, last:%s", obj.Url, article.ArticleName, article.Author, article.LastChapter)
		if article.LastChapter != local.Lastchapter {

		}
		return
	}

	if len(newChapters) > obj.MaxChapterNum {
		b, _ := json.Marshal(NewArticle{
			Url:            obj.Url,
			NewChapterName: obj.NewChapterName,
			Host:           s.wsInfo.Host,
		})
		s.redis.PutUrlToQueue(s.wsInfo.Host+"_many_chapters", string(b))
		log.Infof("process %s, need crawl chapter too many, chapter num: %d, max: %d", obj.Url, len(newChapters), obj.MaxChapterNum)
		return
	}

	addChapterNum := 0

	defer func() {
		if addChapterNum > 0 {
			s.service.GenOpf(local.Articleid)
		}
	}()
	matched := false
	for i, item := range newChapters {
		if s.redis.Pause(s.wsInfo.Host) {
			log.Infof("process %s stop", obj.Url)
			return
		}
		if obj.NewChapterName != "" && obj.NewChapterName == item.ChapterName {
			matched = true
		}
		content, err := s.ws.ChapterContent(item.Url)
		if err != nil {
			log.Infof("process %s, get content error: %v, content is %s, add new chapter: %d", obj.Url, err, content, addChapterNum)
			return
		}
		var contentError error
		if len(content) <= s.wsInfo.ShortContent {
			contentError = errors.New(fmt.Sprintf("process %s content short", obj.Url))
			content = shortContent
		}
		if strings.Contains(content, "@font-face") {
			contentError = errors.New(fmt.Sprintf("process %s content qidian error", obj.Url))
			content = shortContent
		}
		chapter := &model.JieqiChapter{
			Chapterorder: order + 1,
			Chaptername:  item.ChapterName,
			Articleid:    local.Articleid,
			Articlename:  local.Articlename,
		}
		chapter, err = s.service.AddChapter(chapter, content)

		log.Infof("process %s, new chapter name: %s, local article id: %d, ", obj.Url, item.ChapterName, local.Articleid)

		if util.ValidChapterName(item.ChapterName) && contentError != nil && chapter != nil && chapter.Chapterid != 0 {
			s.service.AddErrorChapter(model.ChapterErrorLog{
				Host:      s.wsInfo.Host,
				ArticleId: local.Articleid,
				ChapterId: chapter.Chapterid,
				Url:       item.Url,
				ErrorType: 1,
				RetryNum:  0,
			})
		}
		if err != nil {
			log.Infof("process %s add chapter error: %v", obj.Url, err)
			return
		}
		addChapterNum++
		order += 1

		if i == len(allChapters)-1 && !matched && obj.NewChapterName != "" && obj.NewChapterName != item.ChapterName {
			log.Infof("process %s need retry, new: %s, old:%s", obj.Url, obj.NewChapterName, newChapters[len(newChapters)-1].ChapterName)
			s.retry(s.wsInfo.Host, obj.Url, obj.NewChapterName)
		}
	}
	log.Infof("process %s, success, add %d chapter", obj.Url, addChapterNum)

	return
}

func (s *NovelSpider) NewList() {
	list, err := s.ws.NewList()
	if err != nil {
		return
	}
	for _, u := range list {
		s.redis.PutUrlToQueue(s.wsInfo.Host, u)
	}
}

func (s *NovelSpider) Repair() {
	offset := 0
	for {
		list := s.service.NeedRepairChapterList(s.wsInfo.Host, offset)
		log.Infof("repair, need repair list len is %d", len(list))
		for _, item := range list {
			content, err := s.ws.ChapterContent(item.Url)
			if err != nil {
				log.Infof("repair %s, get content error: %v", item.Url, err)
				continue
			}

			if len(content) <= s.wsInfo.ShortContent {
				s.service.UpdateErrorChapter(item.Id, item.RetryNum+1, 0)
				continue
			}
			err = s.service.PutContent(item.ArticleId, item.ChapterId, content)
			if err != nil {
				s.service.UpdateErrorChapter(item.Id, item.RetryNum+1, 0)
				continue
			}
			s.service.UpdateErrorChapter(item.Id, item.RetryNum+1, 1)
			s.service.RepairSyncSameAll(item.ArticleId)
			log.Infof("repair success %s", item.Url)
		}
		if len(list) == 100 {
			offset += 100
			continue
		} else {
			offset = 0
		}
		time.Sleep(time.Minute * 10)
	}
}

func (s *NovelSpider) Retry() {
	for {
		result, err := s.redis.GetRetryArticle(s.wsInfo.Host)
		if err != nil {
			log.Infof("process %s, retry error: %v", s.wsInfo.Host, err)
			time.Sleep(time.Second * 10)
			continue
		}
		var item NewArticle
		err = json.Unmarshal([]byte(result[0].Member.(string)), &item)
		if err != nil {
			log.Infof("process %s, retry error: %v", s.wsInfo.Host, err)
			time.Sleep(time.Second * 10)
			continue
		}
		s.redis.PutUrlToQueue(item.Host, result[0].Member.(string))
	}
}

func (s *NovelSpider) retry(host, url, chapterName string) {
	b, _ := json.Marshal(NewArticle{
		Url:            url,
		NewChapterName: chapterName,
		Host:           host,
	})
	s.redis.Retry(s.wsInfo.Host, string(b))
}

func (s *NovelSpider) tryFindNewChapter(obj NewArticle, allChapter []NewChapter, local *model.JieqiArticle) (int, error) {
	num := 10
	lastList := s.service.LastChapterList(local.Articleid, num)
	count := s.service.ChapterCount(local.Articleid)

	for _, v := range lastList {
		content, err := s.service.GetLocalContent(v.Articleid, v.Chapterid)
		content = strings.ReplaceAll(content, "\r", "")
		content = strings.ReplaceAll(content, "\n", "")
		if len(content) < 500 && util.ValidChapterName(v.Chaptername) {
			log.Infof("process %s, get local content length short, chapter id: %d", obj.Url, v.Chapterid)
			return 0, errors.New("")
		}
		if err != nil {
			log.Infof("process %s, get local content error: %v", obj.Url, err)
			return 0, errors.New("")
		}
		c1 := v.Chaptername
		c1 = strings.TrimSpace(c1)
		splitFlag := false
		if strings.Contains(c1, " ") {
			splits := strings.Split(c1, " ")
			c1 = strings.ReplaceAll(strings.Join(splits[1:], ""), " ", "")
			splitFlag = true
		}
		for i, chapter := range allChapter {
			c2 := chapter.ChapterName
			c2 = strings.TrimSpace(c2)
			c1Temp := c1
			if splitFlag {
				if strings.Contains(c2, " ") {
					splits := strings.Split(c2, " ")
					c2 = strings.ReplaceAll(strings.Join(splits[1:], ""), " ", "")
				} else {
					c1Temp = v.Chaptername
				}
			}
			score := strsim.Compare(c1Temp, c2, strsim.DiceCoefficient())
			if score > 0.65 && math.Abs(float64(count-i)) <= 100 {
				log.Infof("process %s, try to match all chapter, c1: %s, c2: %s, score: %v", obj.Url, c1, c2, score)
				newContent, err := s.ws.ChapterContent(chapter.Url)
				if err != nil {
					log.Infof("process %s, tryFindNewChapter get content error: %v", obj.Url, err)
					return 0, errors.New("")
				}
				score = strsim.Compare(content, newContent, strsim.DiceCoefficient())
				if score >= 0.75 {
					return i, nil
				}
			}
		}

	}

	return 0, errors.New("")
}
