package test

import (
	"fmt"
	"github.com/antchfx/htmlquery"
	"novel_spider/article"
	"novel_spider/bos_utils"
	"novel_spider/db"
	"novel_spider/log"
	"novel_spider/model"
	"novel_spider/redis"
	"novel_spider/util"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestProcess(t *testing.T) {
	c := make(chan int, 0)
	go func() {
		c <- 1
	}()
	dbConf := db.LoadMysqlConfig("config/conf.yaml")
	bosClient := bos_utils.NewBosClient("config/bos_conf.yaml")
	dbConn := db.New(dbConf)
	redisConn := redis.NewRedis()
	service := db.NewArticleService(dbConn, redisConn, bosClient)
	spider := article.CreateBiqugeBizSpider(service, redisConn, bosClient)
	spider.Process(article.NewArticle{
		Url:            "https://www.biquge.biz/34_34415/",
		NewChapterName: "第五百一十四章 不寒而栗",
	}, c)

}

func TestLog(t *testing.T) {
}

func TestReg(t *testing.T) {
	content := `

<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
<meta http-equiv="Content-Type" content="text/html; charset=gbk" />
<meta http-equiv="Cache-Control" content="no-siteapp" />
<meta http-equiv="Cache-Control" content="no-transform" />
<script type="text/javascript" src="/js/m.js"></script>
<title>总裁爹地宠上天（唐悠悠季枭寒） 贝小爱 第13章 应聘工作 - 笔趣阁</title>
<meta name="keywords" content="第13章 应聘工作,贝小爱,总裁爹地宠上天（唐悠悠季枭寒）" />
<meta name="description" content="总裁爹地宠上天（唐悠悠季枭寒）最新章节第13章 应聘工作在线阅读,所有小说均免费阅读,努力打造最干净的阅读环境,24小时不间断更新,请大家告诉更多的小说迷。" />
<meta http-equiv="mobile-agent" content="format=html5; url=https://m.biquge.biz/22/22779/9217977.html" />
<meta http-equiv="mobile-agent" content="format=xhtml; url=https://m.biquge.biz/22/22779/9217977.html" />
<script type="text/javascript" src="/js/zepto.min.js"></script>
<script type="text/javascript" src="/js/common.js?v1"></script>
<script type="text/javascript" src="/js/read.js?v1"></script>
<script type="text/javascript" src="/js/bookcase.js?v1"></script>
<link rel="stylesheet" href="/css/style.css" />

</head>
<body>
<div id="wrapper">
	<script>login();</script>
	<div class="header">
		<div class="header_logo">
			<a href="/">笔趣阁</a>
		</div>
		<script>panel();</script>
	</div>
	<div class="nav">
		<ul>
<li><a href="https://www.biquge.biz/">首页</a></li>
<li><a href="/xuanhuan/">玄幻小说</a></li>
<li><a href="/xiuzhen/">修真小说</a></li>
<li><a href="/dushi/">都市小说</a></li>
<li><a href="/chuanyue/">穿越小说</a></li>
<li><a href="/wangyou/">网游小说</a></li>
<li><a href="/kehuan/">科幻小说</a></li>
<li><a href="/qita/">其他小说</a></li>
<li><a href="/paihangbang/">排行榜单</a></li>
<li><a href="/quanben/">完本小说</a></li>
<li><a href="/shujia.html">临时书架</a></li>	
</ul>
	</div>
	<div class="content_read">
		<div class="box_con">
			<div class="con_top"><script>textselect();</script>
				<a href="/">笔趣阁</a> &gt; <a href="/dushi/">都市小说</a> &gt; <a href="/22_22779/">总裁爹地宠上天（唐悠悠季枭寒）</a> &gt; 第13章 应聘工作
			</div>
			<div class="bookname">
				<h1>第13章 应聘工作</h1>
				<div class="bottem1">
					<a href="/22_22779/9217975.html">上一章</a> &larr; <a href="/22_22779/">章节列表</a> &rarr; <a href="/22_22779/9217979.html">下一章</a> <a rel="nofollow" href="javascript:;" onclick="addBookCase(22779,9217977,'第13章 应聘工作');">加入书签</a>
				</div>
				<div class="lm">推荐阅读：<a href="/22_22780/">重生之都市仙尊</a><a href="/22_22781/">透视仙医</a><a href="/22_22782/">大宋之重铸山河</a><a href="/22_22783/">铁路子弟</a><a href="/22_22784/">末世大狙霸</a><a href="/22_22778/">天命武君</a><a href="/22_22777/">乡村极品小仙医</a><a href="/22_22776/">我的超神QQ</a><a href="/22_22775/">超级机器人工厂</a><a href="/22_22774/">我去天庭发红包</a></div>
			</div>
			<div style="text-align: center"><script>read2();</script></div>
			<div id="content"><br>&nbsp;&nbsp;&nbsp;&nbsp;第13章&nbsp;&nbsp;应聘工作<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;唐悠悠在国外就让大姨帮她联系了小区里的一家幼儿园，此刻，唐悠悠带着一对宝贝直接去报名了。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;学费还算比较贵的，唐悠悠交完了两个小萌宝的学费，卡里只剩下一万块不到了。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;真的要省着点儿花了，而且，她必须尽快工作。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;两个小家伙在国外已经上了一年的托班，两年的小班，此刻，已经混成了幼儿园里的小油条了，自然不怕生，那懂事又乖巧的小模样，深得女老师的喜欢。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;老师都感叹，还从来没有见过这么漂亮的孩子，都在趁机追问唐悠悠，她的儿女是不是混血儿。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;唐悠悠拜托了老师之后，就蹲下来，摸摸儿子的小脑袋：“小睿，好好照顾妹妹，下午姨奶会来接你们放学，妈咪要去工作了，你们一定要听话，知道吗？”<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;“妈咪放心吧，我一定会照顾好妹妹，不让人欺负了她的，你放心去工作吧。”唐小睿立即一副很有责任感的表情回答。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;唐小奈已经眼眶泛红，小鼻子抽泣了两下：“妈咪，放学的时候，你能不能第一个来接我？”<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;“奈奈，妈咪今天要去工作了，姨奶会第一时间来接你们的。”唐小奈哄着女儿。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;小睿一把牵住妹妹的小手：“走啦，走啦，哥哥带你上楼玩。”<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;有了哥哥的保护，小家伙这才稍稍有了点安全感，回过头，漂亮的大眼睛含着泪珠儿对唐悠悠挥动了一下小手。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;唐悠悠其实是很放心儿女的，他们适应能力很强，相信一天时间不到，他们就会交上小朋友，玩的开心的。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;现在，她要急着去公司报到了。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;唐悠悠打了个车，急急的赶到公司大厅门口。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;公司名子叫唯意国际设计，名子雅致，名声也是超一流的。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;能够挤身进入这家设计公司工作的，都是设计界的名流主角。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;唐悠悠至所以能够应聘进去，除了她有着独出心裁的设计理念之外，还借助了人脉关系。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;她妈妈生前最要好的死党，已经在唯意设计部任职总设计师，唐悠悠小时候就认她做了干妈。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;这个干妈对她也真够义气的，小时候就是受她的影响才接触设计这一行。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;如今，干妈独挡一面，也能够顺带扶持她一把。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;大厅门口，唐悠悠一身黑色的职业套装，一张年轻干净的脸蛋上，妆容素雅精致。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;她身高有一米六六，身段纤细，一头齐腰的长发，也显出几许妩媚的风情。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;“你是唐悠悠小姐吗？”就在她等待之际，一个声音在喊她。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;她回头，含着笑意点头：“我就是！”<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;“我是刘设计师的助手，跟我去人事办理一下入职手续吧。”长相普通的小助手，在看到唐悠悠的外表时，惊震了一下，没想到，竟然是个大美人。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;“好！麻烦你了！”唐悠悠礼貌客气的跟着小助理去了人事，办理了入职手续。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;原本是打算去跟干妈打声招呼的，却很不巧，干妈出去办事了。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;唐悠悠只好打算先离开，明天才是正式上班的日子。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;她低着个头，盘算着一会儿空出的时间要去趁市采购一些生活用品。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;电梯门，突然打开。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;唐悠悠也没去注意到旁边私人直达电梯门口站着的一群人。<br></div>
			<div style="text-align: center"><script>read3();</script></div>
			<div class="bottem2">
				<a href="/22_22779/9217975.html">上一章</a> &larr; <a href="/22_22779/">章节目录</a> &rarr; <a href="/22_22779/9217979.html">下一章</a> <a rel="nofollow" href="javascript:;" onclick="addBookCase(22779,9217977,'第13章 应聘工作');">加入书签</a>
			</div>
			<div style="text-align: center"><script>read4();</script></div>
		</div>
	</div>
	<div class="footer">
		<div class="footer_link">强烈推荐：<a href="/25_25043/">开天录</a><a href="/22_22779/">总裁爹地宠上天</a><a href="/12_12453/">我的绝色总裁未婚妻</a><a href="/5_5164/">圣墟</a><a href="/0_166/">天下第九</a><a href="/9_9668/">女总裁的全能兵王</a><a href="/25_25741/">一剑独尊</a><a href="/23_23658/">太古狂魔</a><a href="/0_686/">武炼巅峰</a></div>
		<div class="footer_cont">
			<p>
《<a href="/22_22779/">总裁爹地宠上天（唐悠悠季枭寒）</a>》情节跌宕起伏、扣人心弦，是一本情节与文笔俱佳的都市小说，笔趣阁转载收集仗剑问仙最新章节。</p>
<p>本站所有小说为转载作品，所有章节均由网友上传，转载至本站只是为了宣传本书让更多读者欣赏。</p>
<p>Copyright ? 2016 笔趣阁 All Rights Reserved.</p>
			<script>footer();</script>
			<script>tan();</script>
		</div>
	</div>
</div>
<script language="javascript">
document.onkeydown=keypage;
var prevpage="/22_22779/9217975.html"
var nextpage="/22_22779/9217979.html"
var index_page = "/22_22779/"
function keypage() {
	if (event.keyCode==37) location=prevpage
	if (event.keyCode==39) location=nextpage
	if (event.keyCode == 13) document.location=index_page
}
addHit(22779);
AddbookMarket('22779', '总裁爹地宠上天（唐悠悠季枭寒）','9217977',  '第13章 应聘工作', '都市小说','贝小爱','/22_22779/',window.location.href);
function postErrorChapter(){
	postError(9217977,22779);
}
</script>
</body>
</html>
`
	reg := regexp.MustCompile(`<div id="content">(.+?)</div>`)
	c := reg.FindStringSubmatch(content)
	c[1] = strings.ReplaceAll(c[1], "&nbsp;", "")
	c[1] = strings.ReplaceAll(c[1], "<br>", "\r\n")
	c[1] = strings.ReplaceAll(c[1], "<br >", "\r\n")
	c[1] = strings.ReplaceAll(c[1], "<br/>", "\r\n")
	t.Log(c[1])
}

func TestBos(t *testing.T) {
	bos := bos_utils.NewBosClient("config/bos_conf.yaml")
	//bos.PutChapter(11111, 11212, "<br>&nbsp;&nbsp;&nbsp;&nbsp;第13章&nbsp;&nbsp;应聘工作<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;唐悠悠在国外就让大姨帮她联系了小区里的一家幼儿园，此刻，唐悠悠带着一对宝贝直接去报名了。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;学费还算比较贵的，唐悠悠交完了两个小萌宝的学费，卡里只剩下一万块不到了。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;真的要省着点儿花了，而且，她必须尽快工作。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;两个小家伙在国外已经上了一年的托班，两年的小班，此刻，已经混成了幼儿园里的小油条了，自然不怕生，那懂事又乖巧的小模样，深得女老师的喜欢。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;老师都感叹，还从来没有见过这么漂亮的孩子，都在趁机追问唐悠悠，她的儿女是不是混血儿。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;唐悠悠拜托了老师之后，就蹲下来，摸摸儿子的小脑袋：“小睿，好好照顾妹妹，下午姨奶会来接你们放学，妈咪要去工作了，你们一定要听话，知道吗？”<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;“妈咪放心吧，我一定会照顾好妹妹，不让人欺负了她的，你放心去工作吧。”唐小睿立即一副很有责任感的表情回答。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;唐小奈已经眼眶泛红，小鼻子抽泣了两下：“妈咪，放学的时候，你能不能第一个来接我？”<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;“奈奈，妈咪今天要去工作了，姨奶会第一时间来接你们的。”唐小奈哄着女儿。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;小睿一把牵住妹妹的小手：“走啦，走啦，哥哥带你上楼玩。”<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;有了哥哥的保护，小家伙这才稍稍有了点安全感，回过头，漂亮的大眼睛含着泪珠儿对唐悠悠挥动了一下小手。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;唐悠悠其实是很放心儿女的，他们适应能力很强，相信一天时间不到，他们就会交上小朋友，玩的开心的。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;现在，她要急着去公司报到了。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;唐悠悠打了个车，急急的赶到公司大厅门口。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;公司名子叫唯意国际设计，名子雅致，名声也是超一流的。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;能够挤身进入这家设计公司工作的，都是设计界的名流主角。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;唐悠悠至所以能够应聘进去，除了她有着独出心裁的设计理念之外，还借助了人脉关系。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;她妈妈生前最要好的死党，已经在唯意设计部任职总设计师，唐悠悠小时候就认她做了干妈。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;这个干妈对她也真够义气的，小时候就是受她的影响才接触设计这一行。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;如今，干妈独挡一面，也能够顺带扶持她一把。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;大厅门口，唐悠悠一身黑色的职业套装，一张年轻干净的脸蛋上，妆容素雅精致。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;她身高有一米六六，身段纤细，一头齐腰的长发，也显出几许妩媚的风情。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;“你是唐悠悠小姐吗？”就在她等待之际，一个声音在喊她。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;她回头，含着笑意点头：“我就是！”<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;“我是刘设计师的助手，跟我去人事办理一下入职手续吧。”长相普通的小助手，在看到唐悠悠的外表时，惊震了一下，没想到，竟然是个大美人。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;“好！麻烦你了！”唐悠悠礼貌客气的跟着小助理去了人事，办理了入职手续。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;原本是打算去跟干妈打声招呼的，却很不巧，干妈出去办事了。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;唐悠悠只好打算先离开，明天才是正式上班的日子。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;她低着个头，盘算着一会儿空出的时间要去趁市采购一些生活用品。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;电梯门，突然打开。<br><br><br><br>&nbsp;&nbsp;&nbsp;&nbsp;唐悠悠也没去注意到旁边私人直达电梯门口站着的一群人。")
	t.Log(bos.PutCover("1", 1))
}

func TestLogger(t *testing.T) {
	log.Infof("this is a log msg")
}

func TestReflect(t *testing.T) {
	util.LoadYaml("../config/bos_conf.yaml", bos_utils.BosConf{})
}

func TestGBKHttp(t *testing.T) {
	content, err := util.Get("https://m.ihxs.la/69_69938/50083136.html", "gbk")
	if err != nil {
		t.Log(err)
	}
	doc, err := htmlquery.Parse(strings.NewReader(content))
	if err != nil {
		t.Log(err)
	}
	cNode, err := htmlquery.Query(doc, `//div[@id="chaptercontent"]`)
	if err != nil {
		t.Log(err)
	}
	content = htmlquery.OutputHTML(cNode, false)
	content = strings.Replace(content, " ", "", -1)
	t.Log(content)

}

func TestRedis(t *testing.T) {
	r := redis.NewRedis()
	r.CanParse("替嫁小妻超甜超可爱", "浅若初秋")
}

func TestGenOpf(t *testing.T) {
	// 37490
	dbConf := db.LoadMysqlConfig("config/conf.yaml")
	bosClient := bos_utils.NewBosClient("config/bos_conf.yaml")
	dbConn := db.New(dbConf)
	redisConn := redis.NewRedis()
	service := db.NewArticleService(dbConn, redisConn, bosClient)
	service.GenOpf(37490)
}

func TestDeferFunc(t *testing.T) {
	dbConf := db.LoadMysqlConfig("config/conf.yaml")
	bosClient := bos_utils.NewBosClient("config/bos_conf.yaml")
	dbConn := db.New(dbConf)
	redisConn := redis.NewRedis()
	service := db.NewArticleService(dbConn, redisConn, bosClient)
	service.AddErrorChapter(
		model.ChapterErrorLog{
			Host:      "biquge.biz",
			ArticleId: 30042,
			ChapterId: 20222211,
			Url:       "https://www.baidu.com",
			ErrorType: 1,
			RetryNum:  0,
		},
	)
}

func TestDate(t *testing.T) {
	a, _ := time.ParseDuration(fmt.Sprintf("-%dh", 24*7))
	n := time.Now().Add(a).Format("2006-01-02 15:04:05")
	fmt.Println(n)
	t.Log(util.ValidChapterName("123"))
}

func TestKanshuSpider(t *testing.T) {
	dbConf := db.LoadMysqlConfig("config/conf.yaml")
	bosClient := bos_utils.NewBosClient("config/bos_conf.yaml")
	dbConn := db.New(dbConf)
	redisConn := redis.NewRedis()
	service := db.NewArticleService(dbConn, redisConn, bosClient)
	spider := article.CreateKanshuLaSpider(service, redisConn, bosClient)
	c := make(chan int, 1)
	c <- 1
	spider.Process(article.NewArticle{
		Url:            "https://www.kanshu5.la/133/133537/",
		NewChapterName: "",
	}, c)
}
