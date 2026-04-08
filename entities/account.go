package entities

import "time"

type TransferResult struct {
	From Account `json:"from"`
	To   Account `json:"to"`
}

type Account struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"size:255;not null"`
	Surname      string    `json:"surname" gorm:"size:255;not null"`
	Email        string    `json:"email" gorm:"size:255;uniqueIndex;not null"`
	Phone        *string   `json:"phone" gorm:"size:50"`
	AddressLine1 string    `json:"addressLine1" gorm:"column:address_line1;size:255;not null"`
	AddressLine2 *string   `json:"addressLine2" gorm:"column:address_line2;size:255"`
	City         string    `json:"city" gorm:"size:100;not null"`
	Postcode     string    `json:"postcode" gorm:"size:20;not null"`
	Country      string    `json:"country" gorm:"size:100;not null"`
	Balance      int64     `json:"balance" gorm:"default:0;not null"`
	Currency     string    `json:"currency" gorm:"size:3;default:'EUR';not null"`
	CreatedAt    time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt    time.Time `json:"updatedAt" gorm:"column:updated_at"`
}
