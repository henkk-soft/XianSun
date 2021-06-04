![img](https://imgur.com/eekqr7P.png)

---
### 项目说明

golang版本1.14或1.15，1.16应该不支持，其他版本未做测试。

sciter tis以及go-sciter请选用对应版本，相应的sciter.dll文件需要做替换，这里的sciter.dll用的是4.4.7 版本，go-sciter请用sciter-api分支


---
### 编译与发布

图标,高DPI等设置(已经存在) 

`rsrc -manifest main.exe.manifest -ico icon.ico -o main.syso`


编译

 `go build -o main.exe  -ldflags "-H windowsgui"`
 

运行文件：
只需要文件夹`app`,`data`,`img`,编译exe文件以及sciter.dll文件即可

 ---
 软件说明
                      
### [基本说明](./doc/基本说明.md)

### [通知说明](./doc/通知说明.md)

### [更新记录](./doc/更新记录.md)

---

参考链接：

https://sciter.com/

https://github.com/sciter-sdk/go-sciter/tree/sciter-api

https://github.com/PuerkitoBio/goquery

https://github.com/antchfx/htmlquery

https://github.com/chromedp/chromedp

https://github.com/tidwall/gjson

https://github.com/mattn/go-sqlite3

---

名字(XianSun)由来：长傲宾于柏谷,妻睹貌而献飧--《西征赋》