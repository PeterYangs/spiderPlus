# spiderPlus
爬取网站文章到Excel


安装

go get github.com/PeterYangs/spiderPlus


使用

```go

import "github.com/PeterYangs/spiderPlus"

spiderPlus.Rule(
"https://www.azg168.cn",//域名
"/bazisuanming/index_[PAGE].html",//栏目，分页用[PAGE]替代
10,//爬取页数
2,//起始页面
"body > div.main.clearfix.w960 > div.main_left.fl.dream_box > ul > li",//列表选择器
"a",//a链接选择器(相对于列表)
"body > div.main.clearfix.w960 > div.main_left.fl > div.art_con_left > h1",//标题选择器
"#azgbody",//内容选择器
"demo",//下载图片路径
"ttt",//内容中的图片前缀
)
```
