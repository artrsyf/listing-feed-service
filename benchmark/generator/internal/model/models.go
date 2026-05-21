package model

import "time"

// =======================
// USERS
// =======================

type User struct {
	ID        int64
	Email     string
	Country   string
	Status    int
	CreatedAt time.Time
}

// =======================
// CATEGORIES
// =======================

type Category struct {
	ID        int64
	ParentID  *int64
	Name      string
	CreatedAt time.Time
}

// =======================
// PRODUCTS
// =======================

type Product struct {
	ID         int64
	CategoryID int64
	SKU        string
	Name       string
	Price      float64
	Rating     float64
	CreatedAt  time.Time
}

// =======================
// ORDERS
// =======================

type Order struct {
	ID          int64
	UserID      int64
	Status      int
	TotalAmount float64
	CreatedAt   time.Time
}

// =======================
// ORDER ITEMS
// =======================

type OrderItem struct {
	ID        int64
	OrderID   int64
	ProductID int64
	Quantity  int
	Price     float64
	CreatedAt time.Time
}
