package main

import (
	"crawler/download"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"
)

var userPath = "conf/user.json"
var loginPagr = "https://urs.earthdata.nasa.gov/home"
var FilePath = "china-2000-2015.txt"
var urlError = []string{}
var folder = "../data"

func main() {
	var now = time.Now()

	url_str := getUrls()
	urls := strings.Split(url_str, "\n")
	//urls = urls[0:1]
	fmt.Println("总文件数:",len(urls))
	fmt.Println("计算总大小中。。。")
	var size = download.GetTotalSize(urls)
	fmt.Println(fmt.Sprintf("总下载大小:%dM", size))

	var ch = make(chan string,3)
	var wg sync.WaitGroup
	wg.Add(len(urls))

	go func(urls []string) {
		for _,url := range urls{
			ch <- url
		}
	}(urls)

	if err := download.CheckOrMakeDir(folder);err !=nil {
		panic(err)
	}

	var consumer =  func() {
		for{
			url,ok := <- ch
			if !ok {
				break
			}

			fmt.Println("下载",url)
			err := download.Down(url, folder)
			if err != nil {
				fmt.Println(url)
				fmt.Println("下载出错：",err)
			}
			wg.Done()
		}
	}

	for i:=0;i<5;i++ {
		go consumer()
		time.Sleep(time.Second)
	}

	wg.Wait()
	fmt.Println("(｡･∀･)ﾉﾞ嗨，小老弟")
	fmt.Println("历经重重艰辛")
	fmt.Println("终于下载完成,总历时",time.Since(now))
}

type RowField struct {
	Row		int
	Field 	int
	Format 	string
}

func (r *RowField)toString()string  {
	return fmt.Sprintf(r.Format, r.Row,r.Field)
}

type RowFieldRange struct {
	Row			int
	FieldStart 	int
	FieldEnd 	int
	areaList 	func(RowFieldRange)[]string
}

func login(u User)  {
	//var req = http.Request{}
	//
	//var client = http.Client{}
}

func getAreaList()[]string  {
	var areaRange = []RowFieldRange{
		RowFieldRange{3,25,26,nil},
		RowFieldRange{4,23,27,nil},
		RowFieldRange{5,23,28,nil},
		RowFieldRange{6,25,29,nil},
		RowFieldRange{7,28,29,nil},
		RowFieldRange{8,28,29,nil},
	}

	var rangeToList = func(r RowFieldRange)[]string {
		var list = []string{}
		for i:=r.FieldStart; i<=r.FieldEnd; i++{
			rowFieldStr := fmt.Sprintf("h%02dv%02d", i,r.Row)
			list = append(list, rowFieldStr)
		}

		return list
	}

	var listFun = func()[]string {
		var arr []string
		for _,area := range areaRange {
			arr = append(arr,rangeToList(area)...)
		}

		return arr
	}

	return listFun()
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func getUserLoginInfoFromJson(path string) (User,error) {
	bytes_data,err := ioutil.ReadFile(path)
	if err != nil {
		return User{},err
	}

	var user User
	err = json.Unmarshal(bytes_data,&user)
	return user,err
}

func getUrls()string  {
	bytes_data,err := ioutil.ReadFile(FilePath)
	if err != nil {
		panic(err)
	}

	return string(bytes_data)
}