package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/HazelnutParadise/Go-Utils/conv"
)

// SearchResponse 用來對應後端 API 回傳的 JSON 結構
type SearchResponse struct {
	Status string          `json:"status"`
	Result []SearchDocInfo `json:"result"`
}

// SearchDocInfo 代表搜尋結果中的一篇文件資訊
type SearchDocInfo struct {
	ID       uint   `json:"id"`
	Title    string `json:"title"`
	Category string `json:"category"`
}

// searchHandler 取代原本的 search.php
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 取得 query string
	query := r.URL.Query().Get("query")
	if query == "" {
		// 直接輸出一段 HTML
		w.Write([]byte(`<p>請輸入搜索關鍵字。</p>`))
		return
	}

	// 2. 呼叫後端 API
	apiURL := "https://server2.hazelnut-paradise.com/supportDocs/searchDocs?keyword=" + url.QueryEscape(query)
	resp, err := http.Get(apiURL)
	if err != nil {
		// 若取得失敗，顯示錯誤
		msg := fmt.Sprintf("無法取得搜索結果：%v", err)
		w.Write([]byte("<p>" + msg + "</p>"))
		return
	}
	defer resp.Body.Close()

	// 讀取 body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		msg := fmt.Sprintf("無法讀取搜索結果：%v", err)
		w.Write([]byte("<p>" + msg + "</p>"))
		return
	}

	// 3. 解析 JSON
	var data SearchResponse
	if err := json.Unmarshal(body, &data); err != nil {
		msg := fmt.Sprintf("JSON 解析失敗：%v", err)
		w.Write([]byte("<p>" + msg + "</p>"))
		return
	}

	// 4. 根據 status 輸出結果
	if data.Status == "success" && len(data.Result) > 0 {
		// 與 PHP 版本類似，逐筆輸出超連結
		for _, result := range data.Result {
			// 為了與原本 PHP 相符，這裡還是寫成 doc.php?id=...
			// 若你的 Go 專案 doc 改成 /doc?id=...，請自行調整
			link := fmt.Sprintf(
				`<a href="doc?id=%s&category=%s&title=%s">[%s] %s</a><br><br><hr>`,
				url.QueryEscape(conv.ToString(result.ID)),
				url.QueryEscape(result.Category),
				url.QueryEscape(result.Title),
				result.Category,
				result.Title,
			)
			w.Write([]byte(link))
		}
	} else {
		// 無法取得搜索結果（或無資料）
		w.Write([]byte("<p>無法取得搜索結果。</p>"))
	}
}
