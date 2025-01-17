package streamprivatelinkendpoint

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/retrystrategy"
	"go.mongodb.org/atlas-sdk/v20241113004/admin"
)

const (
	defaultTimeout    = 20 * time.Minute // The amount of time to wait before timeout
	defaultMinTimeout = 30 * time.Second // Smallest time to wait before refreshes
)

func waitStateTransition(ctx context.Context, projectID, endpointID string, client admin.StreamsApi) (*admin.StreamsPrivateLinkConnection, error) {
	return WaitStateTransitionWithMinTimeout(ctx, defaultMinTimeout, projectID, endpointID, client)
}

func WaitStateTransitionWithMinTimeout(ctx context.Context, minTimeout time.Duration, projectID, endpointID string, client admin.StreamsApi) (*admin.StreamsPrivateLinkConnection, error) {
	return waitStateTransitionForStates(
		ctx,
		[]string{retrystrategy.RetryStrategyIdleState, retrystrategy.RetryStrategyWorkingState},
		[]string{retrystrategy.RetryStrategyDoneState, retrystrategy.RetryStrategyFailedState},
		minTimeout, projectID, endpointID, client)
}

func WaitDeleteStateTransition(ctx context.Context, projectID, endpointID string, client admin.StreamsApi) (*admin.StreamsPrivateLinkConnection, error) {
	return WaitDeleteStateTransitionWithMinTimeout(ctx, defaultMinTimeout, projectID, endpointID, client)
}

func WaitDeleteStateTransitionWithMinTimeout(ctx context.Context, minTimeout time.Duration, projectID, connectionID string, client admin.StreamsApi) (*admin.StreamsPrivateLinkConnection, error) {
	return waitStateTransitionForStates(
		ctx,
		[]string{retrystrategy.RetryStrategyDeleteRequestedState, retrystrategy.RetryStrategyDeletingState},
		[]string{retrystrategy.RetryStrategyDeletedState, retrystrategy.RetryStrategyFailedState},
		minTimeout, projectID, connectionID, client)
}

func waitStateTransitionForStates(ctx context.Context, pending, target []string, minTimeout time.Duration, projectID, connectionID string, client admin.StreamsApi) (*admin.StreamsPrivateLinkConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    refreshFunc(ctx, projectID, connectionID, client),
		Timeout:    defaultTimeout,
		MinTimeout: minTimeout,
		Delay:      0,
	}

	result, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return nil, err
	}
	if privateEndpointResp, ok := result.(*admin.StreamsPrivateLinkConnection); ok {
		return privateEndpointResp, nil
	}
	return nil, errors.New("did not obtain valid result when waiting for state transition")
}

func refreshFunc(ctx context.Context, projectID, connectionID string, client admin.StreamsApi) retry.StateRefreshFunc {
	return func() (any, string, error) {
		model, resp, err := client.GetPrivateLinkConnection(ctx, projectID, connectionID).Execute()
		if err != nil && model == nil && resp == nil {
			return nil, "", err
		}
		if err != nil {
			if resp.StatusCode == http.StatusNotFound {
				return &admin.StreamsPrivateLinkConnection{}, retrystrategy.RetryStrategyDeletedState, nil
			}
			return nil, "", err
		}
		status := model.GetState()
		return model, status, nil
	}
}
