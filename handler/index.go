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

	// 如果有指定 categoryID，再去抓該分類底下的文件列表 (僅顯示已發布的文件)
	var docs []obj.Doc
	var uintCategoryID uint = 0
	var categoryExists bool = true

	if categoryIDStr != "" {
		categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
		if err != nil {
			// 非法的類別ID，顯示404頁面
			log.Println("Invalid category ID:", err)
			NotFoundHandler(w, r)
			return
		}

		uintCategoryID = uint(categoryID)

		// 檢查分類是否存在
		categoryExists = false
		// 增加日誌來顯示所有分類的 ID，以便調試
		categoryIDs := []uint{}
		for _, category := range categories {
			categoryIDs = append(categoryIDs, category.ID)
			if category.ID == uintCategoryID {
				categoryExists = true
				break
			}
		}

		if !categoryExists {
			// 類別不存在，顯示404頁面
			log.Printf("Category not found: %d. Available categories: %v", uintCategoryID, categoryIDs)
			NotFoundHandler(w, r)
			return
		}

		// 只取得已發布的文件
		docs, err = db.GetPublishedDocsByCategory(uintCategoryID)
		if err != nil {
			// 若失敗，記錄錯誤並改用空陣列
			log.Println("GetPublishedDocsByCategory error:", err)
			docs = []obj.Doc{}
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
