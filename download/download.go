package download

import (
	"errors"
	"fmt"
	"github.com/cavaliercoder/grab"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var DownloadUrls []string
var lock sync.RWMutex
var header = map[string]string{
	"Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
	"Accept-Encoding": "gzip, deflate, br",
	"Accept-Language": "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7",
	"Connection": "keep-alive",
	"Cookie": "DATA=XI3@4@pjbS1TePcAGPeHdgAAADY",
	"DNT": "1",
	"Host": "e4ftl01.cr.usgs.gov",
	"Upgrade-Insecure-Requests": "1",
	"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.121 Safari/537.36",
}

func GetTotalSize(urls []string)int64{
	var wg = sync.WaitGroup{}
	wg.Add(len(urls))
	var TotalSize int64

	for _,url := range urls {
		go func(url string) {
			length,err := getFileSize(url)
			if err != nil {
				fmt.Println(err)
			}else{
				lock.Lock()
				TotalSize+=length
				lock.Unlock()
			}
			wg.Done()
		}(url)

		time.Sleep(30*time.Millisecond)
	}

	wg.Wait()

	//返回单位为M
	return TotalSize/(1024*1024)
}

func getFileSize(url string)(int64,error)  {
	//*(http.Header) = http.Header{}
	//for k,v := range header{
	//	http.Header.Set(k,v)
	//}
	resp,err := http.Head(url)
	if err != nil {
		return 0, errors.New("head请求出错："+ err.Error())
	}

	return resp.ContentLength, nil
}

func Down(url,folder string) error {
	fileName := getFileName(url)
	filePath := folder + "/" + fileName

	if exists(filePath) {
		size,err := getFileSize(url)
		if err != nil {
			return errors.New("head获取文件大小出错:"+err.Error())
		}

		complete,err := IsFileComplete(size,filePath)
		if err != nil {
			return errors.New("检查文件完整性出错："+err.Error())
		}

		if complete {
			fmt.Println(fmt.Sprintf("文件%s已完全下载，本次跳过",filePath))
		}else{
			fmt.Println(fmt.Sprintf("文件%s检验不完整,重新下载", filePath))
			if err = os.Remove(filePath);err != nil {
				fmt.Println(fmt.Sprintf("删除%s文件失败:%s",filePath,err.Error()))
			}
		}
	}

	return d(folder,url)
}


func IsFileComplete(size int64,path string) (bool,error) {
	fi, err := os.Stat(path);
	if err == nil{
		return fi.Size()== size, nil
	}

	return false, err
}

func exists(path string)bool  {
	if _, err := os.Stat(path); err != nil{
		if os.IsNotExist(err){
			return false
		}
	}

	return	true
}

func getFileName(url string)string  {
	url_parse := strings.Split(url,"/")
	return url_parse[len(url_parse)-1]
}

func Mkdir(path string)error  {
	return os.Mkdir(path,0777)
}

func CheckOrMakeDir(path string)error  {
	if exist := exists(path);!exist {
		return Mkdir(path)
	}

	return nil
}

func d(folder,url string) error {
	client := grab.NewClient()
	//
	//cookiejar,err := cookiejar.New(nil)
	//if err != nil {
	//	return errors.New("set cookie出错："+err.Error())
	//}
	//client.HTTPClient.Jar = cookiejar

	req, err := grab.NewRequest(folder, url)
	if err != nil {
		return errors.New("创建request出错："+err.Error())
	}

	for k,v := range header{
		req.HTTPRequest.Header.Set(k,v)
	}

	resp := client.Do(req)

	// start UI loop
	t := time.NewTicker(10 * time.Minute)
	defer t.Stop()

	Loop:
		for {
			select {
			case <-t.C:
				fmt.Printf("  transferred %v / %v bytes (%.2f%%)\n",
					resp.BytesComplete(),
					resp.Size,
					100*resp.Progress())

			case <-resp.Done:
				// download is complete
				break Loop
			}
		}

		// check for errors
		if err := resp.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
			return err
		}

		if resp.HTTPResponse.StatusCode != 200 {
			return errors.New("认证失败")
		}

	return nil
}