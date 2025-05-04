package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"support/db"
)

// SearchDocInfo 代表搜尋結果中的一篇文件資訊
type SearchDocInfo struct {
	ID       uint   `json:"id"`
	Title    string `json:"title"`
	Category string `json:"category"`
}

// SearchHandler 處理搜尋請求
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 取得 query string
	query := r.URL.Query().Get("query")
	if query == "" {
		// 直接輸出提示訊息
		w.Write([]byte(`<p>請輸入搜索關鍵字。</p>`))
		return
	}

	// 2. 從本地資料庫搜尋文件
	docs, err := db.SearchDocs(query)
	if err != nil {
		// 若取得失敗，顯示錯誤
		msg := fmt.Sprintf("無法取得搜索結果：%v", err)
		w.Write([]byte("<p>" + msg + "</p>"))
		return
	}

	// 3. 準備搜尋結果
	if len(docs) > 0 {
		// 取得分類資訊，用於顯示分類名稱
		categoryMap := make(map[uint]string)
		categories, err := db.GetCategoryList()
		if err == nil {
			for _, category := range categories {
				categoryMap[category.ID] = category.Name
			}
		}

		// 逐筆輸出搜尋結果連結
		for _, doc := range docs {
			categoryName := categoryMap[doc.CategoryID]

			// 輸出搜尋結果連結
			link := fmt.Sprintf(
				`<a href="doc?id=%s&category=%s&title=%s">[%s] %s</a><br><br><hr>`,
				strconv.FormatUint(uint64(doc.ID), 10),
				url.QueryEscape(categoryName),
				url.QueryEscape(doc.Title),
				categoryName,
				doc.Title,
			)
			w.Write([]byte(link))
		}
	} else {
		// 無搜尋結果
		w.Write([]byte("<p>沒有找到符合的結果。</p>"))
	}
}
