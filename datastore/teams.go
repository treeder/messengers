/*
 Teams are things like Slack teams, or gsuite teams or Microsoft Teams
*/

package datastore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/treeder/firetils"
	"github.com/treeder/gotils/v2"
	"github.com/treeder/messengers/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var (
	CollectionUsers        = "users"
	CollectionServiceUsers = "service_users"
	CollectionTeams        = "teams"
)

func TeamID(t *models.Team) string {
	if t.Service == "" {
		panic("Team.Service required")
	}
	return fmt.Sprintf("%v-%v", t.Service, t.ServiceID)
}

func TeamRef(client *firestore.Client, serviceName, teamID string) *firestore.DocumentRef {
	if serviceName == "" {
		panic("Team.Service required")
	}
	return client.Collection(CollectionTeams).Doc(fmt.Sprintf("%v-%v", serviceName, teamID))
}

func SaveTeam(ctx context.Context, client *firestore.Client, ob *models.Team) (*models.Team, error) {
	firetils.UpdateTimeStamps(ob)
	if ob.Service == "" {
		return nil, gotils.C(ctx).Errorf("Team must have Service field set")
	}
	ref := TeamRef(client, ob.Service, ob.ServiceID)
	_, err := ref.Set(ctx, ob)
	if err != nil {
		return nil, gotils.C(ctx).Errorf("Failed to save team, please try again. %v", err)
	}
	ob.Ref = ref
	return ob, nil
}

func GetTeam(ctx context.Context, client *firestore.Client, serviceName, teamID string) (*models.Team, error) {
	ref := TeamRef(client, serviceName, teamID)
	dsnap, err := ref.Get(ctx)
	if err != nil {
		if grpc.Code(err) == codes.NotFound {
			return nil, gotils.ErrNotFound
		}
		return nil, gotils.C(ctx).Errorf("We had an issue retrieving the Team... \U0001f626 please try again later. %v", err)
	}
	team := &models.Team{}
	err = dsnap.DataTo(team)
	if err != nil {
		return nil, gotils.C(ctx).Errorf("We had an issue retrieving the Team... \U0001f626 please try again later", err)
	}
	team.Ref = ref
	return team, nil
}

// func GetTeamWhere(ctx context.Context, serviceName string, query firestore.Query) (*models.Team, error) {
// 	iter := query.Documents(ctx)
// 	doc, err := iter.Next()
// 	if err == iterator.Done {
// 		return nil, gotils.ErrNotFound
// 	}
// 	if err != nil {
// 		gotils.LogBeta(ctx, "error", "error searching for team", err)
// 		// that's a frown emoji
// 		return nil, gotils.C(ctx).Errorf("We had an issue retrieving the team... \U0001f626 please try again later.")
// 	}
// 	// fmt.Println("FOUND", doc.Data())
// 	ico := &models.Team{}
// 	err = doc.DataTo(ico)
// 	if err != nil {
// 		gotils.LogBeta(ctx, "error", "error on DataTo Team", err)
// 		return nil, gotils.C(ctx).Errorf("We had an issue retrieving the team... \U0001f626 please try again later")
// 	}
// 	ico.Ref = doc.Ref
// 	return ico, nil
// }
