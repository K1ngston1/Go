package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"hospital-api/db"
	"hospital-api/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func DoctorRoutes() {
	http.HandleFunc("/doctors", doctorsHandler)
	http.HandleFunc("/doctors/", doctorHandler)
}

func doctorsHandler(w http.ResponseWriter, r *http.Request) {
	col := db.Client.Database("hospitaldb").Collection("doctors")

	switch r.Method {
	case http.MethodGet:
		cursor, err := col.Find(context.TODO(), bson.M{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.TODO())

		var doctors []models.Doctor
		if err := cursor.All(context.TODO(), &doctors); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, doctors)

	case http.MethodPost:
		var doctor models.Doctor
		if err := json.NewDecoder(r.Body).Decode(&doctor); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err := col.InsertOne(context.TODO(), doctor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		doctor.ID = res.InsertedID.(primitive.ObjectID)
		writeJSON(w, doctor)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func doctorHandler(w http.ResponseWriter, r *http.Request) {
	col := db.Client.Database("hospitaldb").Collection("doctors")
	id := strings.TrimPrefix(r.URL.Path, "/doctors/")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		var doctor models.Doctor
		err := col.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&doctor)
		if err != nil {
			http.Error(w, "Doctor not found", http.StatusNotFound)
			return
		}
		writeJSON(w, doctor)

	case http.MethodPut:
		var update models.Doctor
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		updateMap := bson.M{
			"name":             update.Name,
			"specialty":        update.Specialty,
			"department":       update.Department,
			"experience_years": update.ExperienceYears,
		}

		_, err := col.UpdateOne(context.TODO(), bson.M{"_id": objID}, bson.M{"$set": updateMap})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Doctor updated successfully")

	case http.MethodDelete:
		_, err := col.DeleteOne(context.TODO(), bson.M{"_id": objID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Doctor deleted successfully")

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
