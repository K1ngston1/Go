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

func MedicineRoutes() {
	http.HandleFunc("/medicines", medicinesHandler)
	http.HandleFunc("/medicines/", medicineHandler)
}

func medicinesHandler(w http.ResponseWriter, r *http.Request) {
	col := db.Client.Database("hospitaldb").Collection("medicines")

	switch r.Method {
	case http.MethodGet:
		cursor, err := col.Find(context.TODO(), bson.M{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.TODO())

		var medicines []models.Medicine
		if err := cursor.All(context.TODO(), &medicines); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSONHospital(w, medicines)

	case http.MethodPost:
		var medicine models.Medicine
		if err := json.NewDecoder(r.Body).Decode(&medicine); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err := col.InsertOne(context.TODO(), medicine)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		medicine.ID = res.InsertedID.(primitive.ObjectID)
		writeJSONHospital(w, medicine)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func medicineHandler(w http.ResponseWriter, r *http.Request) {
	col := db.Client.Database("hospitaldb").Collection("medicines")
	id := strings.TrimPrefix(r.URL.Path, "/medicines/")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		var medicine models.Medicine
		err := col.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&medicine)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		writeJSONHospital(w, medicine)

	case http.MethodPut:
		var update models.Medicine
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, err := col.UpdateOne(context.TODO(), bson.M{"_id": objID}, bson.M{"$set": update})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Medicine updated successfully")

	case http.MethodDelete:
		_, err := col.DeleteOne(context.TODO(), bson.M{"_id": objID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Medicine deleted successfully")

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func writeJSONHospital(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
