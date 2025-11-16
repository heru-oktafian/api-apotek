package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	// GORM akan secara otomatis menambahkan ID (sebagai Primary Key default, tapi kita akan override),
	// CreatedAt, UpdatedAt, DeletedAt.
	// Jika kita ingin id_user sebagai primary key dan varchar, kita perlu override default GORM.

	// Field id_user sebagai Primary Key custom
	UserID string `gorm:"primaryKey;type:varchar(15)" json:"user_id"`

	Username string `gorm:"unique;not null;type:varchar(255)" json:"username" validate:"required"`
	Password string `gorm:"not null;type:text" json:"-" validate:"required"` // `-` agar tidak muncul di JSON response

	Name string `gorm:"type:varchar(255);not null" json:"name" validate:"required"`

	// Menggunakan string biasa atau custom type untuk ENUM di Go
	UserRole   UserRole   `gorm:"type:user_role;not null;default:'operator'" json:"user_role" validate:"required"`
	UserStatus DataStatus `gorm:"type:data_status;not null;default:'inactive'" json:"user_status" validate:"required"`

	// Menggabungkan gorm.Model jika ingin CreatedAt/UpdatedAt, tapi kita custom IDUser
	// Jika Anda ingin CreatedAt/UpdatedAt, bisa ditambahkan secara manual seperti ini:
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // Untuk Soft Delete
}

// SetID is function to set ID into User
func (b *User) SetID(id string) {
	b.UserID = id
}

// HashPassword is a function to hash password
func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}
