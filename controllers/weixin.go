package controllers

import (
        "PrometheusAlert/models"
        "bytes"
        "crypto/tls"
        "encoding/json"
        "github.com/astaxie/beego"
        "github.com/astaxie/beego/logs"
        "io/ioutil"
        "net/http"
        "net/url"
        "strings"
)

type TextContent struct {
        Content             string   `json:"content"`
        MentionedList       []string `json:"mentioned_list,omitempty"`
        MentionedMobileList []string `json:"mentioned_mobile_list,omitempty"`
}

type WXMessage struct {
        Msgtype string     `json:"msgtype"`
        Text    TextContent `json:"text"`
}

var client *http.Client

func init() {
        var tr *http.Transport
        maxIdleConns, _ := beego.AppConfig.Int("MaxIdleConns")
        tr = &http.Transport{MaxIdleConns: maxIdleConns}
        if proxyUrl := beego.AppConfig.String("proxy"); proxyUrl != "" {
                proxy := func(_ *http.Request) (*url.URL, error) {
                        return url.Parse(proxyUrl)
                }
                tr = &http.Transport{
                        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
                        Proxy:           proxy,
                }
        } else {
                tr = &http.Transport{
                        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
                }
        }
        client = &http.Client{Transport: tr}
}

func PostToWeiXin(text, WXurl, atuserid, logsign string) string {
        open := beego.AppConfig.String("open-weixin")
        if open != "1" {
                logs.Info(logsign, "[weixin]", "企业微信接口未配置未开启状态,请先配置open-weixin为1")
                return "企业微信接口未配置未开启状态,请先配置open-weixin为1"
        }

        SendContent := text
        var mentionedList []string
        if atuserid != "" {
                userid := strings.Split(atuserid, ",")
                mentionedList = userid
        }

        u := WXMessage{
                Msgtype: "text",
                Text: TextContent{
                        Content:       SendContent,
                        MentionedList: mentionedList,
                },
        }

        b := new(bytes.Buffer)
        json.NewEncoder(b).Encode(u)
        logs.Info(logsign, "[weixin]", b)

        res, err := client.Post(WXurl, "application/json", b)
        if err != nil {
                logs.Error(logsign, "[weixin]", err.Error())
        }
        defer res.Body.Close()

        result, err := ioutil.ReadAll(res.Body)
        if err != nil {
                logs.Error(logsign, "[weixin]", err.Error())
        }

        models.AlertToCounter.WithLabelValues("weixin").Add(1)
        ChartsJson.Weixin += 1
        logs.Info(logsign, "[weixin]", string(result))
        return string(result)
}
