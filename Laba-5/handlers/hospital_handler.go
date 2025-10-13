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

func HospitalRoutes() {
	http.HandleFunc("/hospitals", hospitalsHandler)
	http.HandleFunc("/hospitals/", hospitalHandler)
}

func hospitalsHandler(w http.ResponseWriter, r *http.Request) {
	col := db.Client.Database("hospitaldb").Collection("hospitals")

	switch r.Method {
	case http.MethodGet:
		cursor, err := col.Find(context.TODO(), bson.M{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.TODO())

		var hospitals []models.Hospital
		if err := cursor.All(context.TODO(), &hospitals); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, hospitals)

	case http.MethodPost:
		var hospital models.Hospital
		if err := json.NewDecoder(r.Body).Decode(&hospital); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err := col.InsertOne(context.TODO(), hospital)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		hospital.ID = res.InsertedID.(primitive.ObjectID)
		writeJSON(w, hospital)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func hospitalHandler(w http.ResponseWriter, r *http.Request) {
	col := db.Client.Database("hospitaldb").Collection("hospitals")
	id := strings.TrimPrefix(r.URL.Path, "/hospitals/")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		var hospital models.Hospital
		err := col.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&hospital)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		writeJSON(w, hospital)

	case http.MethodPut:
		var update models.Hospital
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, err := col.UpdateOne(context.TODO(), bson.M{"_id": objID}, bson.M{"$set": update})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Hospital updated successfully")

	case http.MethodDelete:
		_, err := col.DeleteOne(context.TODO(), bson.M{"_id": objID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Hospital deleted successfully")
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
