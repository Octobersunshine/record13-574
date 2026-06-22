package main

import "time"

type Trainer struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Specialty string    `json:"specialty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TimeSlot struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TrainerID uint      `gorm:"not null;index" json:"trainer_id"`
	Trainer   Trainer   `gorm:"foreignKey:TrainerID" json:"trainer,omitempty"`
	StartTime time.Time `gorm:"not null" json:"start_time"`
	EndTime   time.Time `gorm:"not null" json:"end_time"`
	Status    string    `gorm:"not null;default:available;size:20" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (TimeSlot) TableName() string {
	return "time_slots"
}

type Booking struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	TimeSlotID uint      `gorm:"not null;uniqueIndex" json:"time_slot_id"`
	TimeSlot   TimeSlot  `gorm:"foreignKey:TimeSlotID" json:"time_slot,omitempty"`
	UserName   string    `gorm:"not null;size:100" json:"user_name"`
	UserPhone  string    `gorm:"not null;size:20" json:"user_phone"`
	Status     string    `gorm:"not null;default:confirmed;size:20" json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
