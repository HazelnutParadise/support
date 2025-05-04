package db

import (
	"os"
	"path"
	"support/obj"

	"github.com/glebarez/sqlite" // Pure go SQLite driver, checkout https://github.com/glebarez/sqlite for details
	"gorm.io/gorm"
)

const (
	db_dir  = "data"        // Directory where the SQLite database file will be stored
	db_file = "database.db" // Name of the SQLite database file
)

var db *gorm.DB // Global variable to hold the database connection

func DB() (*gorm.DB, error) {
	if db != nil {
		return db, nil // Return the existing database connection if it exists
	}
	var err error
	if _, err := os.Stat(db_dir); os.IsNotExist(err) {
		err = os.Mkdir(db_dir, 0755) // Create the directory if it doesn't exist
		if err != nil {
			return nil, err
		}
	}
	db, err = gorm.Open(sqlite.Open(path.Join(db_dir, db_file)))
	if err != nil {
		return nil, err
	}

	// 自動遷移所有資料結構
	err = db.AutoMigrate(&obj.Category{}, &obj.Doc{})
	if err != nil {
		return nil, err
	}
	return db, nil
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
	return db.Create(doc).Error
}

// UpdateDoc 更新文件
func UpdateDoc(doc *obj.Doc) error {
	db, err := DB()
	if err != nil {
		return err
	}
	return db.Save(doc).Error
}

// DeleteDoc 刪除文件
func DeleteDoc(id uint) error {
	db, err := DB()
	if err != nil {
		return err
	}
	return db.Delete(&obj.Doc{}, id).Error
}
