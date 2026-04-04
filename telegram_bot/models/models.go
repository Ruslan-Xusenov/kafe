package models

import "time"

type UserRole string

const (
	RoleCustomer UserRole = "customer"
	RoleCook     UserRole = "cook"
	RoleCourier  UserRole = "courier"
	RoleAdmin    UserRole = "admin"
)

type User struct {
	ID           int       `json:"id" db:"id"`
	FullName     string    `json:"full_name" db:"full_name"`
	Phone        string    `json:"phone" db:"phone"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         UserRole  `json:"role" db:"role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type Category struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	ImageURL  *string   `json:"image_url" db:"image_url"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Product struct {
	ID          int       `json:"id" db:"id"`
	CategoryID  int       `json:"category_id" db:"category_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Price       float64   `json:"price" db:"price"`
	ImageURL    string    `json:"image_url" db:"image_url"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type OrderStatus string

const (
	StatusNew       OrderStatus = "new"
	StatusPreparing OrderStatus = "preparing"
	StatusReady     OrderStatus = "ready"
	StatusOnWay     OrderStatus = "on_way"
	StatusDelivered OrderStatus = "delivered"
	StatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID          int         `json:"id" db:"id"`
	CustomerID  *int        `json:"customer_id" db:"customer_id"`
	TotalPrice  float64     `json:"total_price" db:"total_price"`
	Status      OrderStatus `json:"status" db:"status"`
	Address     string      `json:"address" db:"address"`
	Phone       string      `json:"phone" db:"phone"`
	Lat         *float64    `json:"lat" db:"lat"`
	Lng         *float64    `json:"lng" db:"lng"`
	CourierID   *int        `json:"courier_id" db:"courier_id"`
	CookID      *int        `json:"cook_id" db:"cook_id"`
	CourierName *string     `json:"courier_name" db:"courier_name"`
	CookName    *string     `json:"cook_name" db:"cook_name"`
	Comment     *string     `json:"comment" db:"comment"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   *time.Time  `json:"updated_at" db:"updated_at"`
	Items       []OrderItem `json:"items,omitempty" db:"-"`
}

type OrderItem struct {
	ID          int       `json:"id" db:"id"`
	OrderID     int       `json:"order_id" db:"order_id"`
	ProductID   int       `json:"product_id" db:"product_id"`
	ProductName *string   `json:"product_name" db:"product_name"`
	Quantity    int       `json:"quantity" db:"quantity"`
	Price       float64   `json:"price" db:"price"`
	Comment     *string   `json:"comment" db:"comment"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type StaffRating struct {
	ID        int       `json:"id" db:"id"`
	OrderID   int       `json:"order_id" db:"order_id"`
	StaffID   int       `json:"staff_id" db:"staff_id"`
	StaffRole string    `json:"staff_role" db:"staff_role"`
	Rating    int       `json:"rating" db:"rating"`
	Comment   string    `json:"comment" db:"comment"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
