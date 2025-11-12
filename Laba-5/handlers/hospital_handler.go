package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"hospital-api/db"
	"hospital-api/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- Middleware для логування ---
func LoggingMiddlewareHospitals(next http.Handler) http.Handler {
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

// --- Middleware для простої авторизації ---
func AuthMiddlewareHospitals(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const apiKey = "my-secret-key"
		key := r.Header.Get("X-API-KEY")
		if key != apiKey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// --- Реєстрація маршрутів ---
func HospitalRoutes() {
	http.Handle("/hospitals", LoggingMiddlewareHospitals(AuthMiddlewareHospitals(http.HandlerFunc(hospitalsHandler))))
	http.Handle("/hospitals/", LoggingMiddlewareHospitals(AuthMiddlewareHospitals(http.HandlerFunc(hospitalHandler))))
}

func hospitalsHandler(w http.ResponseWriter, r *http.Request) {
	col := db.Client.Database("hospital_db").Collection("hospitals")

	switch r.Method {
	case http.MethodGet:
		// --- Фільтрація ---
		filter := bson.M{}
		query := r.URL.Query()

		// Фільтр за назвою
		if name := strings.TrimSpace(query.Get("name")); name != "" {
			filter["name"] = bson.M{"$regex": name, "$options": "i"}
		}

		// Фільтр за містом
		if city := strings.TrimSpace(query.Get("city")); city != "" {
			filter["city"] = bson.M{"$regex": city, "$options": "i"}
		}

		// Фільтр за кількістю ліжок (точно або діапазон)
		if bedsStr := strings.TrimSpace(query.Get("beds")); bedsStr != "" {
			if bedsInt, err := strconv.Atoi(bedsStr); err == nil {
				filter["beds"] = bedsInt
			}
		} else {
			minBedsStr := query.Get("minBeds")
			maxBedsStr := query.Get("maxBeds")
			rangeFilter := bson.M{}

			if minBedsStr != "" {
				if minBeds, err := strconv.Atoi(minBedsStr); err == nil {
					rangeFilter["$gte"] = minBeds
				}
			}
			if maxBedsStr != "" {
				if maxBeds, err := strconv.Atoi(maxBedsStr); err == nil {
					rangeFilter["$lte"] = maxBeds
				}
			}
			if len(rangeFilter) > 0 {
				filter["beds"] = rangeFilter
			}
		}

		// --- Отримання з бази ---
		cursor, err := col.Find(context.TODO(), filter)
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
		writeJSONHospitals(w, hospitals)

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
		writeJSONHospitals(w, hospital)

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
			http.Error(w, "Hospital not found", http.StatusNotFound)
			return
		}
		writeJSONHospitals(w, hospital)

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

func writeJSONHospitals(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
