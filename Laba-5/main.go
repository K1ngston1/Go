package main

import (
	"fmt"
	"log"
	"net/http"

	"hospital-api/db"
	"hospital-api/handlers"
)

func main() {
	db.Connect("mongodb://localhost:27017")

	// Спочатку реєструємо всі CRUD маршрути
	handlers.AppointmentRoutes()
	handlers.StaffRoutes()
	handlers.MedicineRoutes()
	handlers.DoctorRoutes()
	handlers.HospitalRoutes()
	handlers.DepartmentRoutes()

	// Потім catch-all "/" в кінці
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "✅ API працює! Використовуй /hospitals, /appointments, /patients тощо.")
	})

	fmt.Println("🚀 Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
