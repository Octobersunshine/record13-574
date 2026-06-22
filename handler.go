package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) CreateTrainer(c *gin.Context) {
	var input struct {
		Name      string `json:"name" binding:"required"`
		Specialty string `json:"specialty"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trainer := Trainer{
		Name:      input.Name,
		Specialty: input.Specialty,
	}
	if err := h.db.Create(&trainer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, trainer)
}

func (h *Handler) ListTrainers(c *gin.Context) {
	var trainers []Trainer
	if err := h.db.Find(&trainers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, trainers)
}

func (h *Handler) CreateSlot(c *gin.Context) {
	var input struct {
		TrainerID uint      `json:"trainer_id" binding:"required"`
		StartTime time.Time `json:"start_time" binding:"required"`
		EndTime   time.Time `json:"end_time" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !input.EndTime.After(input.StartTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_time must be after start_time"})
		return
	}

	var trainer Trainer
	if err := h.db.First(&trainer, input.TrainerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "trainer not found"})
		return
	}

	var conflictCount int64
	h.db.Model(&TimeSlot{}).
		Where("trainer_id = ? AND status = ? AND start_time < ? AND end_time > ?",
			input.TrainerID, "available", input.EndTime, input.StartTime).
		Count(&conflictCount)
	if conflictCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "time slot conflicts with existing slots"})
		return
	}

	slot := TimeSlot{
		TrainerID: input.TrainerID,
		StartTime: input.StartTime,
		EndTime:   input.EndTime,
		Status:    "available",
	}
	if err := h.db.Create(&slot).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.db.Preload("Trainer").First(&slot, slot.ID)
	c.JSON(http.StatusCreated, slot)
}

func (h *Handler) ListSlots(c *gin.Context) {
	var slots []TimeSlot

	query := h.db.Preload("Trainer")

	if trainerID := c.Query("trainer_id"); trainerID != "" {
		query = query.Where("trainer_id = ?", trainerID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if date := c.Query("date"); date != "" {
		t, err := time.Parse("2006-01-02", date)
		if err == nil {
			nextDay := t.AddDate(0, 0, 1)
			query = query.Where("start_time >= ? AND start_time < ?", t, nextDay)
		}
	}

	if err := query.Order("start_time ASC").Find(&slots).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, slots)
}

func (h *Handler) CreateBooking(c *gin.Context) {
	var input struct {
		TimeSlotID uint   `json:"time_slot_id" binding:"required"`
		UserName   string `json:"user_name" binding:"required"`
		UserPhone  string `json:"user_phone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var slot TimeSlot
	if err := h.db.Preload("Trainer").First(&slot, input.TimeSlotID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "time slot not found"})
		return
	}

	if slot.Status != "available" {
		c.JSON(http.StatusConflict, gin.H{"error": "time slot is not available"})
		return
	}

	tx := h.db.Begin()

	if err := tx.Model(&slot).Update("status", "booked").Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	booking := Booking{
		TimeSlotID: input.TimeSlotID,
		UserName:   input.UserName,
		UserPhone:  input.UserPhone,
		Status:     "confirmed",
	}
	if err := tx.Create(&booking).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tx.Commit()

	h.db.Preload("TimeSlot.Trainer").First(&booking, booking.ID)
	c.JSON(http.StatusCreated, booking)
}

func (h *Handler) CancelBooking(c *gin.Context) {
	id := c.Param("id")

	var booking Booking
	if err := h.db.Preload("TimeSlot").First(&booking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		return
	}

	if booking.Status != "confirmed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "booking is already cancelled"})
		return
	}

	tx := h.db.Begin()

	if err := tx.Model(&booking).Update("status", "cancelled").Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Model(&TimeSlot{}).Where("id = ?", booking.TimeSlotID).Update("status", "available").Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "booking cancelled, slot released"})
}

func (h *Handler) ListBookings(c *gin.Context) {
	var bookings []Booking

	query := h.db.Preload("TimeSlot.Trainer")

	if userName := c.Query("user_name"); userName != "" {
		query = query.Where("user_name = ?", userName)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&bookings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bookings)
}
