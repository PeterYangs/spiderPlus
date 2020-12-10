package spiderPlus

import (
	//"bytes"
	"crypto/rand"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/PeterYangs/tools"
	"github.com/PuerkitoBio/goquery"
	uuid "github.com/satori/go.uuid"
	//"io"
	"log"
	"math/big"
	"os"
	"regexp"
	//"spider/tool"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Config struct {
	pageIndex       int
	titleSelector   string
	contentSelector string
	host            string
	dirs            string
	table           string
}

type Task struct {
	Id            uint
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CategoryId    int
	Content       string
	Img           string
	Title         string
	Desc          string
	Keyword       string
	WriteType     int
	Expand        string
	PushTime      time.Time
	AdminIdCreate int
}

var f *excelize.File

var tasks = make(chan Task, 10)

/**
  host:域名     例：https://www.d1xz.net/
  channel:栏目  例：bazi/list_[PAGE].html
  limit: 爬取总页面
  pageStart:起始页面
  listSelector:列表选择器
  listHrefSelector:#列表a链接选择器
  titleSelector:标题选择器
  contentSelector:内容选择器
  dirs:图片文件夹
  imagePrefix:图片链接前缀

*/
func Rule(
	host string,
	channel string,
	limit int, pageStart int,
	listSelector string,
	listHrefSelector string,
	titleSelector string,
	contentSelector string,
	dirs string,
	imagePrefix string) {

	var config = Config{}

	config.titleSelector = titleSelector
	config.contentSelector = contentSelector
	config.host = host
	config.dirs = dirs

	//新建xlsx文件
	f = excelize.NewFile()

	//创建图片文件夹
_:
	os.Mkdir("static/"+dirs, os.ModePerm)

	// 设置工作簿的默认工作表
	f.SetActiveSheet(f.NewSheet("Sheet1"))

	//开启一个协程写入Excel
	go writeExcel()

	//分页爬取
	for config.pageIndex = pageStart; config.pageIndex <= limit; config.pageIndex++ {

		//body := tool.Get(host+strings.Replace(channel, "[PAGE]", strconv.Itoa(config.pageIndex), -1), 10)

		body, err := tools.GetWithString(host + strings.Replace(channel, "[PAGE]", strconv.Itoa(config.pageIndex), -1))

		if body == "" {

			fmt.Println(err)

			continue
		}

		//goquery加载html
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
		if err != nil {
			log.Fatal(err)
		}

		//列表页面同步锁
		var wait sync.WaitGroup

		//panic(listSelector)

		listIsFInd := false

		//var listIsFInd bool

		//获取列表a链接
		doc.Find(listSelector).Each(func(i int, s *goquery.Selection) {

			listIsFInd = true

			//panic("123")

			href := ""

			isFind := false

			if listHrefSelector == "" {

				href, isFind = s.Attr("href")

			} else {

				href, isFind = s.Find(listHrefSelector).Attr("href")

			}

			//panic(href)

			//panic(isFind)

			//href, isFind := s.Attr("href")

			if href == "" {

				fmt.Println("a链接为空")

				return
			}

			if isFind == true {

				wait.Add(1)

				href = getHref(href, host)

				//根据列表的长度开启协程爬取详情页
				go detail(href, &wait, imagePrefix, config)

			}

		})

		if listIsFInd == false {

			fmt.Println(host + strings.Replace(channel, "[PAGE]", strconv.Itoa(config.pageIndex), -1) + "未找到---------------------")

		}

		//panic(*temp)

		wait.Wait()

	}

	close(tasks)

	f.SaveAs(config.dirs + ".xlsx")

	//f

	fmt.Println("执行完成")

}

func detail(href string, wait *sync.WaitGroup, ImagePrefix string, config Config) {

	defer wait.Done()

	defer func() {

		if r := recover(); r != nil {
			fmt.Println("detail捕获到的错误:", r)
		}
	}()

	body, err := tools.GetWithString(href)

	if err != nil {

		fmt.Println(err)

		return

	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))

	if err != nil {

		fmt.Println(err)

		return
	}

	//标题
	title := doc.Find(config.titleSelector).Text()

	//panic(title)

	content, _ := doc.Find(config.contentSelector).Html()

	//panic(content)

	keyword, _ := doc.Find("meta[name=\"keywords\"]").Attr("content")

	//panic(keyword)

	desc := getDesc(doc, config)

	img := ""

	//tool.GetDb().Save()

	if title == "" {

		//log.Panicln("标题为空")

		fmt.Println("标题为空")

		return

	}

	if content == "" {

		//log.Println(contentSelector)

		//log.Panicln("内容为空")

		fmt.Println("内容为空")

		return
	}

	if keyword == "" {

		//log.Println(contentSelector)

		//log.Panicln("内容为空")

		//fmt.Println("关键字为空")

		keyword = title

		//return
	}

	doc, err = goquery.NewDocumentFromReader(strings.NewReader(content))

	if err != nil {
		log.Fatal(err)
	}

	//var array[]string

	var waitImg sync.WaitGroup

	//保存图片列表
	var imgSaveList []string

	//图片原数据
	var imgOriginalList []string

	var lock sync.Mutex

	imgTotal := 0

	doc.Find("img").Each(func(i int, selection *goquery.Selection) {

		url, isFind := selection.Attr("src")

		//panic(url)

		if isFind {

			imgTotal++

			waitImg.Add(1)

			//if config.imgSrcIsAbsolute==2 {
			//
			//	url=config.host+url
			//}

			//开启协程下载图片
			go downImg(url, &waitImg, &imgSaveList, &lock, &imgOriginalList, config)
		}

	})

	waitImg.Wait()

	//panic(imgTotal)

	//如果这个文章图片为0，就放弃这个文章
	if len(imgOriginalList) == 0 {

		return
	}

	//panic(imgSaveList)

	for i, v := range imgOriginalList {

		if i == 0 {

			img = imgSaveList[i]
		}

		//替换图片中的新链接
		content = strings.Replace(content, v, ImagePrefix+imgSaveList[i], -1)

	}

	task := Task{
		//Id:1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		//CategoryId: categoryId,
		Content:   content,
		Img:       img,
		Title:     title,
		Desc:      desc,
		Keyword:   keyword,
		WriteType: 2,
	}

	tasks <- task

	fmt.Println(title, "---------", config.pageIndex)

}

//下载内容中的图片
func downImg(imgUrl string, waitImg *sync.WaitGroup, imgSaveList *[]string, lock *sync.Mutex, imgOriginalList *[]string, config Config) {

	defer waitImg.Done()

	defer func() {

		if r := recover(); r != nil {
			fmt.Printf("downImg捕获到的错误：%s\n", r)
		}

	}()

	realUrl := imgUrl

	realUrl = getHref(imgUrl, config.host)

	lock.Lock()
	i, _ := rand.Int(rand.Reader, big.NewInt(100))

	dir := config.dirs + "/" + i.String() + "" + time.Now().Format("20060102")

	is, _ := PathExists("static/" + dir)

	if is == false {

		err := os.Mkdir("static/"+dir, os.ModePerm)

		if err != nil {

			log.Println(err)
		}

		//fmt.Println(dir)

	}
	lock.Unlock()

	uuids := uuid.NewV4()

	var err error
	//生成文件名并去掉横杠
	filename := strings.Replace(uuid.Must(uuids, err).String(), "-", "", -1)

	err = tools.DownloadImage(realUrl, "static/"+dir+"/"+filename+".jpg")

	if err != nil {

		fmt.Println(err)

		return

	}

	//加锁
	lock.Lock()
	*imgSaveList = append(*imgSaveList, dir+"/"+filename+".jpg")
	*imgOriginalList = append(*imgOriginalList, imgUrl)
	lock.Unlock()

}

//Excel表格写入协程
func writeExcel() {

	row := 2

	//设置表头
	f.SetCellValue("Sheet1", "A1", "title")
	f.SetCellValue("Sheet1", "B1", "keyword")
	f.SetCellValue("Sheet1", "C1", "desc")
	f.SetCellValue("Sheet1", "D1", "img")
	f.SetCellValue("Sheet1", "E1", "content")
	//f.SetCellValue("Sheet1", "K1", "keyword")

	for v := range tasks {

		f.SetCellValue("Sheet1", "A"+strconv.Itoa(row), v.Title)
		f.SetCellValue("Sheet1", "B"+strconv.Itoa(row), v.Keyword)
		f.SetCellValue("Sheet1", "C"+strconv.Itoa(row), v.Desc)
		f.SetCellValue("Sheet1", "D"+strconv.Itoa(row), v.Img)
		f.SetCellValue("Sheet1", "E"+strconv.Itoa(row), v.Content)
		//f.SetCellValue("Sheet1", "K"+strconv.Itoa(row), v.Keyword)

		row++

	}

}

// 判断文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func replaceSpace(s string) string {
	str := ""
	for _, value := range s {
		if value == ' ' { //也可以写成if value == 32
			str += "%20" //替换空格
		} else {
			str += string(value) //把value转为string追加到str中
		}
	}
	return str
}

//获取正常的链接
func getHref(href string, host string) string {

	case1, _ := regexp.MatchString("^/[a-zA-Z0-9_]+.*", href)

	case2, _ := regexp.MatchString("^//[a-zA-Z0-9_]+.*", href)

	case3, _ := regexp.MatchString("^(http|https).*", href)

	switch true {

	case case1:

		href = host + href

	case case2:

		//获取当前网址的协议
		res := regexp.MustCompile("^(https|http).*").FindStringSubmatch(host)

		href = res[1] + ":" + href

	case case3:

	}

	return href

}

func getDesc(doc *goquery.Document, config Config) string {

	desc, exists := doc.Find("meta[name=\"description\"]").Attr("content")

	if exists == true {

		return desc
	}

	temp := []rune(strings.Replace(strings.Replace(doc.Find(config.contentSelector).Text(), "　", "", -1), " ", "", -1))

	length := len(temp)

	sub := 65

	if sub > length {

		sub = length
	}

	return string(temp[:sub])

}
