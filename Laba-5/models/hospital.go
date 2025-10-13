package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Hospital struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name" bson:"name"`
	Location string             `json:"location" bson:"location"`
	Beds     int                `json:"beds" bson:"beds"`
}
