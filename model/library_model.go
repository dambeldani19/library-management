package model

import (
	"database/sql"
	"time"
)

type Borrower struct {
	ID    int32  `gorm:"primaryKey"`
	Name  string `gorm:"size:255"`
	Email string `gorm:"size:255;unique"`
}

type BorrowingTransaction struct {
	ID         int32    `gorm:"primaryKey"`
	BorrowerID int32    // Foreign key for Borrower
	Borrower   Borrower `gorm:"foreignKey:BorrowerID"` // Specify the foreign key relationship
	BookID     int32    // Foreign key for Book
	Book       Book     `gorm:"foreignKey:BookID"` // Specify the foreign key relationship
	BorrowedAt string
	DueDate    string
	ReturnedAt sql.NullString // Nullable, use pointer for nullable fields
	Status     string         `gorm:"size:50"`
}

type Book struct {
	ID              int32    `gorm:"primaryKey"`
	Title           string   `gorm:"size:255;not null"`
	AuthorID        int32    // Foreign key for Author
	Author          Author   `gorm:"foreignKey:AuthorID"` // Specifies the foreign key relationship
	CategoryID      int32    // Foreign key for Category
	Category        Category `gorm:"foreignKey:CategoryID"` // Specifies the foreign key relationship
	ISBN            string   `gorm:"size:13;unique;not null"`
	PublicationYear int32    `gorm:"not null"`
	Description     string   `gorm:"size:1000"`
}

type Category struct {
	ID          int32  `gorm:"primaryKey"`
	Name        string `gorm:"size:255;not null"`
	Description string `gorm:"size:255"`
}

type Author struct {
	ID   int32  `gorm:"primaryKey"`
	Name string `gorm:"size:255;not null"`
	Bio  string `gorm:"size:500"`
}

type ReturningTransaction struct {
	ID                     int32                `gorm:"primaryKey"`
	BorrowingTransactionID int32                // Foreign key for BorrowingTransaction
	BorrowingTransaction   BorrowingTransaction `gorm:"foreignKey:BorrowingTransactionID"` // Specifies the foreign key relationship
	ReturnedAt             time.Time            `gorm:"not null"`
}

type BookStock struct {
	ID         int `gorm:"primaryKey;autoIncrement" json:"id"`
	BookID     int `gorm:"index" json:"book_id"`
	TotalStock int `gorm:"not null" json:"total_stock"`
}
