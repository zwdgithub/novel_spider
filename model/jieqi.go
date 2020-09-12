package model

/*
siteid, postdate, lastupdate, articlename, keywords, initial, author, posterid, poster, sortid, intro, notice, setting
*/
type JieqiArticle struct {
	Articleid     int `gorm:"primary_key"`
	Articlename   string
	Author        string
	Lastchapter   string
	Lastchapterid int
	Chapters      int
	Postdate      int
	Lastupdate    int
	Keywords      string
	Posterid      int
	Sortid        int
	Intro         string
	Notice        string
	Setting       string
	Fullflag      int
	Imgflag       int
}

type JieqiChapter struct {
	Chapterid    int `gorm:"primary_key"`
	Chaptername  string
	Articleid    int
	Articlename  string
	Size         int
	Chapterorder int
	Siteid       int
	Posterid     int
	Poster       string
	Postdate     int
	Lastupdate   int
	Attachment   string
	Chaptertype  int
}

type ChapterErrorLog struct {
	Id         int
	Host       string
	ArticleId  int
	ChapterId  int
	Url        string
	ErrorType  int
	RetryNum   int
	Repair     int
	CreateTime string
	UpdateTime string
}

func (ChapterErrorLog) TableName() string {
	return "chapter_error_log"
}

func (JieqiChapter) TableName() string {
	return "jieqi_article_chapter"
}

func (JieqiArticle) TableName() string {
	return "jieqi_article_article"
}
