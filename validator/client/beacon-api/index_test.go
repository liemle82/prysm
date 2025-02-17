//go:build use_beacon_api
// +build use_beacon_api

package beacon_api

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	rpcmiddleware "github.com/prysmaticlabs/prysm/v3/beacon-chain/rpc/apimiddleware"
	types "github.com/prysmaticlabs/prysm/v3/consensus-types/primitives"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v3/testing/assert"
	"github.com/prysmaticlabs/prysm/v3/testing/require"
	"github.com/prysmaticlabs/prysm/v3/validator/client/beacon-api/mock"
)

const stringPubKey = "0x8000091c2ae64ee414a54c1cc1fc67dec663408bc636cb86756e0200e41a75c8f86603f104f02c856983d2783116be13"

func getPubKeyAndURL(t *testing.T) ([]byte, string) {
	baseUrl := "/eth/v1/beacon/states/head/validators"
	url := fmt.Sprintf("%s?id=%s", baseUrl, stringPubKey)

	pubKey, err := hexutil.Decode(stringPubKey)
	require.NoError(t, err)

	return pubKey, url
}

func TestIndex_Nominal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pubKey, url := getPubKeyAndURL(t)

	stateValidatorsResponseJson := rpcmiddleware.StateValidatorsResponseJson{}
	jsonRestHandler := mock.NewMockjsonRestHandler(ctrl)

	jsonRestHandler.EXPECT().GetRestJsonResponse(
		url,
		&stateValidatorsResponseJson,
	).Return(
		nil,
		nil,
	).SetArg(
		1,
		rpcmiddleware.StateValidatorsResponseJson{
			Data: []*rpcmiddleware.ValidatorContainerJson{
				{
					Index:  "55293",
					Status: "active_ongoing",
					Validator: &rpcmiddleware.ValidatorJson{
						PublicKey: stringPubKey,
					},
				},
			},
		},
	).Times(1)

	validatorClient := beaconApiValidatorClient{jsonRestHandler: jsonRestHandler}

	validatorIndex, err := validatorClient.ValidatorIndex(
		context.Background(),
		&ethpb.ValidatorIndexRequest{
			PublicKey: pubKey,
		},
	)

	require.NoError(t, err)
	assert.Equal(t, types.ValidatorIndex(55293), validatorIndex.Index)
}

func TestIndex_UnexistingValidator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pubKey, url := getPubKeyAndURL(t)

	stateValidatorsResponseJson := rpcmiddleware.StateValidatorsResponseJson{}
	jsonRestHandler := mock.NewMockjsonRestHandler(ctrl)

	jsonRestHandler.EXPECT().GetRestJsonResponse(
		url,
		&stateValidatorsResponseJson,
	).Return(
		nil,
		nil,
	).SetArg(
		1,
		rpcmiddleware.StateValidatorsResponseJson{
			Data: []*rpcmiddleware.ValidatorContainerJson{},
		},
	).Times(1)

	validatorClient := beaconApiValidatorClient{jsonRestHandler: jsonRestHandler}

	_, err := validatorClient.ValidatorIndex(
		context.Background(),
		&ethpb.ValidatorIndexRequest{
			PublicKey: pubKey,
		},
	)

	wanted := "could not find validator index for public key `0x8000091c2ae64ee414a54c1cc1fc67dec663408bc636cb86756e0200e41a75c8f86603f104f02c856983d2783116be13`"
	assert.ErrorContains(t, wanted, err)
}

func TestIndex_BadIndexError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pubKey, url := getPubKeyAndURL(t)

	stateValidatorsResponseJson := rpcmiddleware.StateValidatorsResponseJson{}
	jsonRestHandler := mock.NewMockjsonRestHandler(ctrl)

	jsonRestHandler.EXPECT().GetRestJsonResponse(
		url,
		&stateValidatorsResponseJson,
	).Return(
		nil,
		nil,
	).SetArg(
		1,
		rpcmiddleware.StateValidatorsResponseJson{
			Data: []*rpcmiddleware.ValidatorContainerJson{
				{
					Index:  "This is not an index",
					Status: "active_ongoing",
					Validator: &rpcmiddleware.ValidatorJson{
						PublicKey: stringPubKey,
					},
				},
			},
		},
	).Times(1)

	validatorClient := beaconApiValidatorClient{jsonRestHandler: jsonRestHandler}

	_, err := validatorClient.ValidatorIndex(
		context.Background(),
		&ethpb.ValidatorIndexRequest{
			PublicKey: pubKey,
		},
	)

	assert.ErrorContains(t, "failed to parse validator index", err)
}

func TestIndex_JsonResponseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pubKey, url := getPubKeyAndURL(t)

	stateValidatorsResponseJson := rpcmiddleware.StateValidatorsResponseJson{}
	jsonRestHandler := mock.NewMockjsonRestHandler(ctrl)

	jsonRestHandler.EXPECT().GetRestJsonResponse(
		url,
		&stateValidatorsResponseJson,
	).Return(
		nil,
		errors.New("some specific json error"),
	).Times(1)

	validatorClient := beaconApiValidatorClient{jsonRestHandler: jsonRestHandler}

	_, err := validatorClient.ValidatorIndex(
		context.Background(),
		&ethpb.ValidatorIndexRequest{
			PublicKey: pubKey,
		},
	)

	assert.ErrorContains(t, "failed to get validator state", err)
}
