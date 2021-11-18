package models

type OwnedI interface {
	GetService() string
	GetServiceUserID() string
	SetService(string)
	SetServiceUserID(string)
}
type Owned struct {
	Service string `firestore:"service" json:"-"`
	// ServiceUserID maps to the `service_users` collection
	ServiceUserID string `firestore:"service_user_id" json:"-"`
}

func (o *Owned) GetService() string {
	return o.Service
}

func (o *Owned) SetService(s string) {
	o.Service = s
}

func (o *Owned) GetServiceUserID() string {
	return o.ServiceUserID
}

func (o *Owned) SetServiceUserID(s string) {
	o.ServiceUserID = s
}

func UpdateOwned(obj OwnedI, su *ServiceUser) OwnedI {
	if obj.GetServiceUserID() != "" {
		return obj
	}
	obj.SetService(su.Service)
	obj.SetServiceUserID(su.ID)
	return obj
}
