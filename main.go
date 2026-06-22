package main

import "log"

func main() {
	db := InitDB()
	handler := NewHandler(db)
	r := SetupRouter(handler)

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Server failed: ", err)
	}
}
