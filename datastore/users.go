package datastore

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/treeder/firetils"
	"github.com/treeder/gotils/v2"
	"github.com/treeder/messengers/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func SaveUser(ctx context.Context, client *firestore.Client, user *models.User) error {
	firetils.UpdateTimeStamps(user)
	_, err := client.Collection(CollectionUsers).Doc(user.ID).Set(ctx, user)
	if err != nil {
		return gotils.C(ctx).Errorf("Failed to save user, please try again. %v", err)
	}
	return nil
}

func GetUser(ctx context.Context, client *firestore.Client, id string) (*models.User, error) {
	doc, err := client.Collection(CollectionUsers).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, gotils.ErrNotFound
		}
		return nil, gotils.C(ctx).Errorf("Failed to get user, please try again. %v", err)
	}
	m := &models.User{}
	err = doc.DataTo(m)
	if err != nil {
		return nil, gotils.C(ctx).Errorf("Error getting user, please try again or contact support. %v", err)
	}
	m.Ref = doc.Ref
	m.ID = doc.Ref.ID
	return m, nil
}
