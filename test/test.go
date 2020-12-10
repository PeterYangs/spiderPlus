package main

import "github.com/PeterYangs/spiderPlus"

func main() {

	//spiderPlus.Rule(
	//	"https://www.925g.com",
	//	"/zixun_page[PAGE].html",
	//	2,
	//	1,
	//	"body > div.ny-container.uk-background-default > div.wrap > div > div.commonLeftDiv.uk-float-left > div > div.bdDiv > div > ul > li",
	//	"a",
	//	"body > div.ny-container.uk-background-default > div.wrap > div > div.commonLeftDiv.uk-float-left > div > div.articleDiv > div.hd > div.title",
	//	"body > div.ny-container.uk-background-default > div.wrap > div > div.commonLeftDiv.uk-float-left > div > div.articleDiv > div.bd",
	//	"test",
	//	"api/uploads/",
	//
	//	)

	spiderPlus.Rule(
		"https://www.azg168.cn",
		"/bazisuanming/index_[PAGE].html",
		10,
		2,
		"body > div.main.clearfix.w960 > div.main_left.fl.dream_box > ul > li",
		"a",
		"body > div.main.clearfix.w960 > div.main_left.fl > div.art_con_left > h1",
		"#azgbody",
		"demo",
		"test2",
	)

}
