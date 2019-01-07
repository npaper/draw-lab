package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"model"
	"net/http"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

func getConfig() (map[string]interface{}, string, error) {
	argNum := len(os.Args)
	if argNum <= 1 {
		return nil, "", errors.New("缺少参数：configPath")
	}
	fileName := os.Args[1]
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err == nil {
		fileName = dir + "/" + fileName
	}
	configData := make(map[string]interface{})
	config, err := ioutil.ReadFile(fileName)

	if err != nil {
		fmt.Println(err)
		return nil, dir, errors.New("读取配置文件失败")
	}

	err = yaml.Unmarshal(config, configData)
	if err != nil {
		fmt.Println(err)
		return nil, dir, errors.New("读取配置文件失败")
	}

	return configData, dir, nil
}

func main() {

	configData, dir, err := getConfig()

	if err != nil {
		fmt.Println(err)
		return
	}

	InitDB(Trans(configData["mysql"]).(map[string]interface{}))

	config := make(map[string]func(*HTTPServer))
	configS := make(map[string]string)

	config["login"] = func(s *HTTPServer) {
		s.SetMsg("ok")
	}

	config["push"] = func(s *HTTPServer) {
		name := s.ObtainStringParam("name", "")
		r := s.ObtainStringParam("r", "")
		g := s.ObtainStringParam("g", "")
		b := s.ObtainStringParam("b", "")
		tags := s.ObtainStringParam("tags", "")
		defualtSet := s.ObtainStringParam("defualt_set", "")

		if name == "" {
			s.Error(-1, "名称不能为空")
			return
		}

		if r == "" || g == "" || b == "" {
			s.Error(-1, "表达式不能为空")
			return
		}

		paper := model.NewWallpaper()
		paper.Name = name
		paper.R = r
		paper.G = g
		paper.B = b
		paper.Tags = tags
		paper.DefaultSet = defualtSet

		id, err := paper.Insert(GetConn())

		if err != nil {
			s.Error(-1, err.Error())
		} else {
			s.SetMsg(id)
		}
		s.SetMsg("ok")
	}

	config["zang"] = func(s *HTTPServer) {
		id := s.ObtainInt64Param("id", 0)
		num, err := model.ZangWallpaper(GetConn(), id)
		if err != nil {
			s.Error(-1, err.Error())
		} else {
			s.SetMsg(num)
		}
	}

	config["list"] = func(s *HTTPServer) {
		page := s.ObtainIntParam("page", 1)
		limit := s.ObtainIntParam("limit", 10)

		if page < 1 {
			page = 1
		}
		if limit < 5 {
			limit = 5
		}
		if limit > 50 {
			limit = 50
		}

		list, err := model.ListWallpaper(GetConn(), (page-1)*limit, page*limit)

		if err != nil {
			s.Error(-1, err.Error())
		} else {
			for _, item := range list {
				s.InsertResult(item)
			}
		}
	}

	serverConfig := Trans(configData["server"]).(map[string]interface{})

	if static, ok := serverConfig["static"]; ok {
		configS["/"] = dir + "/" + static.(string)
	}

	Gets(config, "wp")
	Statics(configS, "")
	port := serverConfig["port"]

	log.Printf("服务器启动，端口为 :%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
