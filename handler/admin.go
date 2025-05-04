package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"support/db"
	"support/obj"
	"text/template"
	"time"
)

// 常數
const (
	AdminSessionCookieName = "admin_session"
	AdminSessionTimeout    = 12 * time.Hour // 會話超時時間（12小時）
)

// AdminLoginHandler 處理管理員登入
func AdminLoginHandler(w http.ResponseWriter, r *http.Request) {
	// 如果是GET請求，顯示登入頁面
	if r.Method == http.MethodGet {
		// 檢查用戶是否已經登入
		_, err := getAdminSession(r)
		if err == nil {
			// 已登入，重定向到儀表板
			http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
			return
		}

		// 解析登入模板
		tmpl, err := template.ParseFiles("templates/admin/login.html")
		if err != nil {
			log.Println("Login template parse error:", err)
			http.Error(w, "模板解析錯誤", http.StatusInternalServerError)
			return
		}

		// 渲染登入頁面
		err = tmpl.Execute(w, nil)
		if err != nil {
			log.Println("Login template execute error:", err)
			http.Error(w, "模板執行錯誤", http.StatusInternalServerError)
		}
		return
	}

	// 處理POST登入請求
	if r.Method == http.MethodPost {
		// 解析表單
		err := r.ParseForm()
		if err != nil {
			log.Println("Login form parse error:", err)
			showLoginError(w, r, "表單解析錯誤")
			return
		}

		// 獲取表單數據
		username := r.FormValue("username")
		password := r.FormValue("password")

		// 驗證表單數據
		if username == "" || password == "" {
			showLoginError(w, r, "用戶名和密碼不能為空")
			return
		}

		// 從資料庫獲取用戶
		user, err := db.GetUserByUsername(username)
		if err != nil {
			log.Println("User query error:", err)
			showLoginError(w, r, "用戶名或密碼錯誤")
			return
		}

		// 驗證密碼 - 使用雜湊密碼驗證
		if !db.VerifyPassword(user.Password, password) {
			showLoginError(w, r, "用戶名或密碼錯誤")
			return
		}

		// 驗證成功，創建會話
		sessionID := generateSessionID()
		expiry := time.Now().Add(AdminSessionTimeout)

		// 設置會話Cookie
		cookie := &http.Cookie{
			Name:     AdminSessionCookieName,
			Value:    sessionID,
			Expires:  expiry,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, cookie)

		// 保存會話到資料庫 (這個功能需要在db包中實現)
		session := &obj.AdminSession{
			SessionID: sessionID,
			Username:  username,
			Expiry:    expiry,
		}

		err = db.SaveAdminSession(session)
		if err != nil {
			log.Println("Session save error:", err)
			showLoginError(w, r, "創建會話失敗")
			return
		}

		// 重定向到儀表板
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	// 不支持的HTTP方法
	http.Error(w, "不支持的HTTP方法", http.StatusMethodNotAllowed)
}

// AdminLogoutHandler 處理管理員登出
func AdminLogoutHandler(w http.ResponseWriter, r *http.Request) {
	// 從Cookie中獲取會話ID
	cookie, err := r.Cookie(AdminSessionCookieName)
	if err == nil {
		// 從資料庫刪除會話
		db.DeleteAdminSession(cookie.Value)
	}

	// 清除Cookie
	expiredCookie := &http.Cookie{
		Name:     AdminSessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	}
	http.SetCookie(w, expiredCookie)

	// 重定向到登入頁面
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

// AuthMiddleware 檢查管理員是否已登入的中間件
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 檢查是否存在會話
		session, err := getAdminSession(r)
		if err != nil {
			// 無有效會話，重定向到登入頁面
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}

		// 檢查會話是否過期
		if time.Now().After(session.Expiry) {
			// 會話已過期，刪除會話並重定向到登入頁面
			db.DeleteAdminSession(session.SessionID)

			// 清除Cookie
			expiredCookie := &http.Cookie{
				Name:     AdminSessionCookieName,
				Value:    "",
				Path:     "/",
				Expires:  time.Unix(0, 0),
				HttpOnly: true,
			}
			http.SetCookie(w, expiredCookie)

			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}

		// 會話有效，調用下一個處理器
		next(w, r)
	}
}

// 從請求中獲取管理員會話
func getAdminSession(r *http.Request) (*obj.AdminSession, error) {
	// 從Cookie中獲取會話ID
	cookie, err := r.Cookie(AdminSessionCookieName)
	if err != nil {
		return nil, err
	}

	// 從資料庫獲取會話
	sessionID := cookie.Value
	session, err := db.GetAdminSession(sessionID)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// 生成唯一的會話ID
func generateSessionID() string {
	// 生成當前時間的納秒級時間戳，轉換為字符串
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

// 顯示登入錯誤
func showLoginError(w http.ResponseWriter, r *http.Request, errorMessage string) {
	tmpl, err := template.ParseFiles("templates/admin/login.html")
	if err != nil {
		log.Println("Login template parse error:", err)
		http.Error(w, "模板解析錯誤", http.StatusInternalServerError)
		return
	}

	// 渲染登入頁面，並顯示錯誤訊息
	data := map[string]interface{}{
		"ErrorMessage": errorMessage,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("Login template execute error:", err)
		http.Error(w, "模板執行錯誤", http.StatusInternalServerError)
	}
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

// AdminDashboardHandler 處理管理面板首頁
func AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	// 獲取當前的管理員會話
	session, err := getAdminSession(r)
	if err != nil {
		log.Println("Session error:", err)
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

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
		"Username":      session.Username,
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
	// 獲取當前的管理員會話
	session, err := getAdminSession(r)
	if err != nil {
		log.Println("Session error:", err)
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

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
		"Username":    session.Username,
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
	// 獲取當前的管理員會話
	session, err := getAdminSession(r)
	if err != nil {
		log.Println("Session error:", err)
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	// 獲取篩選條件
	categoryIDStr := r.URL.Query().Get("category_id")
	filter := r.URL.Query().Get("filter")
	if filter == "" {
		filter = "all" // 默認顯示所有文件
	}

	var docs []obj.Doc
	var categoryID uint = 0

	// 解析分類ID（如果提供）
	if categoryIDStr != "" {
		id, err := strconv.ParseUint(categoryIDStr, 10, 32)
		if err == nil {
			categoryID = uint(id)
		}
	}

	// 根據篩選條件獲取文件
	if categoryID > 0 {
		// 有分類篩選
		switch filter {
		case "published":
			// 取得指定分類的已發布文件
			docs, err = db.GetPublishedDocsByCategory(categoryID)
		case "drafts":
			// 取得指定分類的草稿
			docs, err = db.GetDraftsByCategory(categoryID)
		default:
			// 取得指定分類的所有文件
			docs, err = db.GetDocsByCategory(categoryID)
		}
	} else {
		// 無分類篩選
		switch filter {
		case "published":
			// 取得所有已發布文件
			docs, err = db.GetPublishedDocs()
		case "drafts":
			// 取得所有草稿
			docs, err = db.GetDraftDocs()
		default:
			// 取得所有文件
			docs, err = db.GetAllDocs()
		}
	}

	if err != nil {
		log.Println("Error fetching docs:", err)
	}

	// 獲取分類，用於下拉過濾器
	categories, err := db.GetCategoryList()
	if err != nil {
		log.Println("Error fetching categories:", err)
	}

	// 獲取URL查詢參數
	message := r.URL.Query().Get("message")
	messageType := r.URL.Query().Get("type")

	// 準備今天的日期，用於新增文件的默認發布日期
	today := time.Now().Format("2006-01-02")

	// 準備模板數據
	data := map[string]interface{}{
		"Active":           "docs",
		"Docs":             docs,
		"Categories":       categories,
		"Message":          message,
		"MessageType":      messageType,
		"Username":         session.Username,
		"Today":            today,
		"Filter":           filter,
		"FilterCategoryID": categoryID,
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
	// 獲取當前的管理員會話
	session, err := getAdminSession(r)
	if err != nil {
		log.Println("Session error:", err)
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	// 從 URL 參數中獲取文件 ID
	docIDStr := r.URL.Query().Get("id")
	var doc obj.Doc
	var isNewDoc bool = false

	if docIDStr == "" || docIDStr == "0" {
		// 新增文件的情況
		isNewDoc = true
		doc = obj.Doc{
			PublishDate: func() obj.DateField {
				var dateField obj.DateField
				dateField.FromTime(time.Now())
				return dateField
			}(),
			IsDraft: true, // 默認為草稿
		}
	} else {
		// 編輯現有文件的情況
		docID, err := strconv.ParseUint(docIDStr, 10, 32)
		if err != nil {
			log.Println("Invalid doc ID:", err)
			http.Redirect(w, r, "/admin/docs", http.StatusSeeOther)
			return
		}

		// 從資料庫獲取文件
		doc, err = db.GetDoc(uint(docID))
		if err != nil {
			log.Println("Error fetching doc:", err)
			http.Redirect(w, r, "/admin/docs", http.StatusSeeOther)
			return
		}

		// 解碼文件內容，用於編輯
		decodedContent, err := url.QueryUnescape(doc.Content)
		if err != nil {
			log.Println("URL decode error:", err)
			// 如果解碼失敗，仍使用原始內容
		} else {
			doc.Content = decodedContent
		}
	}

	if r.Method == "POST" {
		// 處理表單提交
		r.ParseForm()

		// 從表單獲取數據
		title := r.PostForm.Get("title")
		categoryIDStr := r.PostForm.Get("category_id")
		publishDateStr := r.PostForm.Get("publish_date")
		content := r.PostForm.Get("content")
		isDraftStr := r.PostForm.Get("is_draft")

		// 驗證必填欄位
		if title == "" || categoryIDStr == "" || content == "" {
			log.Println("Missing required fields")

			// 獲取分類列表用於表單選擇
			categories, _ := db.GetCategoryList()

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

			// 準備錯誤訊息
			data := map[string]interface{}{
				"Active":     "docs",
				"Doc":        doc,
				"IsNewDoc":   isNewDoc,
				"DocContent": doc.Content,
				"Categories": categories,
				"Error":      "請填寫所有必填欄位",
				"Username":   session.Username,
			}

			// 渲染模板
			err = tmpl.Execute(w, data)
			if err != nil {
				log.Println("Template execute error:", err)
				http.Error(w, "Template execute error", http.StatusInternalServerError)
			}
			return
		}

		// 解析分類 ID
		categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
		if err != nil {
			log.Println("Invalid category ID:", err)
			http.Redirect(w, r, "/admin/docs", http.StatusSeeOther)
			return
		}

		// 解析發布日期
		var publishDate string
		if publishDateStr != "" {
			pubDate, err := time.Parse("2006-01-02", publishDateStr)
			if err != nil {
				log.Println("Invalid publish date:", err)
				publishDate = time.Now().Format("2006-01-02") // 使用當前日期作為默認值
			} else {
				publishDate = pubDate.Format("2006-01-02")
			}
		}

		// 判斷是否為草稿
		isDraft := isDraftStr == "true"

		// URL 編碼文件內容
		encodedContent := url.QueryEscape(content)

		// 更新文件對象
		doc.Title = title
		doc.CategoryID = uint(categoryID)
		doc.PublishDate = obj.DateField{}
		err = doc.PublishDate.FromString(publishDate)
		if err != nil {
			log.Println("Invalid publish date:", err)
			doc.PublishDate.FromTime(time.Now()) // Default to current time if parsing fails
		}
		doc.Content = encodedContent
		doc.IsDraft = isDraft

		// 保存到數據庫
		if isNewDoc {
			// 創建新文件
			err = db.AddDoc(&doc)
			if err != nil {
				log.Println("Error creating doc:", err)
				http.Redirect(w, r, "/admin/docs?message=文件創建失敗&type=danger", http.StatusSeeOther)
				return
			}
			http.Redirect(w, r, "/admin/docs?message=文件創建成功&type=success", http.StatusSeeOther)
		} else {
			// 更新現有文件
			err = db.UpdateDoc(&doc)
			if err != nil {
				log.Println("Error updating doc:", err)
				http.Redirect(w, r, "/admin/docs?message=文件更新失敗&type=danger", http.StatusSeeOther)
				return
			}
			http.Redirect(w, r, "/admin/docs?message=文件更新成功&type=success", http.StatusSeeOther)
		}
		return
	}

	// GET 請求 - 顯示編輯表單
	// 獲取分類列表用於表單選擇
	categories, err := db.GetCategoryList()
	if err != nil {
		log.Println("Error fetching categories:", err)
	}

	// 準備模板數據
	data := map[string]interface{}{
		"Active":     "docs",
		"Doc":        doc,
		"DocContent": doc.Content, // 添加 DocContent 變數以符合模板中的引用方式
		"IsNewDoc":   isNewDoc,
		"Categories": categories,
		"Username":   session.Username,
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
	publishDateStr := r.FormValue("publish_date")
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

	// 處理發布日期
	var publishDate obj.DateField
	if isDraft {
		// 如果是草稿，發布日期可以為空
		publishDate.Valid = false
	} else if publishDateStr != "" {
		// 非草稿，且有提供日期
		err = publishDate.FromString(publishDateStr)
		if err != nil {
			log.Println("Invalid publish date:", err)
			// 日期無效，使用當前日期
			publishDate.FromTime(time.Now())
		}
	} else {
		// 非草稿，未提供日期，使用當前日期
		publishDate.FromTime(time.Now())
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
	publishDateStr := r.FormValue("publish_date")
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

	// 處理發布日期
	var publishDate obj.DateField
	if isDraft {
		// 如果是草稿
		if !oldDoc.IsDraft && oldDoc.PublishDate.Valid {
			// 如果原本是公開的，現在改為草稿，保留原發布日期
			publishDate = oldDoc.PublishDate
		} else {
			// 如果一直是草稿，置空發布日期
			publishDate.Valid = false
		}
	} else {
		// 如果非草稿
		if publishDateStr != "" {
			// 有提供日期
			err = publishDate.FromString(publishDateStr)
			if err != nil {
				log.Println("Invalid publish date:", err)
				// 日期無效但非草稿，使用當前日期
				publishDate.FromTime(time.Now())
			}
		} else if oldDoc.PublishDate.Valid {
			// 沒提供日期，但原有日期有效，保留原日期
			publishDate = oldDoc.PublishDate
		} else {
			// 沒提供日期，原日期也無效，使用當前日期
			publishDate.FromTime(time.Now())
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

// AdminChangePasswordHandler 處理密碼修改
func AdminChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	// 獲取當前登入的用戶
	session, err := getAdminSession(r)
	if err != nil {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	// 如果是GET請求，顯示密碼修改頁面
	if r.Method == http.MethodGet {
		// 檢查是否有訊息要顯示
		message := r.URL.Query().Get("message")
		messageType := r.URL.Query().Get("type")
		if messageType == "" && message != "" {
			messageType = "info" // 預設訊息類型
		}

		// 準備模板資料
		data := map[string]interface{}{
			"Active":      "change_password",
			"Username":    session.Username,
			"Message":     message,
			"MessageType": messageType,
		}

		// 解析模板
		tmpl, err := template.ParseFiles(
			"templates/admin/layout.html",
			"templates/admin/change_password.html",
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
		return
	}

	// 處理POST請求 - 修改密碼
	if r.Method == http.MethodPost {
		// 解析表單
		err := r.ParseForm()
		if err != nil {
			log.Println("Form parse error:", err)
			redirectWithMessage(w, r, "/admin/change-password", "表單解析錯誤", "danger")
			return
		}

		// 獲取表單數據 - 修正欄位名稱匹配問題
		currentPassword := r.FormValue("currentPassword")
		newPassword := r.FormValue("newPassword")
		confirmPassword := r.FormValue("confirmPassword")

		// 驗證表單數據
		if currentPassword == "" || newPassword == "" || confirmPassword == "" {
			redirectWithMessage(w, r, "/admin/change-password", "所有欄位都必須填寫", "danger")
			return
		}

		// 檢查新密碼和確認密碼是否匹配
		if newPassword != confirmPassword {
			redirectWithMessage(w, r, "/admin/change-password", "新密碼與確認密碼不匹配", "danger")
			return
		}

		// 檢查新密碼長度
		if len(newPassword) < 6 {
			redirectWithMessage(w, r, "/admin/change-password", "新密碼長度必須至少為6個字符", "danger")
			return
		}

		// 修改密碼
		err = db.ChangeUserPassword(session.Username, currentPassword, newPassword)
		if err != nil {
			log.Println("Password change error:", err)
			redirectWithMessage(w, r, "/admin/change-password", "密碼修改失敗: "+err.Error(), "danger")
			return
		}

		// 重定向到儀表板，並顯示成功訊息
		redirectWithMessage(w, r, "/admin/dashboard", "密碼已成功修改", "success")
		return
	}

	// 不支持的HTTP方法
	http.Error(w, "不支持的HTTP方法", http.StatusMethodNotAllowed)
}

// ----- 圖片相關處理器 -----

// AdminImagesHandler 處理圖片管理頁面
func AdminImagesHandler(w http.ResponseWriter, r *http.Request) {
	// 獲取當前的管理員會話
	session, err := getAdminSession(r)
	if err != nil {
		log.Println("Session error:", err)
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	// 獲取所有圖片
	images, err := db.GetAllImages()
	if err != nil {
		log.Println("Error fetching images:", err)
	}

	// 檢查是否有訊息要顯示
	message := r.URL.Query().Get("message")
	messageType := r.URL.Query().Get("type")
	if messageType == "" && message != "" {
		messageType = "info" // 預設訊息類型
	}

	// 準備模板資料
	data := obj.ImageListData{
		Images:   images,
		Message:  message,
		MsgType:  messageType,
		Username: session.Username,
	}

	// 創建自定義模板函數
	funcMap := template.FuncMap{
		"divideSize": func(size int64) float64 {
			return float64(size) / 1024.0 // 轉換為KB
		},
	}

	// 解析模板
	tmpl, err := template.New("layout.html").Funcs(funcMap).ParseFiles(
		"templates/admin/layout.html",
		"templates/admin/images.html",
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

// AdminImageUploadHandler 處理圖片上傳
func AdminImageUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 如果不是 POST 請求，返回錯誤
	if r.Method != http.MethodPost {
		http.Error(w, "僅支持 POST 請求", http.StatusMethodNotAllowed)
		return
	}

	// 檢查會話
	_, err := getAdminSession(r)
	if err != nil {
		http.Error(w, "未授權", http.StatusUnauthorized)
		return
	}

	// 解析表單，限制內存使用為 32MB，超過會存到臨時文件
	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		log.Println("Parse form error:", err)
		http.Error(w, "表單解析錯誤", http.StatusBadRequest)
		return
	}

	// 獲取上傳的文件
	file, header, err := r.FormFile("image")
	if err != nil {
		log.Println("Getting form file error:", err)
		http.Error(w, "獲取文件失敗", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 檢查文件類型
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		http.Error(w, "僅支持圖片文件", http.StatusBadRequest)
		return
	}

	// 獲取文件擴展名
	extension := ""
	if strings.HasSuffix(header.Filename, ".jpg") || strings.HasSuffix(header.Filename, ".jpeg") {
		extension = ".jpg"
	} else if strings.HasSuffix(header.Filename, ".png") {
		extension = ".png"
	} else if strings.HasSuffix(header.Filename, ".gif") {
		extension = ".gif"
	} else if strings.HasSuffix(header.Filename, ".webp") {
		extension = ".webp"
	} else {
		// 根據 MIME 類型推斷擴展名
		switch contentType {
		case "image/jpeg":
			extension = ".jpg"
		case "image/png":
			extension = ".png"
		case "image/gif":
			extension = ".gif"
		case "image/webp":
			extension = ".webp"
		default:
			extension = ".jpg" // 默認 jpg
		}
	}

	// 生成唯一文件名
	filename := generateUniqueFilename() + extension

	// 確保上傳目錄存在
	err = db.EnsureUploadDir()
	if err != nil {
		log.Println("Error creating upload directory:", err)
		http.Error(w, "創建上傳目錄失敗", http.StatusInternalServerError)
		return
	}

	// 創建文件路徑 (data/uploads/filename) - 實際存儲位置
	filePath := path.Join(db.UploadStoragePath, filename)
	dst, err := os.Create(filePath)
	if err != nil {
		log.Println("File creation error:", err)
		http.Error(w, "創建文件失敗", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// 複製上傳的文件內容到新文件
	written, err := io.Copy(dst, file)
	if err != nil {
		log.Println("File copy error:", err)
		http.Error(w, "保存文件失敗", http.StatusInternalServerError)
		return
	}

	// 生成圖片 URL (相對路徑) - 為了與前端兼容，使用 /uploads/ 作為URL前綴
	imageURL := db.UploadURLPath + filename

	// 創建圖片記錄
	image := obj.Image{
		Filename:    header.Filename,
		Path:        filePath,
		URL:         imageURL,
		Size:        written,
		ContentType: contentType,
		UploadTime:  time.Now(),
	}

	// 保存到資料庫
	err = db.AddImage(&image)
	if err != nil {
		log.Println("Image DB save error:", err)
		http.Error(w, "保存圖片記錄失敗", http.StatusInternalServerError)
		return
	}

	// 根據請求來源返回不同的響應
	source := r.FormValue("source")
	if source == "editor" {
		// 從編輯器上傳，返回 JSON 響應供編輯器使用
		w.Header().Set("Content-Type", "application/json")
		jsonResponse := map[string]interface{}{
			"success": true,
			"url":     imageURL,
			"id":      image.ID,
		}
		json.NewEncoder(w).Encode(jsonResponse)
	} else {
		// 從圖片管理頁面上傳，重定向回圖片列表
		redirectWithMessage(w, r, "/admin/images", "圖片上傳成功", "success")
	}
}

// AdminImageDeleteHandler 處理圖片刪除
func AdminImageDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/images", http.StatusSeeOther)
		return
	}

	// 解析表單
	err := r.ParseForm()
	if err != nil {
		log.Println("Form parse error:", err)
		redirectWithMessage(w, r, "/admin/images", "表單解析錯誤", "danger")
		return
	}

	// 獲取圖片 ID
	idStr := r.FormValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		log.Println("Invalid image ID:", err)
		redirectWithMessage(w, r, "/admin/images", "無效的圖片ID", "danger")
		return
	}

	// 刪除圖片
	err = db.DeleteImage(uint(id))
	if err != nil {
		log.Println("Error deleting image:", err)
		redirectWithMessage(w, r, "/admin/images", "刪除圖片失敗: "+err.Error(), "danger")
		return
	}

	// 重定向回圖片列表，並帶上成功訊息
	redirectWithMessage(w, r, "/admin/images", "圖片已成功刪除", "success")
}

// 生成唯一的文件名
func generateUniqueFilename() string {
	// 使用時間戳和隨機數生成唯一文件名
	timestamp := time.Now().UnixNano()
	random := rand.Intn(10000)
	return fmt.Sprintf("%d_%d", timestamp, random)
}
