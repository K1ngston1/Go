package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Doctor struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name            string             `bson:"name" json:"name"`
	Specialty       string             `bson:"specialty" json:"specialty"`
	Department      string             `bson:"department" json:"department"`
	ExperienceYears int                `bson:"experience_years" json:"experienceYears"`
}
