package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	//"net/http"
	//_ "net/http/pprof"
	"path/filepath"
	"strings"
	"time"

	"github.com/lxn/win"
	"github.com/sciter-sdk/go-sciter"
	"github.com/sciter-sdk/go-sciter/window"
)

var _w *window.Window
var _setInfo map[string](string)
var config map[string]interface{}

func main() {
	if os.Getenv("VSCODE_AMD_ENTRYPOINT") == "" {
		InitPanicFile()
	}
	dbinit()
	go taskRun_go()
	//go http.ListenAndServe("0.0.0.0:8081", nil)
	//screen init
	cxScreen := win.GetSystemMetrics(win.SM_CXSCREEN)
	cyScreen := win.GetSystemMetrics(win.SM_CYSCREEN)
	var formWidth, formHeight int32 = 1200, 750
	var fromPosLeft, formPosTop int32
	fromPosLeft = (cxScreen - formWidth) / 2
	formPosTop = (cyScreen - formHeight) / 2
	_w, _ = window.New(
		sciter.SW_MAIN|sciter.SW_RESIZEABLE,
		&sciter.Rect{fromPosLeft, formPosTop, fromPosLeft + formWidth, formPosTop + formHeight})
	//load file
	_w.SetCallback(&sciter.CallbackHandler{
		OnLoadData: func(params *sciter.ScnLoadData) int {
			if strings.HasPrefix(params.Uri(), "file://") {
				fileData := _w.GetArchiveItem(params.Uri()[7:])
				_w.DataReady(params.Uri()[7:], fileData)
			}
			return 0
		},
	})
	//app run
	fullpath, err := filepath.Abs("./app/index.html")
	if err != nil {
		log.Panic(err)
	}
	if err := _w.LoadFile(fullpath); err != nil {
		log.Panic(err)
	}
	setEventHandler()
	if config["softmin"].(int64) == 0 {
		_w.Show()
	} else {
		_w.Call("ChangeWindow")
	}
	_w.Run()
}
func setEventHandler() {
	_w.DefineFunction("AddTask", func(args ...*sciter.Value) *sciter.Value {
		go taskInsert_go(args[0].String())
		return sciter.NewValue("")
	})
	_w.DefineFunction("TestNet", func(args ...*sciter.Value) *sciter.Value {
		go testRun_go(args[0].String())
		return sciter.NewValue("")
	})
	_w.DefineFunction("GetConfig", func(args ...*sciter.Value) *sciter.Value {
		return sciter.NewValue(SelectToJsons("select * from confinfo where id=1")[0])
	})
	_w.DefineFunction("ConfigUp", func(args ...*sciter.Value) *sciter.Value {
		go configUpdate_go(args[0].String())
		return sciter.NewValue("")
	})
	_w.DefineFunction("GetSound", func(args ...*sciter.Value) *sciter.Value {
		a := sciter.NewValue()
		files, _ := ioutil.ReadDir("./sound")
		for _, file := range files {
			if file.IsDir() {
				continue
			} else {
				a.Append((file.Name()))
			}
		}
		return a
	})
	_w.DefineFunction("GetTask", func(args ...*sciter.Value) *sciter.Value {
		_sql := "select * from task order by lastchangetime DESC"
		tasks := SelectToMaps(_sql)
		ret := sciter.NewValue()
		for _, task := range tasks {
			r1 := sciter.NewValue()
			for key, value := range task {
				r1.Set(key, value)
			}
			ret.Append(r1)
		}
		return ret
	})
	_w.DefineFunction("GetTaskJson", func(args ...*sciter.Value) *sciter.Value {
		_sql := "select * from task where id=" + args[0].String()
		return sciter.NewValue(SelectToJsons(_sql)[0])
	})
	_w.DefineFunction("GetTaskHis", func(args ...*sciter.Value) *sciter.Value {
		tasks := SelectToMaps("select * from \"" + args[0].String() + "\" order by date DESC")
		ret := sciter.NewValue()
		for _, task := range tasks {
			r1 := sciter.NewValue()
			for key, value := range task {
				r1.Set(key, value)
			}
			ret.Append(r1)
		}
		return ret
	})
	_w.DefineFunction("UpdateTask", func(args ...*sciter.Value) *sciter.Value {
		_w.Call("MessgeBox", sciter.NewValue(taskUpdate("task", args[0].String(), "id="+args[1].String())))
		return sciter.NewValue()
	})
	_w.DefineFunction("DeleteTask", func(args ...*sciter.Value) *sciter.Value {
		_w.Call("MessgeBox", sciter.NewValue(taskDelete(args[0].String())))
		return sciter.NewValue()
	})
	_w.DefineFunction("RefTask", func(args ...*sciter.Value) *sciter.Value {

		return sciter.NewValue()
	})
}
func testRun_go(injson string) {
	mes, _ := testRun(injson)
	mes = StringMax(mes, 500)
	_w.Call("MessgeBox", sciter.NewValue(mes))
	_w.Call("loadhide")
}
func taskInsert_go(injson string) {
	_w.Call("MessgeBox", sciter.NewValue(taskInsert(injson)))
	_w.Call("loadhide")
}

var nowTask int64 = 0

func taskRun_go() {
	for {
		if nowTask < (config["runstep"].(int64)) {
			_map := SelectToMaps("SELECT * FROM task  where nextruntime <=strftime('%Y-%m-%d %H:%M:%S','now','localtime') and timeend>=strftime('%Y-%m-%d %H:%M:%S','now','localtime') ORDER BY nextruntime")
			for _, task := range _map {
				temp := make(map[string]interface{})
				temp["lastruntime"] = time.Now().Format(mintime)
				temp["nextruntime"] = nextRuntime(task["timestart"].(string), task["timeend"].(string), task["timestep"].(string))
				temp["status"] = "running"
				MapToUpdate("task", "id="+fmt.Sprintf("%v", task["id"]), temp)
				go runtask(task)
				nowTask++
			}
		}
		time.Sleep(time.Second * time.Duration(config["runstep"].(int64)))
	}
}
func runtask(task map[string]interface{}) {
	temp := make(map[string]interface{})
	his := make(map[string]interface{})
	his["info"], his["title"] = Run(fmt.Sprintf("%v", task["ishight"]), task["address"].(string), task["cookie"].(string), task["csschoose"].(string), task["xpathchoose"].(string))
	his["date"] = time.Now().Format(mintime)
	lastinfo := SelectToMaps("select info from \"" + fmt.Sprintf("%v", task["id"]) + "\" order by date DESC LIMIT 1")[0]
	_ischange, _ismsg := lastinfo["info"].(string) == his["info"].(string), false
	_ismsg, his["status"] = comparehis(lastinfo["info"].(string), his["info"].(string), task)
	if task["history"] == "runtime" || (_ischange && (task["history"] == "changetime")) || (_ismsg && (task["history"] == "messgetime")) {
		MapToInsert(fmt.Sprintf("%v", task["id"]), his)
	}
	if _ischange {
		temp["lastchangetime"] = time.Now().Format(mintime)
	}
	if _ismsg {
		temp["lastmsgtime"] = time.Now().Format(mintime)
		if task["ismessgewin"].(int64) == 1 {
			msgwin(lastinfo["info"].(string), his["info"].(string), task)
		}
		if task["ismessgeemail"].(int64) == 1 {
			msgemail(lastinfo["info"].(string), his["info"].(string), task)
		}
		if task["ismessgesoft"].(int64) == 1 {
			//TODO
		}
	}
	temp["status"] = "sleep"
	MapToUpdate("task", "id="+fmt.Sprintf("%v", task["id"]), temp)
	nowTask--
}
func configUpdate_go(injson string) {
	var lastconfig = config
	msg := JsonToUpdate("confinfo", injson, "id=1")
	config = SelectToMaps("SELECT * FROM confinfo where id=1")[0]

	if lastconfig["softopen"] != config["softopen"] {

		if config["softopen"].(int64) == 0 {
			deletestartup()
		} else {
			openstartup()
		}

	}
	if lastconfig["softclose"] != config["softclose"] {
		if config["softclose"].(int64) == 0 {
			winf, _ := ioutil.ReadFile("./app/window.tis")
			winf = []byte(strings.Replace(string(winf), "ChangeWindow();//close", "CloseWindow();//close", 1))
			ioutil.WriteFile("./app/window.tis", winf, 0666)
		} else {
			winf, _ := ioutil.ReadFile("./app/window.tis")
			winf = []byte(strings.Replace(string(winf), "CloseWindow();//close", "ChangeWindow();//close", 1))
			ioutil.WriteFile("./app/window.tis", winf, 0666)
		}
	}
	_w.Call("MessgeBox", sciter.NewValue(msg))
	_w.Call("loadhide")

}
