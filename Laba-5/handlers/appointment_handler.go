package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"hospital-api/db"
	"hospital-api/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Middleware для логування
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.OpenFile("access.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println("Cannot open log file:", err)
		} else {
			defer f.Close()
			log.SetOutput(f)
			log.Printf("%s %s\n", r.Method, r.URL.Path)
		}
		next.ServeHTTP(w, r)
	})
}

// Реєстрація маршрутів
func AppointmentRoutes() {
	http.Handle("/login", LoggingMiddleware(http.HandlerFunc(LoginHandler)))

	// GET дозволений reader і admin
	http.Handle("/appointments", LoggingMiddleware(
		JWTAuthMiddleware(http.HandlerFunc(appointmentsHandler), "reader", "admin"),
	))
	http.Handle("/appointments/", LoggingMiddleware(
		JWTAuthMiddleware(http.HandlerFunc(appointmentHandler), "reader", "admin"),
	))
}

// Обробник для списку зустрічей
func appointmentsHandler(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	col := db.Client.Database("hospital_db").Collection("appointments")

	switch r.Method {
	case http.MethodGet:
		filter := bson.M{}
		query := r.URL.Query()
		if patient := query.Get("patientId"); patient != "" {
			if objID, err := primitive.ObjectIDFromHex(patient); err == nil {
				filter["patientId"] = objID
			}
		}
		if doctor := query.Get("doctorId"); doctor != "" {
			if objID, err := primitive.ObjectIDFromHex(doctor); err == nil {
				filter["doctorId"] = objID
			}
		}
		if dateStr := query.Get("date"); dateStr != "" {
			if date, err := time.Parse("2006-01-02", dateStr); err == nil {
				start := date
				end := date.Add(24 * time.Hour)
				filter["date"] = bson.M{"$gte": start, "$lt": end}
			}
		}

		cursor, err := col.Find(context.TODO(), filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.TODO())

		var appointments []models.Appointment
		if err := cursor.All(context.TODO(), &appointments); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSONApointemnt(w, appointments)

	case http.MethodPost:
		if claims.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		var appointment models.Appointment
		if err := json.NewDecoder(r.Body).Decode(&appointment); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if appointment.Date.IsZero() {
			appointment.Date = time.Now()
		}

		res, err := col.InsertOne(context.TODO(), appointment)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		appointment.ID = res.InsertedID.(primitive.ObjectID)
		writeJSONApointemnt(w, appointment)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Обробник для конкретної зустрічі
func appointmentHandler(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	col := db.Client.Database("hospital_db").Collection("appointments")
	id := strings.TrimPrefix(r.URL.Path, "/appointments/")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		var appointment models.Appointment
		err := col.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&appointment)
		if err != nil {
			http.Error(w, "Appointment not found", http.StatusNotFound)
			return
		}
		writeJSONApointemnt(w, appointment)

	case http.MethodPut:
		if claims.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		var update models.Appointment
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		updateMap := bson.M{
			"patientId": update.PatientID,
			"doctorId":  update.DoctorID,
			"date":      update.Date,
		}

		_, err := col.UpdateOne(context.TODO(), bson.M{"_id": objID}, bson.M{"$set": updateMap})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Appointment updated successfully")

	case http.MethodDelete:
		if claims.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		_, err := col.DeleteOne(context.TODO(), bson.M{"_id": objID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Appointment deleted successfully")

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func writeJSONApointemnt(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
