package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Medicine struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Dosage       string             `bson:"dosage" json:"dosage"`
	Manufacturer string             `bson:"manufacturer" json:"manufacturer"`
	Stock        int                `bson:"stock" json:"stock"`
}
