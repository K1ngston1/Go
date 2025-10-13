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
func LoggingMiddlewareDepartments(next http.Handler) http.Handler {
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
func AuthMiddlewareDepartments(next http.Handler) http.Handler {
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

func DepartmentRoutes() {
	http.Handle("/departments", LoggingMiddlewareDepartments(AuthMiddlewareDepartments(http.HandlerFunc(departmentsHandler))))
	http.Handle("/departments/", LoggingMiddlewareDepartments(AuthMiddlewareDepartments(http.HandlerFunc(departmentHandler))))
}

func departmentsHandler(w http.ResponseWriter, r *http.Request) {
	col := db.Client.Database("hospitaldb").Collection("departments")

	switch r.Method {
	case http.MethodGet:
		// Фільтр через query params
		filter := bson.M{}
		query := r.URL.Query()

		// Пошук по імені (частковий, нечутливий до регістру)
		if name := strings.TrimSpace(query.Get("name")); name != "" {
			filter["name"] = bson.M{"$regex": name, "$options": "i"}
		}

		// Пошук по hospitalId
		if hospital := strings.TrimSpace(query.Get("hospitalId")); hospital != "" {
			if objID, err := primitive.ObjectIDFromHex(hospital); err == nil {
				filter["hospital_id"] = objID
			}
		}

		// Пошук по floor (точне значення або діапазон)
		if floorStr := strings.TrimSpace(query.Get("floor")); floorStr != "" {
			if floorInt, err := strconv.Atoi(floorStr); err == nil {
				filter["floor"] = floorInt
			} else {
				filter["floor"] = floorStr
			}
		} else {
			// Підтримка minFloor / maxFloor
			minFloorStr := strings.TrimSpace(query.Get("minFloor"))
			maxFloorStr := strings.TrimSpace(query.Get("maxFloor"))
			rangeFilter := bson.M{}

			if minFloorStr != "" {
				if minFloor, err := strconv.Atoi(minFloorStr); err == nil {
					rangeFilter["$gte"] = minFloor
				}
			}
			if maxFloorStr != "" {
				if maxFloor, err := strconv.Atoi(maxFloorStr); err == nil {
					rangeFilter["$lte"] = maxFloor
				}
			}
			if len(rangeFilter) > 0 {
				filter["floor"] = rangeFilter
			}
		}

		// Виконуємо запит до MongoDB
		cursor, err := col.Find(context.TODO(), filter)
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

		writeJSONDepartments(w, departments)

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
		writeJSONDepartments(w, department)

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
		writeJSONDepartments(w, department)

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

// Допоміжна функція для JSON відповіді
func writeJSONDepartments(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
