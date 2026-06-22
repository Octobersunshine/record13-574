package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter(h *Handler) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api/v1")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})

		api.POST("/trainers", h.CreateTrainer)
		api.GET("/trainers", h.ListTrainers)

		api.POST("/slots", h.CreateSlot)
		api.GET("/slots", h.ListSlots)

		api.POST("/bookings", h.CreateBooking)
		api.GET("/bookings", h.ListBookings)
		api.DELETE("/bookings/:id", h.CancelBooking)
	}

	return r
}
