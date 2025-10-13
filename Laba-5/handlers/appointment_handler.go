package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"hospital-api/db"
	"hospital-api/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AppointmentRoutes() {
	http.HandleFunc("/appointments", appointmentsHandler)
	http.HandleFunc("/appointments/", appointmentHandler)
}

func appointmentsHandler(w http.ResponseWriter, r *http.Request) {
	col := db.Client.Database("hospitaldb").Collection("appointments")

	switch r.Method {
	case http.MethodGet:
		cursor, err := col.Find(context.TODO(), bson.M{})
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
		writeJSON(w, appointments)

	case http.MethodPost:
		var appointment models.Appointment
		if err := json.NewDecoder(r.Body).Decode(&appointment); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Якщо дата не передана, ставимо поточну
		if appointment.Date.IsZero() {
			appointment.Date = time.Now()
		}

		res, err := col.InsertOne(context.TODO(), appointment)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		appointment.ID = res.InsertedID.(primitive.ObjectID)
		writeJSON(w, appointment)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func appointmentHandler(w http.ResponseWriter, r *http.Request) {
	col := db.Client.Database("hospitaldb").Collection("appointments")
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
		writeJSON(w, appointment)

	case http.MethodPut:
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
