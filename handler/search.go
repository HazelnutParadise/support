package handler

import (
	"encoding/json"
	"fmt"
	"log"
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

// CategoryDocsHandler 處理獲取特定分類下文章列表的請求
func CategoryDocsHandler(w http.ResponseWriter, r *http.Request) {
	// 設置 HTTP 頭，允許跨域請求
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 從 query string 獲取分類 ID
	categoryIDStr := r.URL.Query().Get("category_id")
	if categoryIDStr == "" {
		// 如果沒有提供分類 ID，返回錯誤
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":"error","message":"缺少分類ID"}`))
		return
	}

	// 轉換分類 ID 為 uint
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
	if err != nil {
		// 分類 ID 格式錯誤，返回錯誤
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":"error","message":"分類ID格式錯誤"}`))
		return
	}

	// 獲取該分類下的所有已發布文檔
	docs, err := db.GetPublishedDocsByCategory(uint(categoryID))
	if err != nil {
		// 獲取文檔失敗，返回錯誤
		log.Println("Error fetching docs by category:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status":"error","message":"獲取文檔失敗"}`))
		return
	}

	// 構造簡化的文檔數據以輸出
	type SimpleDoc struct {
		ID    uint   `json:"id"`
		Title string `json:"title"`
	}
	simpleDocs := make([]SimpleDoc, len(docs))
	for i, doc := range docs {
		simpleDocs[i] = SimpleDoc{
			ID:    doc.ID,
			Title: doc.Title,
		}
	}

	// 將文檔列表轉換為 JSON 並返回
	jsonData := map[string]interface{}{
		"status": "success",
		"result": simpleDocs,
	}

	// 序列化 JSON
	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		// JSON 序列化失敗，返回錯誤
		log.Println("JSON marshal error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status":"error","message":"JSON序列化失敗"}`))
		return
	}

	// 返回 JSON 數據
	w.Write(jsonBytes)
}
