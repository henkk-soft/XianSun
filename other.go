package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-toast/toast"
	"github.com/tidwall/gjson"
	"golang.org/x/sys/windows/registry"
)

const (
	kernel32dll = "kernel32.dll"
)

func InitPanicFile() error {
	panicFile := "./data/error.log"
	file, err := os.OpenFile(panicFile, os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	kernel32 := syscall.NewLazyDLL(kernel32dll)
	setStdHandle := kernel32.NewProc("SetStdHandle")
	sh := syscall.STD_ERROR_HANDLE
	v, _, err := setStdHandle.Call(uintptr(sh), uintptr(file.Fd()))
	if v == 0 {
		return err
	}
	return nil
}
func stringMax(str string, maxcount int) string {
	if len(str) > maxcount {
		str = str[0:maxcount] + "..."
	}
	return str
}
func nextRuntime(start, end, step string) string {
	_start, _ := time.Parse(mintime, start)
	_end, _ := time.Parse(mintime, end)
	dur, err := time.ParseDuration(step)
	if err != nil {
		return maxtime
	}
	_next := time.Now().Add(dur)
	if time.Now().After(_end) {
		return maxtime
	} else if _start.After(time.Now()) {
		return start
	} else if _next.After(_end) {
		return end
	} else {
		return _next.Format(mintime)
	}
}
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
func standardizeSpaces(s string) string {
	ss := strings.Split(s, "\n")
	res := ""
	for _, n := range ss {
		n = strings.TrimSpace(n)
		if n != "" {
			res += n + "\n"
		}
	}
	return strings.Trim(res, "\n")
}
func openstartup() {
	path, _ := os.Executable()
	key, _, _ := registry.CreateKey(registry.CURRENT_USER, `SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Run`, registry.ALL_ACCESS)
	key.SetStringValue("WEBMONITORGO", path)
	key.Close()
}
func deletestartup() {
	key, _, _ := registry.CreateKey(registry.CURRENT_USER, `SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Run`, registry.ALL_ACCESS)

	err := key.DeleteValue("WEBMONITORGO")
	if err != nil {
		log.Println(err.Error())
	}
	key.Close()
}
func comparehis(old, new string, task map[string]interface{}) (bool, string) {
	if task["type"] == "str" {
		if old != new {
			if len(old) < len(new) {
				if task["con"] == "ischange" {
					return true, "文字变化"
				} else if task["con"] == "isadd" {
					return true, "文字增多"
				} else if task["con"] == "ismut" {
					return false, "文字增多"
				}
			} else {
				if task["con"] == "ischange" {
					return true, "文字变化"
				} else if task["con"] == "isadd" {
					return false, "文字减少"
				} else if task["con"] == "ismut" {
					return true, "文字减少"
				}
			}
		} else {
			return false, "文字无变化"
		}
	} else if task["type"] == "num" {
		v1, err := strconv.ParseFloat(old, 32)
		if err != nil {
			return false, "上次获取非数字！"
		}
		v2, err := strconv.ParseFloat(new, 32)
		if err != nil {
			return false, "非数字！"
		}
		if v1 != v2 {
			if v1 < v2 {
				if task["con"] == "ischange" {
					return true, "数值增大"
				} else if task["con"] == "isadd" {
					return true, "数值增大"
				} else if task["con"] == "ismut" {
					return false, "数值增大"
				}
			} else {
				if task["con"] == "ischange" {
					return true, "数值减小"
				} else if task["con"] == "isadd" {
					return false, "数值减小"
				} else if task["con"] == "ismut" {
					return true, "数值减小"
				}
			}
		} else {
			return false, "数值无变化"
		}

	}
	return false, "not change"
}
func msgreplace(old, new string, task map[string]interface{}) string {
	str := strings.ReplaceAll(_config["msgformat"].(string), "\\n", "\n")
	str = strings.ReplaceAll(str, "{{标题}}", task["title"].(string))
	str = strings.ReplaceAll(str, "{{全文}}", new)
	str = strings.ReplaceAll(str, "{{上文}}", old)
	return str
}
func msgwin(old, new string, task map[string]interface{}) {
	str := msgreplace(old, new, task)
	a, _ := toast.Audio(_config["msgwav"].(string))
	notification := toast.Notification{
		AppID:   "Microsoft.Windows.Shell.RunDialog",
		Title:   task["title"].(string),
		Message: str,
		Actions: []toast.Action{
			{"protocol", "点击进入网站", task["address"].(string)},
		},
		Audio: a,
	}
	err := notification.Push()
	if err != nil {
		log.Println(err)
	}

}
func msgemail(old, new string, task map[string]interface{}) {
	if _config["email"].(string) == "" || _config["emailhost"].(string) == "" || _config["emailto"].(string) == "" {
		log.Println("请完善邮箱设置")
		return
	}
	SendToMail(task["title"].(string), msgreplace(old, new, task)+"\n"+task["address"].(string))
}
func msgwexin(old, new string, task map[string]interface{}) {
	if _config["wxid"].(string) == "" || _config["wxsecret"].(string) == "" || _config["wxtid"].(string) == "" || _config["wxto"].(string) == "" {
		log.Println("请完善微信设置")
		return
	}
	str := strings.ReplaceAll(_config["msgformat"].(string), "{{标题}}", task["title"].(string))
	str = strings.ReplaceAll(str, "{{全文}}", new)
	str = strings.ReplaceAll(str, "{{上文}}", old)
	wxSend(stringMax(str, 200), task["address"].(string))
}

var wxtoken string = ""

func wxGetToken() string {
	if wxtoken == "" {
		req := getRun("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=" + _config["wxid"].(string) + "&secret=" + _config["wxsecret"].(string))
		if gjson.Get(req, "errcode").String() != "" {
			return gjson.Get(req, "errmsg").String()
		} else {
			wxtoken = gjson.Get(req, "access_token").String()
		}
	} else {
		req := getRun("https://api.weixin.qq.com/cgi-bin/template/get_all_private_template?access_token=" + wxtoken)
		if gjson.Get(req, "errcode").String() != "" {
			req = getRun("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=" + _config["wxid"].(string) + "&secret=" + _config["wxsecret"].(string))
			if gjson.Get(req, "errcode").String() != "" {
				return gjson.Get(req, "errmsg").String()
			}
		}
	}
	return ok
}
func wxSend(msg, url string) string {
	if gettoken := wxGetToken(); gettoken != ok {
		return gettoken
	}
	users := strings.Split(_config["wxto"].(string), ";")
	for _, user := range users {
		post := `{
			"touser":"` + user + `",
			"template_id":"` + _config["wxtid"].(string) + `",
			"url":"` + url + `",
			"data":{
					"content": {
						"value":"` + msg + `",
						"color":"#0092ff"
					}
			}
		}`
		req := postRun("https://api.weixin.qq.com/cgi-bin/message/template/send?access_token="+wxtoken, post, "application/json")
		if gjson.Get(req, "errcode").String() != "" {
			log.Println(gjson.Get(req, "errmsg").String())
		}
	}

	return ok
}
