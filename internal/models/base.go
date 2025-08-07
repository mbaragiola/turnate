package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UUIDv7 uuid.UUID

func (u *UUIDv7) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			return err
		}
		*u = UUIDv7(parsed)
	case []byte:
		parsed, err := uuid.ParseBytes(v)
		if err != nil {
			return err
		}
		*u = UUIDv7(parsed)
	case nil:
		*u = UUIDv7(uuid.Nil)
	default:
		return fmt.Errorf("cannot scan %T into UUIDv7", value)
	}
	return nil
}

func (u UUIDv7) Value() (driver.Value, error) {
	return uuid.UUID(u).String(), nil
}

func (u UUIDv7) String() string {
	return uuid.UUID(u).String()
}

func NewUUIDv7() UUIDv7 {
	return UUIDv7(uuid.Must(uuid.NewV7()))
}

type BaseModel struct {
	ID        UUIDv7         `json:"id" gorm:"type:text;primaryKey;default:(uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == UUIDv7(uuid.Nil) {
		b.ID = NewUUIDv7()
	}
	return nil
}