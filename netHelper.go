package main

import (
	"context"
	"crypto/tls"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/tidwall/gjson"
)

const csstitle = "title"

func testRun(injson string) (string, string) {
	if gjson.Get(injson, "ishight").Int() == 0 {
		body := simpleRun(gjson.Get(injson, "address").String(), gjson.Get(injson, "cookie").String())
		if gjson.Get(injson, "csschoose").String() != "" {
			return cssPath(strings.NewReader(body), gjson.Get(injson, "csschoose").String()), cssPath(strings.NewReader(body), csstitle)
		} else {
			return xPath(strings.NewReader(body), gjson.Get(injson, "xpathchoose").String()), cssPath(strings.NewReader(body), csstitle)
		}
	} else {
		body := chromedpRun(gjson.Get(injson, "address").String(), gjson.Get(injson, "cookie").String())
		if gjson.Get(injson, "csschoose").String() != "" {
			return cssPath(strings.NewReader(body), gjson.Get(injson, "csschoose").String()), cssPath(strings.NewReader(body), csstitle)
		} else {
			return xPath(strings.NewReader(body), gjson.Get(injson, "xpathchoose").String()), cssPath(strings.NewReader(body), csstitle)
		}
	}
}
func chromedpRun(url, cookie string) string {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, time.Duration(_config["runtime"].(int64))*time.Second)
	defer cancel()
	var outerBefore string
	err := chromedp.Run(ctx,
		network.Enable(),
		network.SetExtraHTTPHeaders(network.Headers(map[string]interface{}{
			"Cookie": cookie,
		})),
		chromedp.Navigate(url),
		chromedp.OuterHTML("html", &outerBefore),
	)
	if err != nil {
		return (err.Error())
	}
	return outerBefore
}
func simpleRun(url, cookie string) string {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(_config["runtime"].(int64)) * time.Second,
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Cookie", cookie)
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return string(body)
}
func cssPath(body io.Reader, choose string) string {
	doc, _ := goquery.NewDocumentFromReader(body)
	return standardizeSpaces(doc.Find(choose).Text())
}
func xPath(body io.Reader, choose string) string {
	if choose == "" {
		data, _ := ioutil.ReadAll(body)
		return string(data)
	}
	doc, _ := htmlquery.Parse(body)
	list := htmlquery.Find(doc, choose)
	res := ""
	for _, n := range list {
		res += htmlquery.InnerText(n)
	}
	return standardizeSpaces(res)
}

func Run(ishight, address, cookie, csschoose, xpathchoose string) (string, string) {
	if ishight == "0" {
		body := simpleRun(address, cookie)
		if csschoose != "" {
			return cssPath(strings.NewReader(body), csschoose), cssPath(strings.NewReader(body), csstitle)
		} else {
			return xPath(strings.NewReader(body), xpathchoose), cssPath(strings.NewReader(body), csstitle)
		}
	} else {
		body := chromedpRun(address, cookie)
		if csschoose != "" {
			return cssPath(strings.NewReader(body), csschoose), cssPath(strings.NewReader(body), csstitle)
		} else {
			return xPath(strings.NewReader(body), xpathchoose), cssPath(strings.NewReader(body), csstitle)
		}
	}
}
func getRun(url string) string {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(_config["runtime"].(int64)) * time.Second,
	}
	req, _ := http.NewRequest("GET", url, nil)
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return string(body)
}
func postRun(url, post, contentType string) string {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(_config["runtime"].(int64)) * time.Second,
	}
	resp, _ := client.Post(url, contentType, strings.NewReader(post))
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return string(body)
}
func SendToMail(subject, body string) error {
	hp := strings.Split(_config["emailhost"].(string), ":")
	auth := smtp.PlainAuth("", _config["email"].(string), _config["emailpass"].(string), hp[0])
	content_type := "Content-Type: text/plain" + "; charset=UTF-8"

	msg := []byte("To: " + _config["emailto"].(string) + "\r\nFrom: " + _config["email"].(string) + ">\r\nSubject: " + subject + "\r\n" + content_type + "\r\n\r\n" + body)
	send_to := strings.Split(_config["emailto"].(string), ";")
	err := smtp.SendMail(_config["emailhost"].(string), auth, _config["email"].(string), send_to, msg)
	if err != nil {
		log.Println(err.Error())
	}
	return err
}
