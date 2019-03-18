package auth

import (
	"encoding/json"
	"github.com/antchfx/htmlquery"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var userPath = "conf/user.json"
var urlHome = "https://urs.earthdata.nasa.gov/home"
var urlLogin = "https://urs.earthdata.nasa.gov/login"
var urlData = `https://search.earthdata.nasa.gov/search/project?p=C194001241-LPDAAC_ECS&pg[0][qt]=2001-01-01T00%3A00%3A00.000Z%2
C2015-12-31T23%3A59%3A59.000Z&pg[0][oo]=true&pg[0][x]=1542726364!1048024675!1047378787!1047017493!1046286897!LPDAAC_ECS&q=MOD13Q1
%2520V006`

var header = map[string]string{
	"Host": "urs.earthdata.nasa.gov",
	"Origin": "https://urs.earthdata.nasa.gov",
	"Referer": "https://urs.earthdata.nasa.gov/home",
	"Upgrade-Insecure-Requests": "1",
	"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.121 Safari/537.36",
}

var form = map[string]string{
	"utf8": "✓",
	"authenticity_token": "",
	"username": "",
	"password": "",
	"client_id":"",
	"redirect_uri":"",
	"commit": "Log in",
}
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func GetUserLoginInfoFromJson(path string) (User,error) {
	bytes_data,err := ioutil.ReadFile(path)
	if err != nil {
		return User{},err
	}

	var user User
	err = json.Unmarshal(bytes_data,&user)
	return user,err
}

func Login() (AuthCookie, error) {
	cookie,token,err := getPreLoginInfo()
	if err != nil {
		return AuthCookie{}, err
	}
	form["authenticity_token"] = token
	form["redirect_uri"] = urlData
	var data = url.Values{}
	for k,v := range form {
		data.Set(k,v)
	}

	req, err := http.NewRequest("POST", urlLogin, strings.NewReader(data.Encode()))
	if err != nil {
		return AuthCookie{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", cookie)
	for k,v := range header {
		req.Header.Set(k,v)
	}

	var cli = http.Client{}
	resp,err :=cli.Do(req)
	var authCookie = AuthCookie{}
	if val,ok := resp.Header["Set-Cookie"]; ok {
		for _,v := range val {
			for _,val := range strings.Split(v,";"){
				if strings.Contains(val, "_urs-gui_session") {
					authCookie.Urs_gui_session = strings.Split(val, "=")[1]
					continue
				}

				if strings.Contains(val, "_ga") {
					authCookie.Gid = append(authCookie.Gid,strings.Split(val, "=")[1])
					continue
				}
			}
		}
	}

	authCookie.Urs_user_already_logged = "yes"

	return authCookie, err
}

func getPreLoginInfo() (string,string,error) {
	resp,err := http.Get(urlHome)
	if err != nil {
		return "","",err
	}
	defer resp.Body.Close()
	body,err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "","",err
	}

	doc, err := htmlquery.Parse(strings.NewReader(string(body)))
	if err != nil {
		return "","",err
	}

	meta := htmlquery.FindOne(doc, "//meta[6]")
	for _,v := range meta.Attr {
		if v.Key == "content"{
			return resp.Cookies()[0].String(),v.Val,nil
		}
	}

	return "","",nil
}

type AuthCookie struct {
	Urs_gui_session string
	Urs_user_already_logged string
	Gid []string
}

func (authCookie *AuthCookie) String()string  {
	var str string

	str += "_urs-gui_session=" + authCookie.Urs_gui_session + ";"
	str += "urs_user_already_logged=yes;"

	for _,gid := range authCookie.Gid {
		str += "_ga="+gid+";"
	}

	return str
}

func GetCookieData(cookie string)(string,error)  {
	uri := "https://e4ftl01.cr.usgs.gov//DP106/MOLT/MOD13Q1.006/2015.12.19/MOD13Q1.A2015353.h26v05.006.2018222150723.hdf"

	req,err := http.NewRequest("GET", uri, strings.NewReader(""))
	if err != nil {
		return "",err
	}

	var headers = map[string]string{
		"Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
		"Accept-Encoding": "gzip, deflate, br",
		"Accept-Language": "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7",
		"Connection": "keep-alive",
		"DNT": "1",
		"Host": "e4ftl01.cr.usgs.gov",
		"Upgrade-Insecure-Requests": "1",
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.121 Safari/537.36",
	}
	if len(cookie) > 0 {
		headers["Cookie"] = cookie
	}

	for k,v := range headers {
		req.Header.Set(k,v)
	}

	var cli = http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},}

	resp,err := cli.Do(req)
	if err != nil {
		return "",err
	}

	if resp.StatusCode == 302 {
		uri := resp.Header.Get("Location")
		req,err := http.NewRequest("GET", uri, strings.NewReader(""))
		if err != nil {
			return "",err
		}

		var headers = map[string]string{
			"Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			"Accept-Encoding": "gzip, deflate, br",
			"Accept-Language": "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7",
			"Connection": "keep-alive",
			"DNT": "1",
			"Host": "e4ftl01.cr.usgs.gov",
			"Upgrade-Insecure-Requests": "1",
			"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.121 Safari/537.36",
		}
		if len(cookie) > 0 {
			headers["Cookie"] = cookie
		}

		for k,v := range headers {
			req.Header.Set(k,v)
		}

		var cli = http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},}

		resp,err := cli.Do(req)
		if err != nil {
			return "",err
		}

		cookie := resp.Header.Get("Set-Cookie")
		if resp.StatusCode == 302 {
			uri := resp.Header.Get("Location")
			req,err := http.NewRequest("GET", uri, strings.NewReader(""))
			if err != nil {
				return "",err
			}

			if len(cookie) > 0 {
				headers["Cookie"] = cookie
			}

			for k,v := range headers {
				req.Header.Set(k,v)
			}

			var cli = http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},}

			resp,err := cli.Do(req)
			if err != nil {
				return "",err
			}

			cookie := resp.Header.Get("Set-Cookie")

			return strings.Split(cookie,";")[0],nil
		}
	}

	return "", nil
}
