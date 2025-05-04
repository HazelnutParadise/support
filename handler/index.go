package handler

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"support/db"
	"support/obj"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// 取得 query string
	categoryIDStr := r.URL.Query().Get("category_id")
	categoryName := r.URL.Query().Get("category_name")

	// 從本地資料庫取回分類列表
	categories, err := db.GetCategoryList()
	if err != nil {
		// 若失敗，記錄錯誤並改用空陣列
		log.Println("GetCategoryList error:", err)
		categories = []obj.Category{}
	}

	// 如果有指定 categoryID，再去抓該分類底下的文件列表
	var docs []obj.Doc
	var uintCategoryID uint = 0

	if categoryIDStr != "" {
		categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
		if err == nil {
			uintCategoryID = uint(categoryID)
			docs, err = db.GetDocsByCategory(uintCategoryID)
			if err != nil {
				// 若失敗，記錄錯誤並改用空陣列
				log.Println("GetDocsByCategory error:", err)
				docs = []obj.Doc{}
			}
		}
	}

	// 設定要傳遞給模板的資料
	data := obj.IndexData{
		Title:             "支援中心 - 榛果繽紛樂",
		Categories:        categories,
		Docs:              docs,
		CurrentCategory:   categoryName,
		CurrentCategoryID: uintCategoryID,
	}

	// 解析模板檔
	tmpl, err := template.ParseFiles("templates/index.html", "templates/header.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Template parsing error", http.StatusInternalServerError)
		return
	}

	// 輸出模板
	err = tmpl.ExecuteTemplate(w, "index", data)
	if err != nil {
		log.Println(err)
		http.Error(w, "Template executing error", http.StatusInternalServerError)
		return
	}
}
