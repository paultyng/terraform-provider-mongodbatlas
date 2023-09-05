package mock

import (
	"context"

	matlas "go.mongodb.org/atlas/mongodbatlas"
)

type DatabaseUsersServiceOp struct{}

func (s *DatabaseUsersServiceOp) Get(ctx context.Context, databaseName, groupID, username string) (*matlas.DatabaseUser, *matlas.Response, error) {
	return &matlas.DatabaseUser{
		GroupID:      groupID,
		Username:     username,
		DatabaseName: databaseName,
		LDAPAuthType: "NONE",
		AWSIAMType:   "NONE",
		X509Type:     "NONE",
		OIDCAuthType: "NONE",
		Roles: []matlas.Role{
			{
				RoleName:     "atlasAdmin",
				DatabaseName: "admin",
			},
		},
	}, nil, nil
}

func (s *DatabaseUsersServiceOp) List(ctx context.Context, groupID string, listOptions *matlas.ListOptions) ([]matlas.DatabaseUser, *matlas.Response, error) {
	return nil, nil, nil
}

func (s *DatabaseUsersServiceOp) Create(ctx context.Context, groupID string, createRequest *matlas.DatabaseUser) (*matlas.DatabaseUser, *matlas.Response, error) {
	return &matlas.DatabaseUser{
		GroupID:      groupID,
		Username:     createRequest.Username,
		DatabaseName: createRequest.DatabaseName,
		LDAPAuthType: "NONE",
		AWSIAMType:   "NONE",
		X509Type:     "NONE",
		OIDCAuthType: "NONE",
		Roles: []matlas.Role{
			{
				RoleName:     "atlasAdmin",
				DatabaseName: "admin",
			},
		},
	}, nil, nil
}

func (s *DatabaseUsersServiceOp) Update(ctx context.Context, groupID, username string, updateRequest *matlas.DatabaseUser) (*matlas.DatabaseUser, *matlas.Response, error) {
	return &matlas.DatabaseUser{
		GroupID:      groupID,
		Username:     updateRequest.Username,
		DatabaseName: updateRequest.DatabaseName,
		LDAPAuthType: "NONE",
		AWSIAMType:   "NONE",
		X509Type:     "NONE",
		OIDCAuthType: "NONE",
		Roles: []matlas.Role{
			{
				RoleName:     "atlasAdmin",
				DatabaseName: "admin",
			},
		},
	}, nil, nil
}

func (s *DatabaseUsersServiceOp) Delete(ctx context.Context, databaseName, groupID, username string) (*matlas.Response, error) {
	return nil, nil
}
