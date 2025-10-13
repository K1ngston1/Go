package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Staff struct {
	ID    primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name  string             `bson:"name" json:"name"`
	Role  string             `bson:"role" json:"role"`
	Shift string             `bson:"shift" json:"shift"`
}
