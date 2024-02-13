package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Service interface {
	Health() map[string]string
}

type service struct {
	db *gorm.DB
}

var (
	database = os.Getenv("DB_DATABASE")
	password = os.Getenv("DB_PASSWORD")
	username = os.Getenv("DB_USERNAME")
	port     = os.Getenv("DB_PORT")
	host     = os.Getenv("DB_HOST")
)

// // Define your model structs
// User represents the users table
type User struct {
	// ID              primitive.ObjectID `json:"_id" bson:"_id"`
	UserID         uint   `gorm:"primaryKey"`
	FirstName      string `gorm:"null"`
	LastName       string `gorm:"null"`
	Email          string `gorm:"null;unique"`
	PhoneNumber    string `gorm:"null;unique"`
	Tag            string `gorm:"null;unique"`
	Role           string `gorm:"null"`
	PasswordHash   string `gorm:"null"`
	RefCode        string `gorm:"unique"`
	ProfilePic     string `gorm:"null"`
	EmailVerified  bool
	PhoneVerified  bool
	BvnVerified    bool
	KycStatus      bool
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
	Groups         []Group   `gorm:"foreignKey:CreatedBy"`
	Transactions   []Transaction
	VirtualAccount VirtualAccount
	Banks          []Bank
	Budgets        []Budget
}

// Group represents the groups table
type Group struct {
	GroupID             uint      `gorm:"primaryKey"`
	GroupName           string    `gorm:"not null"`
	GroupTag            string    `gorm:"not null;unique"`
	CreatedBy           uint      `gorm:"not null"`
	CreatedAt           time.Time `gorm:"autoCreateTime"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime"`
	Members             []User    `gorm:"many2many:group_members"`
	GroupVirtualAccount GroupVirtualAccount
}

// GroupMember represents the group_members table
type GroupMember struct {
	GroupMemberID uint `gorm:"primaryKey"`
	GroupID       uint
	UserID        uint
	JoinedAt      time.Time
}

// Transaction represents the transactions table
type Transaction struct {
	TransactionID   uint `gorm:"primaryKey"`
	UserID          uint
	TransactionType string  `gorm:"not null"`
	Amount          float64 `gorm:"not null"`
	Description     string
	TransactionDate time.Time `gorm:"not null"`
}

// VirtualAccount represents the virtual_accounts table
type VirtualAccount struct {
	VirtualAccountID      uint   `gorm:"primaryKey"`
	VirtualAccountBank    string `gorm:"not null"`
	VirtualAccountAccount string `gorm:"not null;unique"`
	VirtualAccountName    string `gorm:"not null"`
	UserID                uint
	Balance               float64 `gorm:"default:0"`
}

// Bank represents the banks table
type Bank struct {
	BankID        uint   `gorm:"primaryKey"`
	BankName      string `gorm:"not null"`
	AccountNumber string `gorm:"not null;unique"`
	AccountName   string `gorm:"not null"`
	UserID        uint
}

// GroupVirtualAccount represents the group_virtual_accounts table
type GroupVirtualAccount struct {
	GroupVirtualAccountID     uint `gorm:"primaryKey"`
	GroupID                   uint
	GroupVirtualAccountBank   string  `gorm:"not null"`
	GroupVirtualAccountNumber string  `gorm:"not null;unique"`
	GroupVirtualAccountName   string  `gorm:"not null"`
	Balance                   float64 `gorm:"default:0"`
}

// Budget represents the budgets table
type Budget struct {
	BudgetID   uint `gorm:"primaryKey"`
	UserID     uint
	BudgetName string    `gorm:"not null"`
	Category   string    `gorm:"not null"`
	Type       string    `gorm:"not null"`
	Amount     float64   `gorm:"not null"`
	StartDate  time.Time `gorm:"not null"`
	EndDate    time.Time `gorm:"not null"`
}

type Otp struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      // Foreign key referencing the User table
	Token     string    `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null"`
}

var DB *gorm.DB

func New() Service {
	var err error
	// dsn := "host=kashin.db.elephantsql.com user=duexzyld password=vs5RpBCO76k96VKR1lAd5vfY-sOlcPNQ dbname=duexzyld port=5432 sslmode=disable"
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, username, password, database, port)
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Uncomment the line below if you want to log SQL statements
	// db = db.Debug()

	// Automatically create or modify tables based on the struct definitions
	// DB.AutoMigrate(
	// 	&User{},
	// 	&Group{},
	// 	&GroupMember{},
	// 	&Transaction{},
	// 	&VirtualAccount{},
	// 	&Bank{},
	// 	&GroupVirtualAccount{},
	// 	&Budget{},
	// 	&Otp{},
	// )

	s := &service{db: DB}
	return s
}

func (s *service) Health() map[string]string {
	return map[string]string{
		"message": "It's healthy",
	}
}
