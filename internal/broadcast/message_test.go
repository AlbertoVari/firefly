// Copyright © 2021 Kaleido, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package broadcast

import (
	"context"
	"fmt"
	"testing"

	"github.com/hyperledger-labs/firefly/mocks/blockchainmocks"
	"github.com/hyperledger-labs/firefly/mocks/databasemocks"
	"github.com/hyperledger-labs/firefly/mocks/datamocks"
	"github.com/hyperledger-labs/firefly/pkg/fftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBroadcastMessageOk(t *testing.T) {
	bm, cancel := newTestBroadcast(t)
	defer cancel()
	mdi := bm.database.(*databasemocks.Plugin)
	mdm := bm.data.(*datamocks.Manager)
	mbi := bm.blockchain.(*blockchainmocks.Plugin)

	ctx := context.Background()
	rag := mdi.On("RunAsGroup", ctx, mock.Anything)
	rag.RunFn = func(a mock.Arguments) {
		var fn = a[1].(func(context.Context) error)
		rag.ReturnArguments = mock.Arguments{fn(a[0].(context.Context))}
	}
	mbi.On("VerifyIdentitySyntax", ctx, "0x12345").Return("0x12345", nil)
	mdm.On("ResolveInputData", ctx, "ns1", mock.Anything).Return(fftypes.DataRefs{
		{ID: fftypes.NewUUID(), Hash: fftypes.NewRandB32()},
	}, nil)
	mdi.On("InsertMessageLocal", ctx, mock.Anything).Return(nil)

	msg, err := bm.BroadcastMessage(ctx, "ns1", &fftypes.MessageInput{
		Message: fftypes.Message{
			Header: fftypes.MessageHeader{
				Author: "0x12345",
			},
		},
		InputData: fftypes.InputData{
			{Value: fftypes.Byteable(`{"hello": "world"}`)},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, msg.Data[0].ID)
	assert.NotNil(t, msg.Data[0].Hash)
	assert.Equal(t, "ns1", msg.Header.Namespace)

	mdi.AssertExpectations(t)
	mdm.AssertExpectations(t)
}

func TestBroadcastMessageBadInput(t *testing.T) {
	bm, cancel := newTestBroadcast(t)
	defer cancel()
	mdi := bm.database.(*databasemocks.Plugin)
	mdm := bm.data.(*datamocks.Manager)
	mbi := bm.blockchain.(*blockchainmocks.Plugin)

	ctx := context.Background()
	mbi.On("VerifyIdentitySyntax", ctx, mock.Anything).Return("0x12345", nil)
	rag := mdi.On("RunAsGroup", ctx, mock.Anything)
	rag.RunFn = func(a mock.Arguments) {
		var fn = a[1].(func(context.Context) error)
		rag.ReturnArguments = mock.Arguments{fn(a[0].(context.Context))}
	}
	mdm.On("ResolveInputData", ctx, "ns1", mock.Anything).Return(nil, fmt.Errorf("pop"))

	_, err := bm.BroadcastMessage(ctx, "ns1", &fftypes.MessageInput{
		InputData: fftypes.InputData{
			{Value: fftypes.Byteable(`{"hello": "world"}`)},
		},
	})
	assert.EqualError(t, err, "pop")

	mdi.AssertExpectations(t)
	mdm.AssertExpectations(t)
}
