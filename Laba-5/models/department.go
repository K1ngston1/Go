package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Department struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name       string             `bson:"name" json:"name"`
	HospitalID primitive.ObjectID `bson:"hospital_id" json:"hospitalId"`
	Floor      int                `bson:"floor" json:"floor"`
}
