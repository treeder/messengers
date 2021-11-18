package datastore

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/treeder/firetils"
	"github.com/treeder/gotils/v2"
	"github.com/treeder/messengers/models"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func ServiceUserID(ctx context.Context, serviceName string, userID string) string {
	if serviceName == "" {
		panic("ServiceUserID() needs a service name")
	}
	return fmt.Sprintf("%v-%v", serviceName, userID)
}

func SaveServiceUser(ctx context.Context, client *firestore.Client, su *models.ServiceUser) error {
	firetils.UpdateTimeStamps(su)
	_, err := client.Collection(CollectionServiceUsers).Doc(ServiceUserID(ctx, su.Service, su.ID)).Set(ctx, su)
	if err != nil {
		return gotils.C(ctx).Errorf("Failed to save user, please try again: %v", err)
	}
	return nil
}

func GetServiceUserByID(ctx context.Context, client *firestore.Client, service, id string) (*models.ServiceUser, error) {
	dsnap, err := client.Collection(CollectionServiceUsers).Doc(ServiceUserID(ctx, service, id)).Get(ctx)
	if err != nil {
		if grpc.Code(err) != codes.NotFound {
			return nil, gotils.C(ctx).Errorf("Failed to get service account, please try again. %v", err)
		}
		return nil, gotils.ErrNotFound
	}
	// found
	su := &models.ServiceUser{}
	err = dsnap.DataTo(su)
	if err != nil {
		return nil, gotils.C(ctx).Errorf("Failed to create user object: %v", err)
	}
	return su, nil
}

func GetServiceUserByName(ctx context.Context, client *firestore.Client, service, username string) (*models.ServiceUser, error) {
	uNoAt := username
	if strings.HasPrefix(username, "@") {
		uNoAt = username[1:]
	}
	if len(uNoAt) == 0 {
		return nil, fmt.Errorf("empty username")
	}
	fmt.Printf("QUERYING %v %v\n", service, username)
	uLower := strings.ToLower(uNoAt)
	iter := client.Collection(CollectionServiceUsers).Where("service", "==", service).Where("username_lower", "==", uLower).Documents(ctx)
	doc, err := iter.Next()
	if err == iterator.Done {
		// try second search, before we were lowercasing
		iter = client.Collection(CollectionServiceUsers).Where("service", "==", service).Where("username", "==", uNoAt).Documents(ctx)
		doc, err = iter.Next()
		if err == iterator.Done {
			return nil, gotils.ErrNotFound
		}
		if err != nil {
			return nil, gotils.C(ctx).Errorf("Error searching for user...??? %v", err)
		}
	} else if err != nil {
		return nil, gotils.C(ctx).Errorf("Error searching for user...??? %v", err)
	}
	toUser := &models.ServiceUser{}
	err = doc.DataTo(toUser)
	if err != nil {
		return nil, gotils.C(ctx).Errorf("We had an issue retrieving the user, please try again later: %v", err)
	}
	return toUser, nil
}
