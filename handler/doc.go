package handler

import (
	"log"
	"net/http"
	ur "net/url"
	"strconv"
	"strings"
	"text/template"

	"github.com/HazelnutParadise/Go-Utils/conv"
	"github.com/yuin/goldmark"

	"support/db"
	"support/obj"
)

// DocHandler 負責顯示單篇文件內容
func DocHandler(w http.ResponseWriter, r *http.Request) {
	// 從 query string 拿到 doc ID
	docIDStr := r.URL.Query().Get("id")

	// 更完整地處理 docIDStr 可能包含的額外參數
	if idx := strings.IndexAny(docIDStr, "?&"); idx > 0 {
		docIDStr = docIDStr[:idx]
	}

	docID, err := strconv.ParseUint(docIDStr, 10, 32)
	if err != nil {
		log.Println("Invalid doc ID:", err)
		NotFoundHandler(w, r)
		return
	}

	// 從本地資料庫獲取文檔
	doc, err := db.GetDoc(uint(docID))
	if err != nil {
		log.Println("Error fetching doc:", err)
		NotFoundHandler(w, r)
		return
	}

	// 如果文件是草稿，返回 404
	if doc.IsDraft {
		NotFoundHandler(w, r)
		return
	}

	// 獲取分類列表
	categories, err := db.GetCategoryList()
	if err != nil {
		log.Println("Error fetching category list:", err)
		NotFoundHandler(w, r)
		return
	}

	// 當前分類
	currentCategory := r.URL.Query().Get("category")

	// 準備要給模板的資料
	data := obj.DocPageData{
		PageTitle:         "支援中心 - 榛果繽紛樂",
		DocFound:          true,
		DocTitle:          doc.Title,
		PublishDate:       doc.PublishDate,
		LastEditDate:      doc.LastEditDate.Format("2006-01-02"),
		CurrentCategory:   currentCategory,
		CurrentCategoryID: conv.ToString(doc.CategoryID),
		Categories:        categories,
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

	// 解析並執行模板
	tmpl, err := template.ParseFiles("templates/doc.html", "templates/header.html")
	if err != nil {
		log.Println("Template parse error:", err)
		http.Error(w, "Template parse error", http.StatusInternalServerError)
		return
	}

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
