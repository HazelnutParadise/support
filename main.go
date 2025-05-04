package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"support/db"
	"support/handler"
)

func removePHP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, ".php") {
			originalQuery := r.URL.RawQuery
			newPath := strings.Replace(r.URL.Path, ".php", "", -1)
			if originalQuery != "" {
				newPath += "?" + originalQuery
			}
			http.Redirect(w, r, newPath, http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// 初始化資料庫連接
	_, err := db.DB()
	if err != nil {
		log.Fatalf("資料庫初始化失敗: %v", err)
	}
	fmt.Println("資料庫連接成功")

	// 設定路由
	mux := http.NewServeMux()

	// 前台路由
	mux.HandleFunc("/", handler.IndexHandler)
	mux.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		originalQuery := r.URL.RawQuery
		newPath := "/"
		if originalQuery != "" {
			newPath += "?" + originalQuery
		}
		http.Redirect(w, r, newPath, http.StatusMovedPermanently)
	})
	mux.HandleFunc("/doc", handler.DocHandler)
	mux.HandleFunc("/search", handler.SearchHandler)

	// 後台管理路由
	mux.HandleFunc("/admin", handler.AdminDashboardHandler)
	mux.HandleFunc("/admin/categories", handler.AdminCategoriesHandler)
	mux.HandleFunc("/admin/categories/add", handler.AdminCategoryAddHandler)
	mux.HandleFunc("/admin/categories/edit", handler.AdminCategoryEditHandler)
	mux.HandleFunc("/admin/categories/delete", handler.AdminCategoryDeleteHandler)
	mux.HandleFunc("/admin/docs", handler.AdminDocsHandler)
	mux.HandleFunc("/admin/docs/add", handler.AdminDocAddHandler)
	mux.HandleFunc("/admin/docs/edit", handler.AdminDocEditHandler)
	mux.HandleFunc("/admin/docs/update", handler.AdminDocUpdateHandler)
	mux.HandleFunc("/admin/docs/delete", handler.AdminDocDeleteHandler)

	// 啟動服務
	fmt.Println("伺服器執行中... http://localhost:3000/")
	log.Fatal(http.ListenAndServe(":3000", removePHP(mux)))
}
