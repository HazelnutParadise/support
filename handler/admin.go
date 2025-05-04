package handler

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"support/db"
	"support/obj"
	"text/template"
	"time"
)

// AdminDashboardHandler 處理管理面板首頁
func AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	// 從資料庫獲取統計數據
	categories, err := db.GetCategoryList()
	if err != nil {
		log.Println("Error fetching categories:", err)
	}

	// 獲取所有文件，僅用於計數
	docs, err := db.GetAllDocs()
	if err != nil {
		log.Println("Error fetching documents:", err)
	}

	// 準備模板資料
	data := map[string]interface{}{
		"Active":        "dashboard",
		"CategoryCount": len(categories),
		"DocCount":      len(docs),
	}

	// 解析模板
	tmpl, err := template.ParseFiles(
		"templates/admin/layout.html",
		"templates/admin/dashboard.html",
	)
	if err != nil {
		log.Println("Template parse error:", err)
		http.Error(w, "Template parse error", http.StatusInternalServerError)
		return
	}

	// 渲染模板
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("Template execute error:", err)
		http.Error(w, "Template execute error", http.StatusInternalServerError)
		return
	}
}

// AdminCategoriesHandler 處理分類管理頁面
func AdminCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	// 從資料庫獲取分類列表
	categories, err := db.GetCategoryList()
	if err != nil {
		log.Println("Error fetching categories:", err)
	}

	// 檢查是否有訊息要顯示（例如操作成功訊息）
	message := r.URL.Query().Get("message")
	messageType := r.URL.Query().Get("type")
	if messageType == "" && message != "" {
		messageType = "success" // 預設訊息類型
	}

	// 準備模板資料
	data := map[string]interface{}{
		"Active":      "categories",
		"Categories":  categories,
		"Message":     message,
		"MessageType": messageType,
	}

	// 解析模板
	tmpl, err := template.ParseFiles(
		"templates/admin/layout.html",
		"templates/admin/categories.html",
	)
	if err != nil {
		log.Println("Template parse error:", err)
		http.Error(w, "Template parse error", http.StatusInternalServerError)
		return
	}

	// 渲染模板
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("Template execute error:", err)
		http.Error(w, "Template execute error", http.StatusInternalServerError)
		return
	}
}

// AdminCategoryAddHandler 處理新增分類
func AdminCategoryAddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
		return
	}

	// 解析表單
	err := r.ParseForm()
	if err != nil {
		log.Println("Form parse error:", err)
		redirectWithMessage(w, r, "/admin/categories", "表單解析錯誤", "danger")
		return
	}

	// 獲取分類名稱
	name := r.FormValue("name")
	if name == "" {
		redirectWithMessage(w, r, "/admin/categories", "分類名稱不能為空", "danger")
		return
	}

	// 創建新分類
	category := obj.Category{
		Name:       name,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	// 保存到資料庫
	err = db.AddCategory(&category)
	if err != nil {
		log.Println("Error adding category:", err)
		redirectWithMessage(w, r, "/admin/categories", "新增分類失敗: "+err.Error(), "danger")
		return
	}

	// 重定向回分類列表，並帶上成功訊息
	redirectWithMessage(w, r, "/admin/categories", "成功新增分類: "+name, "success")
}

// AdminCategoryEditHandler 處理編輯分類
func AdminCategoryEditHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
		return
	}

	// 解析表單
	err := r.ParseForm()
	if err != nil {
		log.Println("Form parse error:", err)
		redirectWithMessage(w, r, "/admin/categories", "表單解析錯誤", "danger")
		return
	}

	// 獲取分類 ID 和名稱
	idStr := r.FormValue("id")
	name := r.FormValue("name")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		log.Println("Invalid ID:", err)
		redirectWithMessage(w, r, "/admin/categories", "無效的ID", "danger")
		return
	}

	if name == "" {
		redirectWithMessage(w, r, "/admin/categories", "分類名稱不能為空", "danger")
		return
	}

	// 更新分類
	err = db.UpdateCategory(uint(id), name)
	if err != nil {
		log.Println("Error updating category:", err)
		redirectWithMessage(w, r, "/admin/categories", "更新分類失敗: "+err.Error(), "danger")
		return
	}

	// 重定向回分類列表，並帶上成功訊息
	redirectWithMessage(w, r, "/admin/categories", "成功更新分類", "success")
}

// AdminCategoryDeleteHandler 處理刪除分類
func AdminCategoryDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
		return
	}

	// 解析表單
	err := r.ParseForm()
	if err != nil {
		log.Println("Form parse error:", err)
		redirectWithMessage(w, r, "/admin/categories", "表單解析錯誤", "danger")
		return
	}

	// 獲取分類 ID
	idStr := r.FormValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		log.Println("Invalid ID:", err)
		redirectWithMessage(w, r, "/admin/categories", "無效的ID", "danger")
		return
	}

	// 刪除分類及其下的所有文檔
	err = db.DeleteCategory(uint(id))
	if err != nil {
		log.Println("Error deleting category:", err)
		redirectWithMessage(w, r, "/admin/categories", "刪除分類失敗: "+err.Error(), "danger")
		return
	}

	// 重定向回分類列表，並帶上成功訊息
	redirectWithMessage(w, r, "/admin/categories", "成功刪除分類及其文檔", "success")
}

// AdminDocsHandler 處理文件管理頁面
func AdminDocsHandler(w http.ResponseWriter, r *http.Request) {
	// 從資料庫獲取分類列表
	categories, err := db.GetCategoryList()
	if err != nil {
		log.Println("Error fetching categories:", err)
	}

	// 檢查是否有分類篩選
	var docs []obj.Doc
	var filterCategoryID uint = 0
	var filter string = "all" // 預設顯示所有文件

	// 取得篩選器參數
	categoryIDStr := r.URL.Query().Get("category_id")
	filter = r.URL.Query().Get("filter")
	if filter == "" {
		filter = "all"
	}

	// 根據篩選參數取得文件
	if categoryIDStr != "" {
		categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
		if err == nil {
			filterCategoryID = uint(categoryID)

			// 根據文件狀態篩選
			if filter == "drafts" {
				// 取得指定分類下的草稿
				docs, err = db.GetDraftsByCategory(filterCategoryID)
			} else if filter == "published" {
				// 取得指定分類下的已發布文件
				docs, err = db.GetPublishedDocsByCategory(filterCategoryID)
			} else {
				// 取得指定分類下的所有文件
				docs, err = db.GetDocsByCategory(filterCategoryID)
			}

			if err != nil {
				log.Println("Error fetching docs by category:", err)
			}
		}
	} else {
		// 沒有分類篩選條件，根據狀態篩選
		if filter == "drafts" {
			// 取得所有草稿
			docs, err = db.GetDraftDocs()
		} else if filter == "published" {
			// 取得所有已發布文件
			docs, err = db.GetPublishedDocs()
		} else {
			// 取得所有文件
			docs, err = db.GetAllDocs()
		}

		if err != nil {
			log.Println("Error fetching docs:", err)
		}
	}

	// 檢查是否有訊息要顯示
	message := r.URL.Query().Get("message")
	messageType := r.URL.Query().Get("type")
	if messageType == "" && message != "" {
		messageType = "success" // 預設訊息類型
	}

	// 今天的日期，用於新增文件的默認發布日期
	today := time.Now().Format("2006-01-02")

	// 準備模板資料
	data := map[string]interface{}{
		"Active":           "docs",
		"Categories":       categories,
		"Docs":             docs,
		"FilterCategoryID": filterCategoryID,
		"Filter":           filter,
		"Message":          message,
		"MessageType":      messageType,
		"Today":            today,
	}

	// 解析模板
	tmpl, err := template.ParseFiles(
		"templates/admin/layout.html",
		"templates/admin/docs.html",
	)
	if err != nil {
		log.Println("Template parse error:", err)
		http.Error(w, "Template parse error", http.StatusInternalServerError)
		return
	}

	// 渲染模板
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("Template execute error:", err)
		http.Error(w, "Template execute error", http.StatusInternalServerError)
		return
	}
}

// AdminDocEditHandler 處理編輯文件頁面
func AdminDocEditHandler(w http.ResponseWriter, r *http.Request) {
	// 獲取文件 ID
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		redirectWithMessage(w, r, "/admin/docs", "缺少文件ID", "danger")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		log.Println("Invalid doc ID:", err)
		redirectWithMessage(w, r, "/admin/docs", "無效的文件ID", "danger")
		return
	}

	// 獲取文件詳情
	doc, err := db.GetDoc(uint(id))
	if err != nil {
		log.Println("Error fetching doc:", err)
		redirectWithMessage(w, r, "/admin/docs", "無法找到指定文件", "danger")
		return
	}

	// 獲取所有分類
	categories, err := db.GetCategoryList()
	if err != nil {
		log.Println("Error fetching categories:", err)
	}

	// URL 解碼文件內容
	content, err := url.QueryUnescape(doc.Content)
	if err != nil {
		log.Println("Error decoding content:", err)
		content = doc.Content // 如果解碼失敗，使用原始內容
	}

	// 檢查是否有訊息要顯示
	message := r.URL.Query().Get("message")
	messageType := r.URL.Query().Get("type")
	if messageType == "" && message != "" {
		messageType = "success"
	}

	// 準備模板資料
	data := map[string]interface{}{
		"Active":      "docs",
		"Doc":         doc,
		"DocContent":  content,
		"Categories":  categories,
		"Message":     message,
		"MessageType": messageType,
	}

	// 解析模板
	tmpl, err := template.ParseFiles(
		"templates/admin/layout.html",
		"templates/admin/doc_edit.html",
	)
	if err != nil {
		log.Println("Template parse error:", err)
		http.Error(w, "Template parse error", http.StatusInternalServerError)
		return
	}

	// 渲染模板
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("Template execute error:", err)
		http.Error(w, "Template execute error", http.StatusInternalServerError)
		return
	}
}

// AdminDocAddHandler 處理新增文件
func AdminDocAddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/docs", http.StatusSeeOther)
		return
	}

	// 解析表單
	err := r.ParseForm()
	if err != nil {
		log.Println("Form parse error:", err)
		redirectWithMessage(w, r, "/admin/docs", "表單解析錯誤", "danger")
		return
	}

	// 獲取表單資料
	title := r.FormValue("title")
	categoryIDStr := r.FormValue("category_id")
	publishDate := r.FormValue("publish_date")
	content := r.FormValue("content")
	isDraftStr := r.FormValue("is_draft")

	if title == "" || categoryIDStr == "" || content == "" {
		redirectWithMessage(w, r, "/admin/docs", "標題、分類和內容不能為空", "danger")
		return
	}

	// 解析分類 ID
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
	if err != nil {
		log.Println("Invalid category ID:", err)
		redirectWithMessage(w, r, "/admin/docs", "無效的分類ID", "danger")
		return
	}

	// 確定是否為草稿
	isDraft := isDraftStr == "true"

	// 如果是草稿，發布日期設為空字串；否則，若發布日期為空，使用當前日期
	if isDraft {
		publishDate = ""
	} else if publishDate == "" {
		publishDate = time.Now().Format("2006-01-02")
	}

	// URL 編碼文件內容
	encodedContent := url.QueryEscape(content)

	// 創建新文件
	doc := obj.Doc{
		Title:        title,
		Content:      encodedContent,
		PublishDate:  publishDate,
		LastEditDate: time.Now(),
		CategoryID:   uint(categoryID),
		IsDraft:      isDraft,
	}

	// 保存到資料庫
	err = db.AddDoc(&doc)
	if err != nil {
		log.Println("Error adding doc:", err)
		redirectWithMessage(w, r, "/admin/docs", "新增文件失敗: "+err.Error(), "danger")
		return
	}

	// 顯示成功消息，根據是否為草稿顯示不同內容
	var successMsg string
	if isDraft {
		successMsg = "成功儲存草稿: " + title
	} else {
		successMsg = "成功新增文件: " + title
	}

	// 重定向回文件列表，並帶上成功訊息
	redirectWithMessage(w, r, "/admin/docs", successMsg, "success")
}

// AdminDocUpdateHandler 處理更新文件
func AdminDocUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/docs", http.StatusSeeOther)
		return
	}

	// 解析表單
	err := r.ParseForm()
	if err != nil {
		log.Println("Form parse error:", err)
		redirectWithMessage(w, r, "/admin/docs", "表單解析錯誤", "danger")
		return
	}

	// 獲取表單資料
	idStr := r.FormValue("id")
	title := r.FormValue("title")
	categoryIDStr := r.FormValue("category_id")
	publishDate := r.FormValue("publish_date")
	content := r.FormValue("content")
	isDraftStr := r.FormValue("is_draft")

	if idStr == "" || title == "" || categoryIDStr == "" || content == "" {
		redirectWithMessage(w, r, "/admin/docs", "必填欄位不能為空", "danger")
		return
	}

	// 解析 ID 和分類 ID
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		log.Println("Invalid doc ID:", err)
		redirectWithMessage(w, r, "/admin/docs", "無效的文件ID", "danger")
		return
	}

	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
	if err != nil {
		log.Println("Invalid category ID:", err)
		redirectWithMessage(w, r, "/admin/docs", "無效的分類ID", "danger")
		return
	}

	// URL 編碼文件內容
	encodedContent := url.QueryEscape(content)

	// 確定是否為草稿 - 由於未勾選的複選框不會發送值，所以只有當值為 "true" 時才是草稿
	isDraft := isDraftStr == "true"

	// 取得原有文件的資料來比較狀態和發布日期
	oldDoc, err := db.GetDoc(uint(id))
	if err != nil {
		log.Println("Error fetching original doc:", err)
		redirectWithMessage(w, r, "/admin/docs/edit?id="+idStr, "無法獲取原始文件資料", "danger")
		return
	}

	// 處理發布日期邏輯
	// 1. 如果當前要設為草稿，但發布日期不為空（如已發布過），保留原發布日期
	// 2. 如果從草稿變為公開，且沒有發布日期，設為當前日期
	// 3. 如果仍是草稿，置空發布日期
	if isDraft {
		if !oldDoc.IsDraft && oldDoc.PublishDate != "" {
			// 如果原本是公開的，現在改為草稿，保留原發布日期
			publishDate = oldDoc.PublishDate
		} else {
			// 如果一直是草稿，置空發布日期
			publishDate = ""
		}
	} else if !isDraft && (publishDate == "" || oldDoc.IsDraft) {
		// 如果是從草稿變為公開，且沒有發布日期，使用當前日期
		if oldDoc.PublishDate == "" {
			publishDate = time.Now().Format("2006-01-02")
		} else {
			// 如果已有發布日期，保留原值
			publishDate = oldDoc.PublishDate
		}
	}

	// 更新文件
	doc := obj.Doc{
		ID:           uint(id),
		Title:        title,
		Content:      encodedContent,
		PublishDate:  publishDate,
		LastEditDate: time.Now(),
		CategoryID:   uint(categoryID),
		IsDraft:      isDraft,
	}

	err = db.UpdateDoc(&doc)
	if err != nil {
		log.Println("Error updating doc:", err)
		redirectWithMessage(w, r, "/admin/docs/edit?id="+idStr, "更新文件失敗: "+err.Error(), "danger")
		return
	}

	// 顯示成功消息，根據是否為草稿顯示不同內容
	var successMsg string
	if isDraft {
		successMsg = "草稿已成功更新"
	} else {
		successMsg = "文件已成功更新"
	}

	// 檢查是否需要自動關閉頁面
	autoClose := r.FormValue("autoClose")
	redirectURL := "/admin/docs/edit?id=" + idStr

	if autoClose == "true" {
		// 添加自動關閉參數
		redirectURL += "&autoClose=true"
	}

	// 重定向回編輯頁面，並帶上成功訊息
	redirectWithMessage(w, r, redirectURL, successMsg, "success")
}

// AdminDocDeleteHandler 處理刪除文件
func AdminDocDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/docs", http.StatusSeeOther)
		return
	}

	// 解析表單
	err := r.ParseForm()
	if err != nil {
		log.Println("Form parse error:", err)
		redirectWithMessage(w, r, "/admin/docs", "表單解析錯誤", "danger")
		return
	}

	// 獲取文件 ID
	idStr := r.FormValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		log.Println("Invalid doc ID:", err)
		redirectWithMessage(w, r, "/admin/docs", "無效的文件ID", "danger")
		return
	}

	// 刪除文件
	err = db.DeleteDoc(uint(id))
	if err != nil {
		log.Println("Error deleting doc:", err)
		redirectWithMessage(w, r, "/admin/docs", "刪除文件失敗: "+err.Error(), "danger")
		return
	}

	// 重定向回文件列表，並帶上成功訊息
	redirectWithMessage(w, r, "/admin/docs", "文件已成功刪除", "success")
}

// 輔助函數：帶訊息重定向
func redirectWithMessage(w http.ResponseWriter, r *http.Request, path, message, messageType string) {
	redirectURL := path

	// 檢查路徑是否已包含查詢參數
	containsQuery := strings.Contains(path, "?")

	if message != "" {
		if containsQuery {
			// 如果路徑已經包含 "?"，使用 "&" 連接參數
			redirectURL += "&message=" + url.QueryEscape(message)
		} else {
			// 否則使用 "?" 開始查詢參數
			redirectURL += "?message=" + url.QueryEscape(message)
		}

		if messageType != "" {
			redirectURL += "&type=" + url.QueryEscape(messageType)
		}
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}
