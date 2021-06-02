package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/tidwall/gjson"
)

const (
	version         = "v0.1.1.1" //版本号依次表示重大改变、任务内容改变、任务列表改变、配置改变
	ok              = "OK!"
	dbpath          = "./data/task.db"
	tasktablecreate = `CREATE TABLE "_table" (
	"title" MEDIUMTEXT NULL,
	"info" MEDIUMTEXT NOT NULL,
	"date" MEDIUMTEXT NOT NULL,
	"status" MEDIUMTEXT NOT NULL
	);
	CREATE  TRIGGER max__table_trigger AFTER INSERT
ON "_table"
BEGIN
delete from "_table" where (select count(date) from "_table" )> %G and date in (select date from "_table" order by date desc limit (select count(date) from "_table") offset %G);
END;
	`
	config_table = `CREATE TABLE "confinfo" (
		"id" integer PRIMARY KEY,
		"runtime" INTEGER NOT NULL,
		"runnum" INTEGER NOT NULL,
		"runstep" INTEGER NOT NULL,
		"email" MEDIUMTEXT NULL,
		"emailpass" MEDIUMTEXT NULL,
		"emailhost" MEDIUMTEXT NULL,
		"emailto" MEDIUMTEXT NULL,
		"wxid" MEDIUMTEXT NULL,
		"wxsecret" MEDIUMTEXT NULL,
		"wxtid" MEDIUMTEXT NULL,
		"wxto" MEDIUMTEXT NULL,
		"msgwav" MEDIUMTEXT NULL,
		"msgformat" MEDIUMTEXT NOT NULL,
		"softopen" INTEGER NOT NULL,
		"softmin" INTEGER NOT NULL,
		"softclose" INTEGER NOT NULL,
		"version" MEDIUMTEXT NOT NULL
	   );`
	task_table = `CREATE TABLE "task" (
		"id" integer PRIMARY KEY autoincrement,
		"title" MEDIUMTEXT NULL,
		"address" MEDIUMTEXT NOT NULL,
		"ishight" INTEGER NOT NULL,
		"cookie" MEDIUMTEXT NULL,
		"csschoose" MEDIUMTEXT NULL,
		"xpathchoose" MEDIUMTEXT NULL,
		"type" MEDIUMTEXT NOT NULL,
		"con" MEDIUMTEXT NOT NULL,
		"istext" MEDIUMTEXT NOT NULL,
		"timestart" MEDIUMTEXT NULL,
		"timeend" MEDIUMTEXT NULL,
		"timestep" MEDIUMTEXT NOT NULL,
		"ismessgewx" INTEGER NOT NULL,
		"ismessgewin" INTEGER NOT NULL,
		"ismessgeemail" INTEGER NOT NULL,
		"history" MEDIUMTEXT NOT NULL,
		"historynum" INTEGER NOT NULL,
		"addtime" MEDIUMTEXT NOT NULL,
		"lastruntime" MEDIUMTEXT NOT NULL,
		"lastchangetime" MEDIUMTEXT NOT NULL,
		"lastmsgtime" MEDIUMTEXT NOT NULL,
		"nextruntime" MEDIUMTEXT NOT NULL,
		"status" MEDIUMTEXT NOT NULL
	  );
	  `
	config_init   = `INSERT INTO confinfo values(1,5,5,1,"","","","","","","","","default","{{标题}} 已更新\n 此次内容：{{全文}}\n 上次内容：{{上文}}",0,0,0,"` + version + `");`
	historyinsert = `INSERT INTO "%d"(title,info,date,status) values(?,?,?,?)`
	taskupdate    = ""
	taskdelete    = ""
	mintime       = "2006-01-02 15:04:05"
	maxtime       = "2406-01-02 15:04:05"
)

var db *sql.DB

func dbinit() {
	if !PathExists(dbpath) {
		os.Create(dbpath)
		db, _ = sql.Open("sqlite3", dbpath)
		db.Exec(config_table)
		db.Exec(config_init)
		db.Exec(task_table)
		_config = SelectToMaps("SELECT * FROM confinfo where id=1")[0]
	} else {
		db, _ = sql.Open("sqlite3", dbpath)
		_config = SelectToMaps("SELECT * FROM confinfo where id=1")[0]
		if _config["version"].(string) != version {
			oldversion := strings.Split(_config["version"].(string), ".")
			newversion := strings.Split(version, ".")
			if oldversion[3] != newversion[3] {
				db.Exec("drop table 'confinfo'")
				db.Exec(config_table)
				temp, _ := db.Query("select * from confinfo")
				newconfig, _ := temp.Columns()
				for t, _ := range _config {

					flag := false
					for _, new := range newconfig {
						if new == t {
							flag = true
							break
						}
					}
					if !flag {
						delete(_config, t)
					}
				}
				temp.Close()
				MapToInsert("confinfo", _config)
				_config = SelectToMaps("SELECT * FROM confinfo where id=1")[0]
			}
			if oldversion[2] != newversion[2] {
				tasks := SelectToMaps("SELECT * FROM task")
				db.Exec("drop table 'task'")
				db.Exec(task_table)
				temp, _ := db.Query("select * from task")
				newtask, _ := temp.Columns()
				for _, t := range tasks {
					for k, _ := range t {
						flag := false
						for _, new := range newtask {
							if new == k {
								flag = true
								break
							}
						}
						if !flag {
							delete(t, k)
						}

					}
					MapToInsert("task", t)
				}
				temp.Close()
			}
			if oldversion[1] != newversion[1] {
				//TODO
			}
		}
	}
	_config = SelectToMaps("SELECT * FROM confinfo where id=1")[0]
}
func taskUpdate(table, injson, where string) string {
	timestart := gjson.Get(injson, "timeline.0").Time()
	timeend := gjson.Get(injson, "timeline.1").Time()
	var timestarts, timeends string
	if timestart.After(timeend) || timestart == timeend {
		timestarts = mintime
		timeends = maxtime
	} else {
		timestarts = timestart.Format(mintime)
		timeends = timeend.Format(mintime)
	}
	_map := make(map[string]interface{})
	json.Unmarshal([]byte(injson), &_map)
	delete(_map, "timeline")
	_map["timestart"] = timestarts
	_map["timeend"] = timeends
	return MapToUpdate(table, where, _map)
}
func taskInsert(injson string) string {
	timestart := gjson.Get(injson, "timeline.0").Time()
	timeend := gjson.Get(injson, "timeline.1").Time()
	var timestarts, timeends string
	if timestart.After(timeend) || timestart == timeend {
		timestarts = mintime
		timeends = maxtime
	} else {
		timestarts = timestart.Format(mintime)
		timeends = timeend.Format(mintime)
	}
	_map := make(map[string]interface{})
	json.Unmarshal([]byte(injson), &_map)
	delete(_map, "timeline")
	_map["timestart"] = timestarts
	_map["timeend"] = timeends
	_map["addtime"] = time.Now().Format(mintime)
	_map["lastruntime"] = time.Now().Format(mintime)
	_map["lastchangetime"] = time.Now().Format(mintime)
	_map["lastmsgtime"] = time.Now().Format(mintime)
	_map["nextruntime"] = nextRuntime(timestarts, timeends, gjson.Get(injson, "timestep").String())
	if _map["nextruntime"] == time.Now().Format(mintime) {
		_map["status"] = "overdue"
	} else {
		_map["status"] = "sleep"
	}
	mes1, title := testRun(injson)
	if gjson.Get(injson, "title").String() != "" {
		title = gjson.Get(injson, "title").String()
	} else {
		_map["title"] = title
	}
	res, s := MapToInsert("task", _map)
	if res != ok {
		return res
	}
	id, _ := s.LastInsertId()
	_, err := db.Exec(fmt.Sprintf(strings.ReplaceAll(tasktablecreate, "_table", strconv.FormatInt(id, 10)), _map["historynum"], _map["historynum"]))
	if err != nil {
		return err.Error()
	}

	db.Exec(fmt.Sprintf(historyinsert, id), title, mes1, time.Now().Format(mintime), "First Request")
	return ok
}
func taskDelete(id string) string {
	_, err := db.Exec("DELETE FROM task WHERE id=" + id)
	if err != nil {
		return err.Error()
	}
	_, err = db.Exec("drop table \"" + id + "\"")
	if err != nil {
		return err.Error()
	}
	return ok
}

func SelectToJsons(qur string) []string {
	ret := []string{}
	_maps := SelectToMaps(qur)
	for i := 0; i < len(_maps); i++ {
		mjson, _ := json.Marshal(_maps[i])
		ret = append(ret, string(mjson))
	}
	return ret
}
func SelectToMaps(qur string) []map[string]interface{} {
	var ret []map[string]interface{}
	res, _ := db.Query(qur)
	Col, _ := res.Columns()
	a := make([]interface{}, len(Col))
	b := make([]interface{}, len(Col))
	for key := range a {
		b[key] = &a[key]
	}
	for res.Next() {
		_map := make(map[string]interface{})
		res.Scan(b...)
		for key, value := range Col {
			_map[value] = a[key]
		}
		ret = append(ret, _map)
	}
	return ret
}
func JsonToInsert(table, injson string) (string, sql.Result) {
	_map := make(map[string]interface{})
	json.Unmarshal([]byte(injson), &_map)
	return MapToInsert(table, _map)
}
func JsonToUpdate(table, injson, where string) string {
	_map := make(map[string]interface{})
	json.Unmarshal([]byte(injson), &_map)
	return MapToUpdate(table, where, _map)
}
func MapToInsert(table string, _map map[string]interface{}) (string, sql.Result) {
	keys := make([]string, 0, len(_map))
	values := make([]interface{}, 0, len(_map))
	for k := range _map {
		keys = append(keys, k)
		values = append(values, _map[k])
	}
	_sql := "INSERT INTO \"" + table + "\" (" + strings.Join(keys, ",") + ") VALUES (?" + strings.Repeat(",?", len(keys)-1) + ")"
	res, err := db.Exec(_sql, values...)
	if err != nil {
		return err.Error(), res
	}
	return ok, res
}
func MapToUpdate(table, where string, _map map[string]interface{}) string {
	keys := make([]string, 0, len(_map))
	values := make([]interface{}, 0, len(_map))
	for k := range _map {
		keys = append(keys, k)
		values = append(values, _map[k])
	}
	v := keys[0] + "=?"
	for count := 1; count < len(keys); count++ {
		v += "," + keys[count] + "=?"
	}
	_sql := "UPDATE \"" + table + "\" SET " + v + " WHERE " + where
	_, err := db.Exec(_sql, values...)
	if err != nil {
		return err.Error()
	}
	return ok
}
