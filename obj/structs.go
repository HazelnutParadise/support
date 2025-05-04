package obj

import (
	"database/sql/driver"
	"errors"
	"time"
)

// DateField 自定義類型，用於處理資料庫中混合型態的日期欄位
// 能夠兼容資料庫中的時間類型或字串類型
type DateField struct {
	Time  time.Time
	Valid bool // 指示是否為有效時間
}

// Scan 實現 sql.Scanner 接口，用於從資料庫讀取時的轉換
func (d *DateField) Scan(value interface{}) error {
	if value == nil {
		d.Time, d.Valid = time.Time{}, false
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		// 直接是時間類型
		d.Time, d.Valid = v, true
	case []byte:
		// 是字節陣列 (可能來自資料庫中的字串)
		t, err := time.Parse("2006-01-02", string(v))
		if err != nil {
			d.Valid = false
			return err
		}
		d.Time, d.Valid = t, true
	case string:
		// 是字串
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			d.Valid = false
			return err
		}
		d.Time, d.Valid = t, true
	default:
		d.Valid = false
		return errors.New("不支援的類型")
	}
	return nil
}

// Value 實現 driver.Valuer 接口，用於寫入資料庫時的轉換
func (d DateField) Value() (driver.Value, error) {
	if !d.Valid {
		return nil, nil
	}
	return d.Time, nil
}

// String 返回日期的字串格式
func (d DateField) String() string {
	if !d.Valid {
		return ""
	}
	return d.Time.Format("2006-01-02")
}

// Format 以指定格式返回日期字串
func (d DateField) Format(layout string) string {
	if !d.Valid {
		return ""
	}
	return d.Time.Format(layout)
}

// FromString 從字串設置時間值
func (d *DateField) FromString(s string) error {
	if s == "" {
		d.Time, d.Valid = time.Time{}, false
		return nil
	}

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		d.Valid = false
		return err
	}

	d.Time, d.Valid = t, true
	return nil
}

// FromTime 從 time.Time 設置時間值
func (d *DateField) FromTime(t time.Time) {
	var zero time.Time
	if t == zero {
		d.Time, d.Valid = time.Time{}, false
		return
	}
	d.Time, d.Valid = t, true
}

// 這只是示範結構，實際欄位可自行調整
type Category struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Name       string    `json:"name" gorm:"unique"`
	CreateTime time.Time `json:"create_time" gorm:"autoCreateTime"`
	UpdateTime time.Time `json:"update_time" gorm:"autoUpdateTime"`
}

// User 管理員使用者
type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Username string `json:"username" gorm:"unique"`
	Password string `json:"password"`
}

// AdminSession 管理員會話
type AdminSession struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	SessionID string    `json:"session_id" gorm:"unique"`
	Username  string    `json:"username"`
	Expiry    time.Time `json:"expiry"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
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
	PublishDate  DateField `json:"publish_date"` // 使用 DateField 類型，可兼容時間和字串
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

// Image 圖片資料結構
type Image struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Filename    string    `json:"filename"`
	Path        string    `json:"path"`
	URL         string    `json:"url"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	UploadTime  time.Time `json:"upload_time" gorm:"autoCreateTime"`
}

// ImageListData 圖片列表頁面資料
type ImageListData struct {
	Images   []Image
	Message  string
	MsgType  string
	Username string
	Active   string // 添加 Active 字段，用於控制側邊欄選中狀態
}
