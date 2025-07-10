package models

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null;uniqueIndex"`
	Description string         `json:"description"`
	Slug        string         `json:"slug" gorm:"uniqueIndex;not null"`
	Active      bool           `json:"active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	Products    []Product      `json:"products,omitempty" gorm:"foreignKey:CategoryID"`
}

type Product struct {
	ID               uint           `json:"id" gorm:"primaryKey"`
	Index            int            `json:"index" gorm:"index"`
	Name             string         `json:"name" gorm:"not null;index"`
	Description      string         `json:"description"`
	ShortDescription string         `json:"short_description"`
	Brand            string         `json:"brand" gorm:"index"`
	Category         string         `json:"category" gorm:"index"`
	Price            float64        `json:"price" gorm:"not null;index"`
	Currency         string         `json:"currency" gorm:"default:'USD'"`
	Stock            int            `json:"stock" gorm:"not null;default:0"`
	EAN              string         `json:"ean" gorm:"index"`
	Color            string         `json:"color"`
	Size             string         `json:"size"`
	Availability     string         `json:"availability" gorm:"index"`
	Image            string         `json:"image"`
	InternalID       string         `json:"internal_id" gorm:"index"`
	Slug             string         `json:"slug" gorm:"not null;index"`
	SKU              string         `json:"sku" gorm:"not null;index"`
	CategoryID       uint           `json:"category_id" gorm:"index"`
	CategoryModel    Category       `json:"category_model,omitempty" gorm:"foreignKey:CategoryID"`
	Active           bool           `json:"active" gorm:"default:true"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
}
