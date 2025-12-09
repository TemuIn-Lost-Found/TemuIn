package models

import (
	"time"
)

// User maps to core_customuser
type User struct {
    ID          int64     `gorm:"primaryKey;autoIncrement"`
    Password    string    `gorm:"size:128;not null"`
    LastLogin   *time.Time
    IsSuperuser bool      `gorm:"column:is_superuser;default:false"`
    Username    string    `gorm:"size:150;unique;not null"`
    FirstName   string    `gorm:"column:first_name;size:150;not null"`
    LastName    string    `gorm:"column:last_name;size:150;not null"`
    Email       string    `gorm:"size:254;not null"`
    IsStaff     bool      `gorm:"column:is_staff;default:false"`
    IsActive    bool      `gorm:"column:is_active;default:true"`
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
    ID         int64  `gorm:"primaryKey;autoIncrement"`
    Name       string `gorm:"size:100;not null"`
    CategoryID int64  `gorm:"column:category_id;not null"`
    Category   Category `gorm:"foreignKey:CategoryID"`
}

func (SubCategory) TableName() string {
    return "core_subcategory"
}

type LostItem struct {
    ID              int64     `gorm:"primaryKey;autoIncrement"`
    Title           string    `gorm:"size:200;not null"`
    Description     string    `gorm:"type:longtext;not null"`
    Image           string    `gorm:"size:100;default:null"`
    Status          string    `gorm:"size:10;default:'LOST'"` // LOST, FOUND, RETURNED
    BountyCoins     int       `gorm:"column:bounty_coins;default:0"`
    IsHighlighted   bool      `gorm:"column:is_highlighted;default:false"`
    HighlightExpiry *time.Time `gorm:"column:highlight_expiry;default:null"`
    CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
    UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime"`
    UserID          int64     `gorm:"column:user_id;not null"`
    CategoryID      int64     `gorm:"column:category_id;default:null"`
    SubCategoryID   *int64    `gorm:"column:subcategory_id;default:null"`
    FinderID        *int64    `gorm:"column:finder_id;default:null"`
    OwnerConfirmed  bool      `gorm:"column:owner_confirmed;default:false"`
    FinderConfirmed bool      `gorm:"column:finder_confirmed;default:false"`
    Location        string    `gorm:"size:255;default:null"`
    
    User        User        `gorm:"foreignKey:UserID"`
    Category    Category    `gorm:"foreignKey:CategoryID"`
    SubCategory *SubCategory `gorm:"foreignKey:SubCategoryID"`
    Finder      *User       `gorm:"foreignKey:FinderID"`
    Comments    []Comment   `gorm:"foreignKey:ItemID"`
}

func (LostItem) TableName() string {
    return "core_lostitem"
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
