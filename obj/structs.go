package obj

import "time"

// 這只是示範結構，實際欄位可自行調整
type Category struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Name       string    `json:"name" gorm:"unique"`
	CreateTime time.Time `json:"create_time" gorm:"autoCreateTime"`
	UpdateTime time.Time `json:"update_time" gorm:"autoUpdateTime"`
}

type IndexData struct {
	Title             string
	Categories        []Category
	Docs              []Doc
	CurrentCategory   string
	CurrentCategoryID uint
}

type Doc struct {
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	PublishDate  string    `json:"publish_date"`
	LastEditDate time.Time `json:"last_edit_date" gorm:"autoUpdateTime"`
	CategoryID   uint      `json:"category_id"`
	IsDraft      bool      `json:"is_draft" gorm:"default:false"` // 新增草稿標記
}

// 為了渲染模板，我們再做一個結構把需要的全部資料包起來
type DocPageData struct {
	PageTitle         string
	DocFound          bool
	DocTitle          string
	PublishDate       string
	LastEditDate      string
	HTMLContent       string // 把轉成HTML後的內容存這
	CurrentCategory   string // 給前端JS用
	CurrentCategoryID string
	Categories        []Category
}
