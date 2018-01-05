package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"strings"
	"time"
	"strconv"
	"os"
	"io/ioutil"
	"github.com/donnie4w/json4g"
	"encoding/json"
	"path/filepath"
	"os/exec"
	"github.com/robfig/cron"
)

type GuiConfig struct {
	Remarks string `json:"remarks"`
	Id string `json:"id"`
	Server string `json:"server"`
	Server_port int `json:"server_port"`
	Server_udp_port int `json:"server_udp_port"`
	Password string `json:"password"`
	Method string `json:"method"`
	Protocol string `json:"protocol"`
	Protocolparam string `json:"protocolparam"`
	Obfs string `json:"obfs"`
	Obfsparam string `json:"obfsparam"`
	Remarks_base64 string `json:"remarks_base_64"`
	Group string `json:"group"`
	Enable bool `json:"enable"`
	Udp_over_tcp bool `json:"udp_over_tcp"`
}

func main() {
	
	LoadConfig()
	
	// 0 5 0,6,12 * * ?
	/*
	2017/12/22 0:05:00
	2017/12/22 6:05:00
	2017/12/22 12:05:00
	 */

	c := cron.New()
	spec := "0 5 0,6,12 * * ?"
	c.AddFunc(spec, func() {
		fmt.Println("更新配置开始", time.Now())
		LoadConfig()
		fmt.Println("更新配置结束")
	})

	c.Start()

	select {}

}

func LoadConfig(){

	var(
		configFile string = "spider.json"
		url        string = "https://global.ishadowx.net/index_cn.html"
		guiConfig  string = "gui-config.json"
	)

	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Split(file)
	fmt.Println(path)

	configFile = filepath.Join(path,configFile)
	guiConfig = filepath.Join(path,guiConfig)

	config, err := os.Open(configFile)
	if err!=nil {
		panic("spider.config 配置文件没找到")
		return
	}

	all, err := ioutil.ReadAll(config)
	if err!=nil {
		panic("读取配置出错")
		return
	}

	configNode,err := json4g.LoadByString(string(all))
	if err!=nil {
		panic("json解析出错")
	}

	if nodeUrl:=configNode.GetNodeByName("url"); configNode.GetNodeByName("url") != nil {
		url = nodeUrl.ValueString
	}

	if guiConfigNode :=  configNode.GetNodeByName("guiConfig");  configNode.GetNodeByName("guiConfig") !=nil {
		guiConfig = guiConfigNode.ValueString
	}

	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}
	configs := []GuiConfig{}

	doc.Find(".hover-text").Each(func(i int, selection *goquery.Selection) {
		var guiConfig GuiConfig = GuiConfig{}
		guiConfig.Remarks = ""
		guiConfig.Id = strconv.FormatInt(time.Now().UnixNano(),10)
		guiConfig.Server_udp_port = 0
		guiConfig.Protocol="origin"
		guiConfig.Protocolparam=""
		guiConfig.Obfs="plain"
		guiConfig.Obfsparam=""
		guiConfig.Remarks_base64=""
		guiConfig.Group=""
		guiConfig.Enable=true
		guiConfig.Udp_over_tcp=false

		selection.Find("h4").Each(func(j int, h4 *goquery.Selection) {
			text := h4.Text()
			if text != "" {
				split := strings.Split(text, ":")
				for k,v := range split {
					if k==1 {
						v = strings.Replace(v,"\n","",-1)
						v = strings.Replace(v," ", "", -1)
						switch j {
						case 0:
							guiConfig.Server = v
						case 1:
							port,err := strconv.Atoi(v)
							if err != nil {
								fmt.Println("err:\t",err)
								continue
							}
							guiConfig.Server_port = port
						case 2:
							guiConfig.Password = v
						case 3:
							guiConfig.Method = v
						}
					}
				}
			}

			if j==4 {
				v,b := h4.Find("a").Attr("title")
				if b {
					guiConfig.Remarks = v
				} else {
					guiConfig.Remarks = h4.Text()
				}
			}
		})

		if guiConfig.Server_port > 0 {
			configs = append(configs,guiConfig)
		}

	})

	f,err:=os.OpenFile(guiConfig, os.O_RDWR,0666)
	if err!=nil {
		fmt.Println("err:\t",err)
		return
	}
	defer f.Close()

	buf, err:=ioutil.ReadAll(f)
	if err!=nil {
		fmt.Println("err:\t",err)
		return
	}
	jsonData := string(buf)

	jsonNode, err:=json4g.LoadByString(jsonData)
	if err!=nil{
		fmt.Println("err:\t",err)
		return
	}

	jsonNode.DelNode("configs")
	marshal, err := json.Marshal(configs)
	if err!=nil {
		fmt.Println("err:\t",err)
		return
	}

	jsonNode.GetNodeByPath("random").SetValue(true)
	jsonNode.GetNodeByPath("randomAlgorithm").SetValue(2)
	jsonNode.GetNodeByPath("autoBan").SetValue(true)

	jsonNode.AddNode(json4g.NowJsonNodeByString("configs",string(marshal)))
	jsonStr := jsonNode.ToString()

	f.Truncate(0)
	f.Seek(0,0)
	f.Write([]byte(jsonStr))
	
	cmd := exec.Command("cmd.exe", "/c", "taskkill /f /im ShadowsocksR-dotnet4.0.exe")
	err = cmd.Run()
	if err != nil {
		fmt.Println("ShadowsocksR-dotnet4.0.exe 关闭出错...............")
	}
	cmd = exec.Command("cmd.exe", "/c", "start ShadowsocksR-dotnet4.0.exe")
	err = cmd.Run()
	if err != nil {
		fmt.Println("ShadowsocksR-dotnet4.0.exe 启动出错...............")
	}
}
