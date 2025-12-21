package models

import (
	"time"
)

// User maps to core_customuser
type User struct {
	ID          int64  `gorm:"primaryKey;autoIncrement"`
	Password    string `gorm:"size:128;not null"`
	LastLogin   *time.Time
	IsSuperuser bool      `gorm:"column:is_superuser;default:false"`
	Username    string    `gorm:"size:150;unique;not null"`
	FirstName   string    `gorm:"column:first_name;size:150;not null"`
	LastName    string    `gorm:"column:last_name;size:150;not null"`
	Email       string    `gorm:"size:254;not null"`
	IsStaff     bool      `gorm:"column:is_staff;default:false"`
	IsActive    bool      `gorm:"column:is_active;default:true"`
	IsBanned    bool      `gorm:"column:is_banned;default:false"`
	DateJoined  time.Time `gorm:"column:date_joined;autoCreateTime"`
	PhoneNumber string    `gorm:"column:phone_number;size:15;default:null"`
	CoinBalance int       `gorm:"column:coin_balance;default:0"`
}

// TableName overrides the table name to match Django's
func (User) TableName() string {
	return "core_customuser"
}

type Category struct {
	ID   int64  `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"size:100;not null"`
	Icon string `gorm:"size:50;default:'folder'"`

	SubCategories []SubCategory `gorm:"foreignKey:CategoryID"`
}

func (Category) TableName() string {
	return "core_category"
}

type SubCategory struct {
	ID         int64    `gorm:"primaryKey;autoIncrement"`
	Name       string   `gorm:"size:100;not null"`
	CategoryID int64    `gorm:"column:category_id;not null"`
	Category   Category `gorm:"foreignKey:CategoryID"`
}

func (SubCategory) TableName() string {
	return "core_subcategory"
}

type LostItem struct {
	ID              int64      `gorm:"primaryKey;autoIncrement"`
	Title           string     `gorm:"size:200;not null"`
	Description     string     `gorm:"type:longtext;not null"`
	Image           string     `gorm:"size:100;default:null"`
	Status          string     `gorm:"size:10;default:'LOST'"` // LOST, FOUND, RETURNED
	BountyCoins     int        `gorm:"column:bounty_coins;default:0"`
	IsHighlighted   bool       `gorm:"column:is_highlighted;default:false"`
	HighlightExpiry *time.Time `gorm:"column:highlight_expiry;default:null"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	UserID          int64      `gorm:"column:user_id;not null"`
	CategoryID      int64      `gorm:"column:category_id;default:null"`
	SubCategoryID   *int64     `gorm:"column:subcategory_id;default:null"`
	FinderID        *int64     `gorm:"column:finder_id;default:null"`
	OwnerConfirmed  bool       `gorm:"column:owner_confirmed;default:false"`
	FinderConfirmed bool       `gorm:"column:finder_confirmed;default:false"`
	Location        string     `gorm:"size:255;default:null"`

	User        User         `gorm:"foreignKey:UserID"`
	Category    Category     `gorm:"foreignKey:CategoryID"`
	SubCategory *SubCategory `gorm:"foreignKey:SubCategoryID"`
	Finder      *User        `gorm:"foreignKey:FinderID"`
	Comments    []Comment    `gorm:"foreignKey:ItemID"`
}

func (LostItem) TableName() string {
	return "core_lostitem"
}

type LostItemImage struct {
	ItemID      int64  `gorm:"primaryKey"`
	ImageData   []byte `gorm:"type:longblob"`
	ContentType string `gorm:"size:50"`
}

func (LostItemImage) TableName() string {
	return "core_lostitem_image"
}

type Comment struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Content   string    `gorm:"type:longtext;not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UserID    int64     `gorm:"column:user_id;not null"`
	ItemID    int64     `gorm:"column:item_id;not null"`

	User User     `gorm:"foreignKey:UserID"`
	Item LostItem `gorm:"foreignKey:ItemID"`
}

func (Comment) TableName() string {
	return "core_comment"
}

type CoinTransaction struct {
	ID              int64     `gorm:"primaryKey;autoIncrement"`
	Amount          int       `gorm:"not null"`
	TransactionType string    `gorm:"column:transaction_type;size:20;not null"`
	Timestamp       time.Time `gorm:"column:timestamp;autoCreateTime"`
	UserID          int64     `gorm:"column:user_id;not null"`

	User User `gorm:"foreignKey:UserID"`
}

func (CoinTransaction) TableName() string {
	return "core_cointransaction"
}

type ItemClaim struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	ItemID    int64     `gorm:"column:item_id;not null"`
	UserID    int64     `gorm:"column:user_id;not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`

	Item LostItem `gorm:"foreignKey:ItemID"`
	User User     `gorm:"foreignKey:UserID"`
}

func (ItemClaim) TableName() string {
	return "core_itemclaim"
}

type ItemReport struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	ItemID      int64     `gorm:"column:item_id;not null"`
	ReporterID  int64     `gorm:"column:reporter_id;not null"`
	Reason      string    `gorm:"size:50;not null"` // fraud, spam, buying_selling, inappropriate, other
	Description string    `gorm:"type:text"`
	Status      string    `gorm:"size:20;default:'pending'"` // pending, reviewed, resolved
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`

	Item     LostItem `gorm:"foreignKey:ItemID"`
	Reporter User     `gorm:"foreignKey:ReporterID"`
}

func (ItemReport) TableName() string {
	return "core_itemreport"
}

type Notification struct {
	ID              int64     `gorm:"primaryKey;autoIncrement"`
	UserID          int64     `gorm:"column:user_id;not null"`
	Type            string    `gorm:"size:30;not null"` // report, warning, system_update
	Title           string    `gorm:"size:200;not null"`
	Message         string    `gorm:"type:text;not null"`
	IsRead          bool      `gorm:"column:is_read;default:false"`
	ReferenceURL    string    `gorm:"column:reference_url;size:255"`
	RelatedItemID   *int64    `gorm:"column:related_item_id"`
	RelatedReportID *int64    `gorm:"column:related_report_id"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`

	User          User        `gorm:"foreignKey:UserID"`
	RelatedItem   *LostItem   `gorm:"foreignKey:RelatedItemID"`
	RelatedReport *ItemReport `gorm:"foreignKey:RelatedReportID"`
}

func (Notification) TableName() string {
	return "core_notification"
}

type TopUpTransaction struct {
	ID              int64      `gorm:"primaryKey;autoIncrement"`
	OrderID         string     `gorm:"column:order_id;size:100;unique;not null"`
	UserID          int64      `gorm:"column:user_id;not null"`
	Amount          int        `gorm:"not null"`                  // coin amount (100, 500, 1000)
	Price           int        `gorm:"not null"`                  // price in IDR
	Status          string     `gorm:"size:20;default:'pending'"` // pending, success, failed, expired
	PaymentType     string     `gorm:"column:payment_type;size:50"`
	TransactionTime *time.Time `gorm:"column:transaction_time"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime"`

	User User `gorm:"foreignKey:UserID"`
}

func (TopUpTransaction) TableName() string {
	return "core_topuptransaction"
}

type WithdrawalRequest struct {
	ID            int64      `gorm:"primaryKey;autoIncrement"`
	UserID        int64      `gorm:"column:user_id;not null"`
	Amount        int        `gorm:"not null"` // saldo IDR
	Coins         int        `gorm:"not null"` // jumlah coin yang direseve/dikonversi
	Method        string     `gorm:"size:30;not null"`
	AccountName   string     `gorm:"size:100;not null"`
	AccountNumber string     `gorm:"size:50;not null"`
	Status        string     `gorm:"size:20;default:'pending'"`
	Note          string     `gorm:"type:text"`
	ProcessedAt   *time.Time `gorm:"column:processed_at"`
	CreatedAt     time.Time  `gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime"`
	User          User       `gorm:"foreignKey:UserID"`
}

func (WithdrawalRequest) TableName() string {
	return "core_withdrawalrequest"
}
