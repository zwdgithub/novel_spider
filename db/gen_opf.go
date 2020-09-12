package db

import (
	"fmt"
	"novel_spider/log"
	"novel_spider/model"
	"strings"
	"time"
)

var (
	xmlClear = map[string]string{
		"<":  "&lt;",
		">":  "&gt;",
		"\"": "&quot;",
		"'":  "&apos;",
		"&":  "&amp;",
	}
)

func clearXml(str string) string {
	for k, v := range xmlClear {
		str = strings.ReplaceAll(str, k, v)
	}
	return str
}

func createManifestTag(chapter model.JieqiChapter) string {
	text := `<item id="%s" href="%d.txt" media-type="text/html" content-type="%s" />`
	t := "chapter"
	if chapter.Chaptertype != 0 {
		t = "volume"
	}
	return fmt.Sprintf(text, strings.ReplaceAll(chapter.Chaptername, "&", "&amp;"), chapter.Chapterid, t)

}

func createSpineTag(chapter model.JieqiChapter) string {
	text := `<itemref idref= "%s" />`
	return fmt.Sprintf(text, strings.ReplaceAll(chapter.Chaptername, "&", "&amp;"))
}

func formDate(t int) string {
	return time.Unix(int64(t), 0).Format("2006-01-02 15:04:05")
}

func (service *ArticleService) GenOpf(aid int) {
	var (
		list    []model.JieqiChapter
		article model.JieqiArticle
	)
	err := service.db.Where("articleid = ?", aid).Order("chapterorder asc").Find(&list).Error
	if err != nil {
		log.Infof("gen opf %d, get chapter error", aid)
		return
	}
	err = service.db.Where("articleid = ?", aid).Find(&article).Error
	if err != nil {
		log.Infof("gen opf %d, get article error", aid)
		return
	}
	hostName := "http://http://www.2hxs.com/"
	content := ""
	content += `<?xml version="1.0" encoding="ISO-8859-1"?>`
	content += "<package "
	content += fmt.Sprintf("unique-identifier=%s_%d>", hostName, aid)
	content += "<metadata>"
	content += "<dc-metadata>"
	content += fmt.Sprintf("<dc:Title>%s</dc:Title>", article.Articlename)
	content += fmt.Sprintf("<dc:Creator>%s</dc:Creator>", article.Author)
	content += "<dc:Subject></dc:Subject>"

	intro := strings.ReplaceAll(article.Intro, "\r", "")
	intro = strings.ReplaceAll(article.Intro, "\t", "")
	intro = strings.ReplaceAll(article.Intro, "\n", "")

	content += fmt.Sprintf("<dc:Description>%s</dc:Description>", intro)
	content += fmt.Sprintf("<dc:Publisher>%s</dc:Publisher>", "爱好小说")
	content += fmt.Sprintf("<dc:Contributorid>%s</dc:Contributorid>", "1")
	content += fmt.Sprintf("<dc:Contributor>%s</dc:Contributor>", "admin")
	content += fmt.Sprintf("<dc:Sortid>%d</dc:Sortid>", article.Sortid)
	content += fmt.Sprintf("<dc:Typeid>%d</dc:Typeid>", 0)
	content += fmt.Sprintf("<dc:Articletype>%s</dc:Articletype>", "0")
	content += fmt.Sprintf("<dc:Permission>%d</dc:Permission>", 0)
	content += fmt.Sprintf("<dc:Firstflag>%d</dc:Firstflag>", article.Fullflag)
	content += fmt.Sprintf("<dc:Imgflag>%d</dc:Imgflag>", article.Imgflag)
	content += fmt.Sprintf("<dc:Power>%d</dc:Power>", 0)
	content += fmt.Sprintf("<dc:Display>%d</dc:Display>", 0)
	content += fmt.Sprintf("<dc:Date>%s</dc:Date>", formDate(article.Lastupdate))
	content += fmt.Sprintf("<dc:Type>%s</dc:Type>", "Text")
	content += fmt.Sprintf("<dc:Format>%s</dc:Format>", "text")
	content += fmt.Sprintf("<dc:Language>%s</dc:Language>", "ZH")
	content += "</dc-metadata>"
	content += "</metadata>"
	content += "<manifest>"

	mainfestText := ""
	spineText := ""
	if len(list) > 0 {
		for _, chapter := range list {
			mainfestText += createManifestTag(chapter)
		}
		for _, chapter := range list {
			chapter.Chaptername = clearXml(chapter.Chaptername)
			mainfestText += createManifestTag(chapter)
			spineText += createSpineTag(chapter)
		}

	}
	content += mainfestText
	content += "</manifest>"
	content += "<spine>"
	content += spineText
	content += "</spine></package>"
	err = service.bos.PutOpf(aid, content)
	if err != nil {
		log.Infof("bos put %d index opf file error, msg: %v", aid, err)
	}
}
