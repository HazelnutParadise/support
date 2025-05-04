package db

import (
	"os"
	"path"
	"support/obj"
	"time"

	"errors"

	"github.com/glebarez/sqlite" // Pure go SQLite driver, checkout https://github.com/glebarez/sqlite for details
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	DATA_DIR   = "data"            // Directory where the SQLite database file will be stored
	DB_FILE    = "database.sqlite" // Name of the SQLite database file
	UPLOAD_DIR = "uploads"         // Directory where uploaded files will be stored
)

// 保存文件的實際路徑
var UploadStoragePath = path.Join(DATA_DIR, UPLOAD_DIR)

// 圖片訪問的URL路徑
var UploadURLPath = "/" + UPLOAD_DIR + "/"

var db *gorm.DB // Global variable to hold the database connection

func DB() (*gorm.DB, error) {
	if db != nil {
		return db, nil // Return the existing database connection if it exists
	}
	var err error
	if _, err := os.Stat(DATA_DIR); os.IsNotExist(err) {
		err = os.Mkdir(DATA_DIR, 0755) // Create the directory if it doesn't exist
		if err != nil {
			return nil, err
		}
	}
	db, err = gorm.Open(sqlite.Open(path.Join(DATA_DIR, DB_FILE)))
	if err != nil {
		return nil, err
	}

	// 自動遷移所有資料結構
	err = db.AutoMigrate(&obj.Category{}, &obj.Doc{}, &obj.User{}, &obj.AdminSession{}, &obj.Image{})
	if err != nil {
		return nil, err
	}

	// 檢查是否需要創建預設用戶
	var count int64
	db.Model(&obj.User{}).Count(&count)
	if count == 0 {
		// 創建一個預設使用者，預設用戶名和密碼都是 admin
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		defaultUser := &obj.User{
			Username: "admin",
			Password: string(hashedPassword), // 使用雜湊存儲密碼
		}
		db.Create(defaultUser)
	}

	return db, nil
}

// HashPassword 將密碼加密為雜湊值
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerifyPassword 比較密碼和雜湊值
func VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// ChangeUserPassword 修改用戶密碼
func ChangeUserPassword(username, currentPassword, newPassword string) error {
	user, err := GetUserByUsername(username)
	if err != nil {
		return err
	}

	// 驗證當前密碼
	if !VerifyPassword(user.Password, currentPassword) {
		return errors.New("當前密碼不正確")
	}

	// 雜湊新密碼
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	// 更新密碼
	db, err := DB()
	if err != nil {
		return err
	}

	return db.Model(&obj.User{}).Where("username = ?", username).Update("password", hashedPassword).Error
}

// GetCategoryList 獲取所有分類
func GetCategoryList() ([]obj.Category, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var categories []obj.Category
	result := db.Find(&categories)
	return categories, result.Error
}

// GetCategory 獲取特定分類
func GetCategory(id uint) (obj.Category, error) {
	db, err := DB()
	if err != nil {
		return obj.Category{}, err
	}
	var category obj.Category
	result := db.First(&category, id)
	return category, result.Error
}

// GetDocsByCategory 獲取特定分類下的所有文章
func GetDocsByCategory(categoryID uint) ([]obj.Doc, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var docs []obj.Doc
	result := db.Where("category_id = ?", categoryID).Find(&docs)
	return docs, result.Error
}

// GetDoc 獲取特定文章
func GetDoc(id uint) (obj.Doc, error) {
	db, err := DB()
	if err != nil {
		return obj.Doc{}, err
	}
	var doc obj.Doc
	result := db.First(&doc, id)
	return doc, result.Error
}

// SearchDocs 搜尋文章
func SearchDocs(keyword string) ([]obj.Doc, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var docs []obj.Doc
	result := db.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%").Find(&docs)
	return docs, result.Error
}

// GetAllDocs 獲取所有文件
func GetAllDocs() ([]obj.Doc, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var docs []obj.Doc
	result := db.Find(&docs)
	return docs, result.Error
}

// GetPublishedDocs 獲取所有已發布（非草稿）的文件
func GetPublishedDocs() ([]obj.Doc, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var docs []obj.Doc
	result := db.Where("is_draft = ?", false).Find(&docs)
	return docs, result.Error
}

// GetDraftDocs 獲取所有草稿
func GetDraftDocs() ([]obj.Doc, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var docs []obj.Doc
	result := db.Where("is_draft = ?", true).Find(&docs)
	return docs, result.Error
}

// GetPublishedDocsByCategory 獲取特定分類下的所有已發布文章
func GetPublishedDocsByCategory(categoryID uint) ([]obj.Doc, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var docs []obj.Doc
	result := db.Where("category_id = ? AND is_draft = ?", categoryID, false).Find(&docs)
	return docs, result.Error
}

// GetDraftsByCategory 獲取特定分類下的所有草稿
func GetDraftsByCategory(categoryID uint) ([]obj.Doc, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var docs []obj.Doc
	result := db.Where("category_id = ? AND is_draft = ?", categoryID, true).Find(&docs)
	return docs, result.Error
}

// AddCategory 添加新分類
func AddCategory(category *obj.Category) error {
	db, err := DB()
	if err != nil {
		return err
	}
	return db.Create(category).Error
}

// UpdateCategory 更新分類
func UpdateCategory(id uint, name string) error {
	db, err := DB()
	if err != nil {
		return err
	}
	return db.Model(&obj.Category{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":        name,
		"update_time": gorm.Expr("CURRENT_TIMESTAMP"),
	}).Error
}

// DeleteCategory 刪除分類及其所有文件
func DeleteCategory(id uint) error {
	db, err := DB()
	if err != nil {
		return err
	}

	// 開始事務
	tx := db.Begin()

	// 刪除該分類下的所有文件
	if err := tx.Where("category_id = ?", id).Delete(&obj.Doc{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 刪除分類
	if err := tx.Delete(&obj.Category{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事務
	return tx.Commit().Error
}

// AddDoc 添加新文件
func AddDoc(doc *obj.Doc) error {
	db, err := DB()
	if err != nil {
		return err
	}

	// 轉換為資料庫可存儲的格式
	dbDoc := map[string]interface{}{
		"title":       doc.Title,
		"content":     doc.Content,
		"category_id": doc.CategoryID,
		"is_draft":    doc.IsDraft,
	}

	// 處理發布日期
	if doc.PublishDate.Valid {
		dbDoc["publish_date"] = doc.PublishDate.Time
	}

	// 設置最後編輯日期
	if doc.LastEditDate.IsZero() {
		dbDoc["last_edit_date"] = time.Now()
	} else {
		dbDoc["last_edit_date"] = doc.LastEditDate
	}

	// 創建文檔
	result := db.Model(&obj.Doc{}).Create(dbDoc)
	if result.Error != nil {
		return result.Error
	}

	// 獲取插入的 ID
	var lastInsertID int64
	row := db.Raw("SELECT last_insert_rowid()").Row()
	err = row.Scan(&lastInsertID)
	if err != nil {
		return err
	}

	doc.ID = uint(lastInsertID)
	return nil
}

// UpdateDoc 更新文件
func UpdateDoc(doc *obj.Doc) error {
	db, err := DB()
	if err != nil {
		return err
	}

	// 由於 PublishDate 是自訂類型，這裡需要特別處理
	// 我們將使用 map 來進行更新，以便正確處理 DateField 類型
	updates := map[string]interface{}{
		"title":          doc.Title,
		"content":        doc.Content,
		"category_id":    doc.CategoryID,
		"is_draft":       doc.IsDraft,
		"last_edit_date": doc.LastEditDate,
	}

	// 只有當 PublishDate 是有效的，才將其添加到更新中
	if doc.PublishDate.Valid {
		updates["publish_date"] = doc.PublishDate.Time
	} else {
		updates["publish_date"] = nil
	}

	return db.Model(&obj.Doc{}).Where("id = ?", doc.ID).Updates(updates).Error
}

// DeleteDoc 刪除文件
func DeleteDoc(id uint) error {
	db, err := DB()
	if err != nil {
		return err
	}
	return db.Delete(&obj.Doc{}, id).Error
}

// GetUserByUsername 通過用戶名獲取用戶
func GetUserByUsername(username string) (obj.User, error) {
	db, err := DB()
	if err != nil {
		return obj.User{}, err
	}
	var user obj.User
	result := db.Where("username = ?", username).First(&user)
	return user, result.Error
}

// UpdateUserPassword 更新用戶密碼
func UpdateUserPassword(userID uint, newPassword string) error {
	// 對新密碼進行雜湊處理
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	db, err := DB()
	if err != nil {
		return err
	}
	return db.Model(&obj.User{}).Where("id = ?", userID).Update("password", string(hashedPassword)).Error
}

// AdminSession 相關的資料庫操作
// 保存管理員會話到資料庫
func SaveAdminSession(session *obj.AdminSession) error {
	db, err := DB()
	if err != nil {
		return err
	}

	// 刪除該用戶之前的會話
	err = db.Exec("DELETE FROM admin_sessions WHERE username = ?", session.Username).Error
	if err != nil {
		return err
	}

	// 插入新會話
	result := db.Exec("INSERT INTO admin_sessions (session_id, username, expiry) VALUES (?, ?, ?)",
		session.SessionID, session.Username, session.Expiry)
	err = result.Error
	return err
}

// 根據會話ID獲取管理員會話
func GetAdminSession(sessionID string) (*obj.AdminSession, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}

	var session obj.AdminSession
	err = db.Raw("SELECT session_id, username, expiry FROM admin_sessions WHERE session_id = ?", sessionID).
		Scan(&session).Error
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// 刪除管理員會話
func DeleteAdminSession(sessionID string) error {
	db, err := DB()
	if err != nil {
		return err
	}

	err = db.Exec("DELETE FROM admin_sessions WHERE session_id = ?", sessionID).Error
	return err
}

// 清理過期的管理員會話
func CleanExpiredAdminSessions() error {
	db, err := DB()
	if err != nil {
		return err
	}

	err = db.Exec("DELETE FROM admin_sessions WHERE expiry < ?", time.Now()).Error
	return err
}

// ---- 圖片相關功能 ----

// 圖片上傳目錄
var uploadDir = UploadStoragePath

// 確保上傳目錄存在
func EnsureUploadDir() error {
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		return os.MkdirAll(uploadDir, 0755)
	}
	return nil
}

// 添加圖片記錄到資料庫
func AddImage(image *obj.Image) error {
	db, err := DB()
	if err != nil {
		return err
	}
	return db.Create(image).Error
}

// 獲取圖片記錄
func GetImage(id uint) (obj.Image, error) {
	db, err := DB()
	if err != nil {
		return obj.Image{}, err
	}
	var image obj.Image
	result := db.First(&image, id)
	return image, result.Error
}

// 獲取所有圖片
func GetAllImages() ([]obj.Image, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var images []obj.Image
	result := db.Order("upload_time DESC").Find(&images)
	return images, result.Error
}

// 刪除圖片記錄和檔案
func DeleteImage(id uint) error {
	// 先獲取圖片記錄
	image, err := GetImage(id)
	if err != nil {
		return err
	}

	// 刪除實體檔案
	err = os.Remove(image.Path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// 從資料庫刪除記錄
	db, err := DB()
	if err != nil {
		return err
	}
	return db.Delete(&obj.Image{}, id).Error
}

// 根據文件名搜尋圖片
func SearchImagesByFilename(keyword string) ([]obj.Image, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var images []obj.Image
	result := db.Where("filename LIKE ?", "%"+keyword+"%").Order("upload_time DESC").Find(&images)
	return images, result.Error
}
