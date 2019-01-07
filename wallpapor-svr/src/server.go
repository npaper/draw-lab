package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	yaml "gopkg.in/yaml.v2"
)

// HTTPServer ...
type HTTPServer struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	message        *Message
	Params         map[string][]string
}

// Message ...
type Message struct {
	Result    []interface{}          `json:"result"`
	Data      map[string]interface{} `json:"data"`
	Msg       interface{}            `json:"msg"`
	ErrorCode int                    `json:"errorCode"`
}

// NewHTTPServer ...
func NewHTTPServer(w http.ResponseWriter, r *http.Request) *HTTPServer {
	server := &HTTPServer{ResponseWriter: w, Request: r}
	server.ResponseWriter.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	server.ResponseWriter.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	server.ResponseWriter.Header().Set("content-type", "application/json")             //返回数据格式是json
	message := &Message{}
	message.ErrorCode = 0
	// message.Data = make(map[string]interface{})
	// message.Result = make([]interface{}, 0)
	// message.Msg = "ok"

	r.ParseForm()
	server.Params = r.Form
	server.message = message
	return server
}

// Trans 转换map
func Trans(data interface{}) interface{} {
	if data == nil {
		return nil
	}
	data0, ok := data.(map[interface{}]interface{})
	if !ok {
		data1, ok := data.([]interface{})
		if ok {
			arr := make([]interface{}, 0)
			for _, c := range data1 {
				_v := Trans(c)
				if _v == nil {
					arr = append(arr, c)
				} else {
					arr = append(arr, _v)
				}
			}
			return arr
		}
		return nil
	}
	m := make(map[string]interface{})
	for k, v := range data0 {
		k0, ok := k.(string)
		if ok {
			_v := Trans(v)
			if _v == nil {
				m[k0] = v
			} else {
				m[k0] = _v
			}
		}
	}
	return m
}

func toString(data interface{}, server *HTTPServer) {
	if data == nil {
		return
	}

	switch data.(type) {
	case string:
		// server.SetMsg(data.(string))
		server.SetMsg(data)
	case int:
		// server.SetMsg(strconv.Itoa(data.(int)))
		server.SetMsg(data)
	case int64:
		// server.SetMsg(strconv.FormatInt(data.(int64), 10))
		server.SetMsg(data)
	case float32:
		// server.SetMsg(strconv.FormatFloat(float64(data.(float32)), 'f', 6, 64))
		server.SetMsg(data)
	case float64:
		// server.SetMsg(strconv.FormatFloat(data.(float64), 'f', 6, 64))
		server.SetMsg(data)
	case bool:
		// server.SetMsg(strconv.FormatBool(data.(bool)))
		server.SetMsg(data)
	case []interface{}:
		a := Trans(data)
		a0, ok := a.([]interface{})
		if ok {
			server.InsertResults(a0)
		}

	case map[string]interface{}:
		server.PutDatas(data.(map[string]interface{}))
	case map[interface{}]interface{}:
		m := Trans(data)
		m0, ok := m.(map[string]interface{})
		if ok {
			server.PutDatas(m0)
		}
	}
}

// Get ...
func Get(url string, handler func(*HTTPServer)) {
	http.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
		server := NewHTTPServer(w, r)
		if server != nil {
			if testData != nil && server.ObtainBooleanParam("test", false) {
				data, ok := testData[url]
				if ok && data != nil {
					toString(data, server)
				}
			} else {
				handler(server)
			}
			bytes, err := json.Marshal(server.message)
			if err == nil {
				server.ResponseWriter.Write(bytes)
			} else {
				server.ResponseWriter.Write([]byte("{\"ErrorCode\":-1,\"msg\":\"json数据转化失败!\"}"))
			}

			server.ResponseWriter.WriteHeader(http.StatusOK)
		}
	})
}

// Statics 静态页面
func Statics(config map[string]string, root string) {
	if root != "" && root[0] != '/' {
		root = "/" + root
	}
	for k, v := range config {
		_k := k
		if _k == "" {
			continue
		} else if _k[0] != '/' {
			_k = "/" + _k
		}
		http.Handle(root+_k, http.FileServer(http.Dir(v)))
	}
}

// Gets ...
func Gets(config map[string]func(*HTTPServer), root string) {
	if root != "" && root[0] != '/' {
		root = "/" + root
	}

	for k, v := range config {
		_k := k
		if _k == "" {
			continue
		} else if _k[0] != '/' {
			_k = "/" + _k
		}
		Get(root+_k, v)
	}
}

var testData map[string]interface{}

//SupportTest 支持测试数据的配置
func SupportTest(path string, path2 string) bool {
	// config, err := yaml.ReadFile(path)
	config, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)

		if path2 != "" {
			return SupportTest(path2, "")
		}
		return false
	}

	if testData == nil {
		testData = make(map[string]interface{})
	}

	err = yaml.Unmarshal(config, testData)
	if err != nil {
		fmt.Println(err)
		return false
	}

	println("config success !")

	return true
}

// SetMsg 设置字符串信息
func (s *HTTPServer) SetMsg(msg interface{}) {
	s.message.Msg = msg
}

// Error 设置错误信息
func (s *HTTPServer) Error(errorCode int, msg string) {
	s.message.ErrorCode = errorCode
	s.message.Msg = msg
}

// PutData 设置json数据 放到map中
func (s *HTTPServer) PutData(key string, value interface{}) {
	// s.message.Data = append(s.message.Data, )
	if value == nil {
		value = ""
	}
	if s.message.Data == nil {
		s.message.Data = make(map[string]interface{})
	}
	s.message.Data[key] = value
}

// PutDatas 设置json数据 放到map中
func (s *HTTPServer) PutDatas(data map[string]interface{}) {
	for k, v := range data {
		s.PutData(k, v)
	}
}

// InsertResult 设置数组数据 放到result中
func (s *HTTPServer) InsertResult(data interface{}) {
	if s.message.Result == nil {
		s.message.Result = make([]interface{}, 0)
	}
	s.message.Result = append(s.message.Result, data)
}

// InsertResults 设置数组数据 放到result中
func (s *HTTPServer) InsertResults(data []interface{}) {
	if s.message.Result == nil {
		s.message.Result = make([]interface{}, 0)
	}
	s.message.Result = append(s.message.Result, data...)
}

// ObtainStringParam 获取字符串参数
func (s *HTTPServer) ObtainStringParam(key string, defualtStr string) string {
	if s.Params != nil {
		v := s.Params[key]
		if len(v) > 0 {
			return v[0]
		}
		return defualtStr
	}
	return defualtStr
}

// ObtainIntParam 获取int参数
func (s *HTTPServer) ObtainIntParam(key string, defaultInt int) int {
	str := s.ObtainStringParam(key, strconv.Itoa(defaultInt))
	i, err := strconv.Atoi(str)
	if err != nil {
		return defaultInt
	}
	return i
}

// ObtainInt64Param 获取int64位的参数
func (s *HTTPServer) ObtainInt64Param(key string, defaultInt int64) int64 {
	str := s.ObtainStringParam(key, strconv.FormatInt(defaultInt, 10))
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return defaultInt
	}
	return i
}

// ObtainBooleanParam 获取bool参数
func (s *HTTPServer) ObtainBooleanParam(key string, defaultBool bool) bool {
	str := s.ObtainStringParam(key, strconv.FormatBool(defaultBool))
	i, err := strconv.ParseBool(str)
	if err != nil {
		return defaultBool
	}
	return i
}

// ObtainFloat64Param 获取float64位的参数
func (s *HTTPServer) ObtainFloat64Param(key string, defaultFloat float64) float64 {
	str := s.ObtainStringParam(key, "")
	i, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return defaultFloat
	}
	return i
}
