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

func DepartmentRoutes() {
	http.HandleFunc("/departments", departmentsHandler)
	http.HandleFunc("/departments/", departmentHandler)
}

func departmentsHandler(w http.ResponseWriter, r *http.Request) {
	col := db.Client.Database("hospitaldb").Collection("departments")

	switch r.Method {
	case http.MethodGet:
		cursor, err := col.Find(context.TODO(), bson.M{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.TODO())

		var departments []models.Department
		if err := cursor.All(context.TODO(), &departments); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, departments)

	case http.MethodPost:
		var department models.Department
		if err := json.NewDecoder(r.Body).Decode(&department); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err := col.InsertOne(context.TODO(), department)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		department.ID = res.InsertedID.(primitive.ObjectID)
		writeJSON(w, department)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func departmentHandler(w http.ResponseWriter, r *http.Request) {
	col := db.Client.Database("hospitaldb").Collection("departments")
	id := strings.TrimPrefix(r.URL.Path, "/departments/")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		var department models.Department
		err := col.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&department)
		if err != nil {
			http.Error(w, "Department not found", http.StatusNotFound)
			return
		}
		writeJSON(w, department)

	case http.MethodPut:
		var update models.Department
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		updateMap := bson.M{
			"name":        update.Name,
			"hospital_id": update.HospitalID,
			"floor":       update.Floor,
		}

		_, err := col.UpdateOne(context.TODO(), bson.M{"_id": objID}, bson.M{"$set": updateMap})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Department updated successfully")

	case http.MethodDelete:
		_, err := col.DeleteOne(context.TODO(), bson.M{"_id": objID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Department deleted successfully")

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
