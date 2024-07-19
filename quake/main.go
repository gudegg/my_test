package main

import (
	"bufio"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"
	"github.com/tidwall/gjson"
	"log"
	"os"
	"strings"
	"time"
)

var validClient = resty.New()

func init() {
	validClient.SetRetryCount(2)
	validClient.SetRetryWaitTime(time.Second)
	validClient.SetTimeout(10 * time.Second)
}

var UserAgent = "Mozilla/5.0 (iPhone;CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko)Version/9.0 Mobile/13B143 Safari/601.1 (compatible; Baiduspider-render/2.0;+http://www.baidu.com/search/spider.html) "

func main() {
	fileList := readSyncFile()
	for _, s := range fileList {
		if len(s) == 0 {
			continue
		}
		split := strings.Split(s, "=")
		if len(split) != 2 {
			continue
		}
		valid := Valid(Source{
			ShareId:  split[1],
			SharePwd: "",
		})
		if !valid {
			fmt.Println("失效链接:" + s)
		}
		time.Sleep(time.Second)
	}
}

func readSyncFile() []string {
	result := make([]string, 0)
	f, err := os.OpenFile("quake_top.txt", os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Fatalf("open file error: %v", err)
		return result
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		str := strings.TrimSpace(sc.Text()) // GET the line string
		if len(str) > 0 {
			result = append(result, str)
		}
	}
	if err := sc.Err(); err != nil {
		log.Fatalf("scan file error: %v", err)
		return result
	}
	return result
}

func Valid(source Source) bool {
	body := fmt.Sprintf(`{"pwd_id":"%v","passcode":"%v"}`, source.ShareId, source.SharePwd)
	resp, err := validClient.R().SetHeader("user-agent", UserAgent).
		SetHeader("Content-Type", "application/json").SetBody(body).
		Post(fmt.Sprintf("https://drive-pc.quark.cn/1/clouddrive/share/sharepage/token?pr=ucpro&fr=pc&uc_param_str=&__dt=746&__t=%v", time.Now().UnixMilli()))
	if err != nil {
		fmt.Sprintf("AsyncValid err %v", err)
		return true
	}
	if resp.StatusCode() != 200 {
		statusList := []int64{403, 404}
		code := gjson.Get(resp.String(), "code").String()
		if lo.Contains(statusList, gjson.Get(resp.String(), "status").Int()) && strings.HasPrefix(code, "410") {
			fmt.Sprintf("资源失效,shareid=%v,resp=%v", source.ShareId, resp.String())
			return false
		}
		fmt.Sprintf("获取token失败,%v", resp.String())
	}
	fmt.Sprintf("资源有效,shareid=%v,fileid=%v,resp=%v", source.ShareId, resp.String())
	return true
}

type Source struct {
	ShareId  string `json:"shareId"`
	SharePwd string `json:"sharePwd"`
}
