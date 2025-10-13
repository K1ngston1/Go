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

// Middleware для логування
func LoggingMiddlewareDoctors(next http.Handler) http.Handler {
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

// Middleware для простого ключа авторизації
func AuthMiddlewareDoctors(next http.Handler) http.Handler {
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

func DoctorRoutes() {
	http.Handle("/doctors", LoggingMiddlewareDoctors(AuthMiddlewareDoctors(http.HandlerFunc(doctorsHandler))))
	http.Handle("/doctors/", LoggingMiddlewareDoctors(AuthMiddlewareDoctors(http.HandlerFunc(doctorHandler))))
}

func doctorsHandler(w http.ResponseWriter, r *http.Request) {
	col := db.Client.Database("hospitaldb").Collection("doctors")

	switch r.Method {
	case http.MethodGet:
		// --- Фільтрація ---
		filter := bson.M{}
		query := r.URL.Query()

		// Фільтр по імені (частковий, нечутливий до регістру)
		if name := strings.TrimSpace(query.Get("name")); name != "" {
			filter["name"] = bson.M{"$regex": name, "$options": "i"}
		}

		// Фільтр по спеціалізації
		if specialty := strings.TrimSpace(query.Get("specialty")); specialty != "" {
			filter["specialty"] = bson.M{"$regex": specialty, "$options": "i"}
		}

		// Фільтр по департаменту (ObjectID)
		if dep := strings.TrimSpace(query.Get("department")); dep != "" {
			if objID, err := primitive.ObjectIDFromHex(dep); err == nil {
				filter["department"] = objID
			}
		}

		// Фільтр по досвіду (точне або діапазон)
		if expStr := strings.TrimSpace(query.Get("experience_years")); expStr != "" {
			if expInt, err := strconv.Atoi(expStr); err == nil {
				filter["experience_years"] = expInt
			} else {
				filter["experience_years"] = expStr
			}
		} else {
			minExpStr := strings.TrimSpace(query.Get("minExperience"))
			maxExpStr := strings.TrimSpace(query.Get("maxExperience"))
			rangeFilter := bson.M{}

			if minExpStr != "" {
				if minExp, err := strconv.Atoi(minExpStr); err == nil {
					rangeFilter["$gte"] = minExp
				}
			}
			if maxExpStr != "" {
				if maxExp, err := strconv.Atoi(maxExpStr); err == nil {
					rangeFilter["$lte"] = maxExp
				}
			}
			if len(rangeFilter) > 0 {
				filter["experience_years"] = rangeFilter
			}
		}

		cursor, err := col.Find(context.TODO(), filter)
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

		writeJSONDoctor(w, doctors)

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
		writeJSONDoctor(w, doctor)

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

// Допоміжна функція для JSON-відповіді
func writeJSONDoctor(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
