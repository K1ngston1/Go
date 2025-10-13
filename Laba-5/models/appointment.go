package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Appointment struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PatientID primitive.ObjectID `bson:"patientId" json:"patientId"`
	DoctorID  primitive.ObjectID `bson:"doctorId" json:"doctorId"`
	Date      time.Time          `bson:"date" json:"date"`
}
