package handler

import (
	"log"
	"net/http"
	"text/template"
)

// NotFoundHandler 處理 404 錯誤頁面
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)

	// 解析模板
	tmpl, err := template.ParseFiles("templates/404.html", "templates/header.html")
	if err != nil {
		log.Println("Template parse error:", err)
		http.Error(w, "頁面未找到", http.StatusNotFound)
		return
	}

	// 傳遞空的資料結構
	err = tmpl.Execute(w, nil)
	if err != nil {
		log.Println("Template execute error:", err)
		http.Error(w, "頁面未找到", http.StatusNotFound)
		return
	}
}
