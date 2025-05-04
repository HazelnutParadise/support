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

// 包裝 IndexHandler 以處理精確路徑匹配
func ExactPathIndexHandler(w http.ResponseWriter, r *http.Request) {
	// 只有當路徑確實是 "/" 時才調用 IndexHandler
	if r.URL.Path != "/" {
		handler.NotFoundHandler(w, r)
		return
	}
	handler.IndexHandler(w, r)
}

// 自定義 NotFound 處理器
type CustomNotFoundHandler struct {
	Mux *http.ServeMux
}

func (h *CustomNotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 使用內部 mux 來嘗試處理請求
	handlerFunc, pattern := h.Mux.Handler(r)

	if pattern != "" {
		// 有匹配的路由，使用原有處理器
		handlerFunc.ServeHTTP(w, r)
		return
	}

	// 沒有匹配的路由，調用自定義 404 處理函數
	handler.NotFoundHandler(w, r)
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
	mux.HandleFunc("/", ExactPathIndexHandler)
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
	mux.HandleFunc("/category-docs", handler.CategoryDocsHandler) // 新增分類文章列表路由

	// 後台登入/登出路由 (不需要驗證)
	mux.HandleFunc("/admin/login", handler.AdminLoginHandler)
	mux.HandleFunc("/admin/logout", handler.AdminLogoutHandler)

	// 後台管理路由 (需要身份驗證)
	mux.HandleFunc("/admin", handler.AuthMiddleware(handler.AdminDashboardHandler))
	mux.HandleFunc("/admin/dashboard", handler.AuthMiddleware(handler.AdminDashboardHandler))
	mux.HandleFunc("/admin/categories", handler.AuthMiddleware(handler.AdminCategoriesHandler))
	mux.HandleFunc("/admin/categories/add", handler.AuthMiddleware(handler.AdminCategoryAddHandler))
	mux.HandleFunc("/admin/categories/edit", handler.AuthMiddleware(handler.AdminCategoryEditHandler))
	mux.HandleFunc("/admin/categories/delete", handler.AuthMiddleware(handler.AdminCategoryDeleteHandler))
	mux.HandleFunc("/admin/docs", handler.AuthMiddleware(handler.AdminDocsHandler))
	mux.HandleFunc("/admin/docs/add", handler.AuthMiddleware(handler.AdminDocAddHandler))
	mux.HandleFunc("/admin/docs/edit", handler.AuthMiddleware(handler.AdminDocEditHandler))
	mux.HandleFunc("/admin/docs/update", handler.AuthMiddleware(handler.AdminDocUpdateHandler))
	mux.HandleFunc("/admin/docs/delete", handler.AuthMiddleware(handler.AdminDocDeleteHandler))
	// 添加密碼修改路由
	mux.HandleFunc("/admin/change-password", handler.AuthMiddleware(handler.AdminChangePasswordHandler))

	// 創建一個自定義的 NotFound 處理器
	notFoundWrapper := &CustomNotFoundHandler{Mux: mux}

	// 啟動服務
	fmt.Println("伺服器執行中... http://localhost:3000/")
	log.Fatal(http.ListenAndServe(":3000", removePHP(notFoundWrapper)))
}
