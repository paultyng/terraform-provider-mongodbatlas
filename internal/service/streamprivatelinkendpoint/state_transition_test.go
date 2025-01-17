package streamprivatelinkendpoint_test

import (
	"context"
	"time"

	"errors"
	"net/http"
	"testing"

	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/retrystrategy"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/streamprivatelinkendpoint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/atlas-sdk/v20241113004/admin"
	"go.mongodb.org/atlas-sdk/v20241113004/mockadmin"
)

type testCase struct {
	expectedState *string
	mockResponses []response
	expectedError bool
}

type response struct {
	state      *string
	statusCode *int
	err        error
}

func TestDeleteStateTransition(t *testing.T) {
	testCases := map[string]testCase{
		"Successful transitioning from DELELTING to deleted": {
			mockResponses: []response{
				{state: conversion.StringPtr(retrystrategy.RetryStrategyDeletingState)},
				{statusCode: admin.PtrInt(http.StatusNotFound), err: errors.New("does not exist")},
			},
			expectedError: false,
		},
		"Return model without error when transitioning to FAILED state": {
			mockResponses: []response{
				{state: conversion.StringPtr(retrystrategy.RetryStrategyDeletingState)},
				{state: conversion.StringPtr(retrystrategy.RetryStrategyFailedState)},
			},
			expectedError: false,
			expectedState: conversion.StringPtr(retrystrategy.RetryStrategyFailedState),
		},
		"Error when API responds with error": {
			mockResponses: []response{
				{statusCode: admin.PtrInt(http.StatusInternalServerError), err: errors.New("Internal server error")},
			},
			expectedState: nil,
			expectedError: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			m := mockadmin.NewStreamsApi(t)
			m.EXPECT().GetPrivateLinkConnection(mock.Anything, mock.Anything, mock.Anything).Return(admin.GetPrivateLinkConnectionApiRequest{ApiService: m})

			for _, resp := range tc.mockResponses {
				modelResp, httpResp, err := resp.get()
				m.EXPECT().GetPrivateLinkConnectionExecute(mock.Anything).Return(modelResp, httpResp, err).Once()
			}
			resp, err := streamprivatelinkendpoint.WaitDeleteStateTransitionWithMinTimeout(context.Background(), 1*time.Second, "project-id", "connection-id", m)
			assert.Equal(t, tc.expectedError, err != nil)
			if resp != nil {
				assert.Equal(t, tc.expectedState, resp.State)
			}
		})
	}
}

func (r *response) get() (*admin.StreamsPrivateLinkConnection, *http.Response, error) {
	var httpResp *http.Response
	if r.statusCode != nil {
		httpResp = &http.Response{StatusCode: *r.statusCode}
	}
	return &admin.StreamsPrivateLinkConnection{State: r.state}, httpResp, r.err
}
