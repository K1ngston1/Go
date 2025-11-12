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

// JWT + Logging Middleware додається тут
func StaffRoutes() {
	http.Handle("/staff", LoggingMiddleware(
		JWTAuthMiddleware(http.HandlerFunc(staffHandler), "reader", "admin"),
	))
	http.Handle("/staff/", LoggingMiddleware(
		JWTAuthMiddleware(http.HandlerFunc(staffMemberHandler), "reader", "admin"),
	))
}

func staffHandler(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r) // для перевірки ролі при POST
	col := db.Client.Database("hospital_db").Collection("staff")

	switch r.Method {
	case http.MethodGet:
		filter := bson.M{}
		query := r.URL.Query()
		if name := strings.TrimSpace(query.Get("name")); name != "" {
			filter["name"] = bson.M{"$regex": name, "$options": "i"}
		}
		if role := strings.TrimSpace(query.Get("role")); role != "" {
			filter["role"] = bson.M{"$regex": role, "$options": "i"}
		}
		if shift := strings.TrimSpace(query.Get("shift")); shift != "" {
			filter["shift"] = bson.M{"$regex": shift, "$options": "i"}
		}

		cursor, err := col.Find(context.TODO(), filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.TODO())

		var staff []models.Staff
		if err := cursor.All(context.TODO(), &staff); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, staff)

	case http.MethodPost:
		if claims.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		var staffMember models.Staff
		if err := json.NewDecoder(r.Body).Decode(&staffMember); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err := col.InsertOne(context.TODO(), staffMember)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		staffMember.ID = res.InsertedID.(primitive.ObjectID)
		writeJSON(w, staffMember)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func staffMemberHandler(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	col := db.Client.Database("hospital_db").Collection("staff")
	id := strings.TrimPrefix(r.URL.Path, "/staff/")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		var staffMember models.Staff
		err := col.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&staffMember)
		if err != nil {
			http.Error(w, "Staff member not found", http.StatusNotFound)
			return
		}
		writeJSON(w, staffMember)

	case http.MethodPut:
		if claims.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		var update models.Staff
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		updateMap := bson.M{
			"name":  update.Name,
			"role":  update.Role,
			"shift": update.Shift,
		}

		_, err := col.UpdateOne(context.TODO(), bson.M{"_id": objID}, bson.M{"$set": updateMap})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Staff member updated successfully")

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
		fmt.Fprintf(w, "Staff member deleted successfully")

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Допоміжна функція для JSON
func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
