package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"support/obj"

	"github.com/HazelnutParadise/Go-Utils/conv"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// 取得 query string
	categoryID := r.URL.Query().Get("category_id")
	categoryName := r.URL.Query().Get("category_name")

	// 從後端 API 取回分類列表
	categories, err := fetchCategories()
	if err != nil {
		// 若失敗，記錄錯誤並改用空陣列
		log.Println("fetchCategories error:", err)
		categories = []obj.Category{}
	}

	// 如果有指定 categoryID，再去抓該分類底下的文件列表
	var docs []obj.Doc
	if categoryID != "" {
		docs, err = fetchDocs(categoryID)
		if err != nil {
			// 若失敗，記錄錯誤並改用空陣列
			log.Println("fetchDocs error:", err)
			docs = []obj.Doc{}
		}
	}

	uintCategoryID := uint(0)
	if categoryID != "" {
		uintCategoryID = uint(conv.ParseInt(categoryID))
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

// 示範抓分類列表
func fetchCategories() ([]obj.Category, error) {
	url := "https://server2.hazelnut-paradise.com/supportDocs/categoriesList"
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response struct {
		Status string         `json:"status"`
		Result []obj.Category `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	if response.Status != "success" {
		return nil, fmt.Errorf("取得分類列表失敗")
	}

	return response.Result, nil
}

// 示範抓文件列表
func fetchDocs(categoryID string) ([]obj.Doc, error) {
	url := "https://server2.hazelnut-paradise.com/supportDocs/docsList?category_id=" + categoryID
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response struct {
		Status string    `json:"status"`
		Result []obj.Doc `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	if response.Status != "success" {
		return nil, fmt.Errorf("取得文件列表失敗")
	}

	return response.Result, nil
}
