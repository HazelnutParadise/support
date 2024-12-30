package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	ur "net/url"
	"strings"
	"text/template"

	"github.com/HazelnutParadise/Go-Utils/conv"
	"github.com/yuin/goldmark"

	"support/obj"
)

// docHandler 負責顯示單篇文件內容（原本的 doc.php）
func DocHandler(w http.ResponseWriter, r *http.Request) {
	// 從 query string 拿到 doc ID
	docID := r.URL.Query().Get("id")
	// 也許你有需要拿 category, title 之類
	currentCategory := r.URL.Query().Get("category")
	// currentTitle := r.URL.Query().Get("title")

	// 去呼叫遠端 API 拿文件資料
	url := "https://server2.hazelnut-paradise.com/supportDocs/doc?doc_id=" + docID
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error fetching doc:", err)
		renderErrorPage(w, "無法取得文件") // 可自行實作一個簡單的 error page
		return
	}
	defer resp.Body.Close()

	// 把 body 讀出來
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading doc response:", err)
		renderErrorPage(w, "無法讀取文件回應")
		return
	}

	// 解析 JSON
	var jsonData struct {
		Status  string  `json:"status"`
		Result  obj.Doc `json:"result"`
		Message string  `json:"message"`
	}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		log.Println("JSON parse error:", err)
		renderErrorPage(w, "文件格式錯誤")
		return
	}

	// 準備要給模板的資料
	data := obj.DocPageData{
		PageTitle:         "支援中心 - 榛果繽紛樂",
		DocFound:          false,
		DocTitle:          "",
		PublishDate:       "",
		LastEditDate:      "",
		HTMLContent:       "",
		CurrentCategory:   currentCategory,
		CurrentCategoryID: "", // 稍後在成功時再填
	}

	if resp.StatusCode != 200 {
		// API 回傳非 200 時
		msg := fmt.Sprintf("發生錯誤 (HTTP %d)。", resp.StatusCode)
		if jsonData.Message != "" {
			msg += " 錯誤訊息：" + jsonData.Message
		}
		// 直接在模板上顯示錯誤
		data.HTMLContent = msg
	} else {
		// status == success
		if jsonData.Status == "success" {
			doc := jsonData.Result
			data.DocFound = true
			data.DocTitle = doc.Title
			data.CurrentCategoryID = conv.ToString(doc.CategoryID)
			// 只取 yyyy-mm-dd
			if len(doc.PublishDate) >= 10 {
				data.PublishDate = doc.PublishDate[:10]
			}
			if len(doc.LastEditDate) >= 10 {
				data.LastEditDate = doc.LastEditDate[:10]
			}

			// URL 解碼
			md, err := ur.QueryUnescape(doc.Content)
			if err != nil {
				log.Println("URL decode error:", err)
				data.HTMLContent = "<p>內容解析失敗</p>"
			} else {
				var htmlBuilder strings.Builder
				if err := goldmark.Convert([]byte(md), &htmlBuilder); err != nil {
					log.Println("Markdown parse error:", err)
					data.HTMLContent = "<p>內容解析失敗</p>"
				} else {
					data.HTMLContent = htmlBuilder.String()
				}
			}

			data.PageTitle = doc.Title + " | " + data.PageTitle
		} else {
			// API 回傳 success != 'success'
			data.HTMLContent = "<p>文章未找到。 (status != success)</p>"
		}
	}

	// 解析並執行模板
	tmpl, err := template.ParseFiles("template/doc.html", "template/header.html")
	if err != nil {
		log.Println("Template parse error:", err)
		http.Error(w, "Template parse error", http.StatusInternalServerError)
		return
	}

	// 與之前一樣，如果你的模板有用 `{{ define "doc" }}`，要用 ExecuteTemplate
	// 若 template 直接就是 doc.tmpl 裡全部，那就 tmpl.Execute(w, data) 即可
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("Template execute error:", err)
		http.Error(w, "Template execute error", http.StatusInternalServerError)
		return
	}
}

// renderErrorPage 只是簡單顯示錯誤頁
func renderErrorPage(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte("<h1>發生錯誤</h1><p>" + message + "</p>"))
}
