package models

import (
	"github.com/treeder/firetils"
)

// User is for the master user, can attach ServiceUser's to this master account
// Once there's a web interface with more functionality
type User struct {
	firetils.Firestored
	firetils.IDed
	firetils.TimeStamped

	// WARNING: before adding new fields here, be sure they can be public, this is returned in the get user endpoint

	// ID the user's top level ID, issued by firebase Auth
	// ID          string `firestore:"id" json:"id"`
	DisplayName string `firestore:"display_name" json:"display_name"`
	Email       string `firestore:"email" json:"email"`
	PhotoURL    string `firestore:"photo_url" json:"photo_url"`
}
