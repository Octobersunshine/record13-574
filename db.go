package main

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("pt_booking.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	err = db.AutoMigrate(&Trainer{}, &TimeSlot{}, &Booking{})
	if err != nil {
		panic("failed to migrate: " + err.Error())
	}

	return db
}
