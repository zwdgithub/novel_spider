package model

/*
siteid, postdate, lastupdate, articlename, keywords, initial, author, posterid, poster, sortid, intro, notice, setting
*/
type JieqiArticle struct {
	Articleid   int `gorm:"primary_key"`
	Articlename string
	Author      string
	Lastchapter string
	Chapters    int
	Postdate    int
	Lastupdate  int
	Keywords    string
	Posterid    int
	Sortid      int
	Intro       string
	Notice      string
	Setting     string
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
}

func (JieqiChapter) TableName() string {
	return "jieqi_article_chapter"
}

func (JieqiArticle) TableName() string {
	return "jieqi_article_article"
}
