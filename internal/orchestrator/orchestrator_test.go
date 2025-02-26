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

package orchestrator

import (
	"context"
	"fmt"
	"testing"

	"github.com/hyperledger-labs/firefly/internal/config"
	"github.com/hyperledger-labs/firefly/mocks/batchmocks"
	"github.com/hyperledger-labs/firefly/mocks/blockchainmocks"
	"github.com/hyperledger-labs/firefly/mocks/broadcastmocks"
	"github.com/hyperledger-labs/firefly/mocks/databasemocks"
	"github.com/hyperledger-labs/firefly/mocks/dataexchangemocks"
	"github.com/hyperledger-labs/firefly/mocks/datamocks"
	"github.com/hyperledger-labs/firefly/mocks/eventmocks"
	"github.com/hyperledger-labs/firefly/mocks/identitymocks"
	"github.com/hyperledger-labs/firefly/mocks/networkmapmocks"
	"github.com/hyperledger-labs/firefly/mocks/privatemessagingmocks"
	"github.com/hyperledger-labs/firefly/mocks/publicstoragemocks"
	"github.com/hyperledger-labs/firefly/pkg/fftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const configDir = "../../test/data/config"

type testOrchestrator struct {
	orchestrator

	mdi *databasemocks.Plugin
	mdm *datamocks.Manager
	mbm *broadcastmocks.Manager
	mba *batchmocks.Manager
	mem *eventmocks.EventManager
	mnm *networkmapmocks.Manager
	mps *publicstoragemocks.Plugin
	mpm *privatemessagingmocks.Manager
	mbi *blockchainmocks.Plugin
	mii *identitymocks.Plugin
	mdx *dataexchangemocks.Plugin
}

func newTestOrchestrator() *testOrchestrator {
	tor := &testOrchestrator{
		orchestrator: orchestrator{
			ctx: context.Background(),
		},
		mdi: &databasemocks.Plugin{},
		mdm: &datamocks.Manager{},
		mbm: &broadcastmocks.Manager{},
		mba: &batchmocks.Manager{},
		mem: &eventmocks.EventManager{},
		mnm: &networkmapmocks.Manager{},
		mps: &publicstoragemocks.Plugin{},
		mpm: &privatemessagingmocks.Manager{},
		mbi: &blockchainmocks.Plugin{},
		mii: &identitymocks.Plugin{},
		mdx: &dataexchangemocks.Plugin{},
	}
	tor.orchestrator.database = tor.mdi
	tor.orchestrator.data = tor.mdm
	tor.orchestrator.batch = tor.mba
	tor.orchestrator.broadcast = tor.mbm
	tor.orchestrator.events = tor.mem
	tor.orchestrator.networkmap = tor.mnm
	tor.orchestrator.publicstorage = tor.mps
	tor.orchestrator.messaging = tor.mpm
	tor.orchestrator.blockchain = tor.mbi
	tor.orchestrator.identity = tor.mii
	tor.orchestrator.dataexchange = tor.mdx
	tor.mdi.On("Name").Return("mock-di").Maybe()
	tor.mem.On("Name").Return("mock-ei").Maybe()
	tor.mps.On("Name").Return("mock-ps").Maybe()
	tor.mbi.On("Name").Return("mock-bi").Maybe()
	tor.mii.On("Name").Return("mock-ii").Maybe()
	tor.mdx.On("Name").Return("mock-dx").Maybe()
	return tor
}

func TestNewOrchestrator(t *testing.T) {
	or := NewOrchestrator()
	assert.NotNil(t, or)
}

func TestBadDatabasePlugin(t *testing.T) {
	or := newTestOrchestrator()
	config.Set(config.DatabaseType, "wrong")
	or.database = nil
	ctx, cancelCtx := context.WithCancel(context.Background())
	err := or.Init(ctx, cancelCtx)
	assert.Regexp(t, "FF10122.*wrong", err)
}

func TestBadDatabaseInitFail(t *testing.T) {
	or := newTestOrchestrator()
	config.Set(config.DatabaseType, "wrong")
	or.mdi.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("pop"))
	ctx, cancelCtx := context.WithCancel(context.Background())
	err := or.Init(ctx, cancelCtx)
	assert.EqualError(t, err, "pop")
}

func TestBadIdentityPlugin(t *testing.T) {
	or := newTestOrchestrator()
	or.mdi.On("GetConfigRecords", mock.Anything, mock.Anything, mock.Anything).Return([]*fftypes.ConfigRecord{}, nil)
	config.Set(config.IdentityType, "wrong")
	or.identity = nil
	or.mdi.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	ctx, cancelCtx := context.WithCancel(context.Background())
	err := or.Init(ctx, cancelCtx)
	assert.Regexp(t, "FF10212.*wrong", err)
}

func TestBadIdentityInitFail(t *testing.T) {
	or := newTestOrchestrator()
	or.mdi.On("GetConfigRecords", mock.Anything, mock.Anything, mock.Anything).Return([]*fftypes.ConfigRecord{}, nil)
	or.mdi.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mii.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("pop"))
	ctx, cancelCtx := context.WithCancel(context.Background())
	err := or.Init(ctx, cancelCtx)
	assert.EqualError(t, err, "pop")
}

func TestBadBlockchainPlugin(t *testing.T) {
	or := newTestOrchestrator()
	config.Set(config.BlockchainType, "wrong")
	or.blockchain = nil
	or.mdi.On("GetConfigRecords", mock.Anything, mock.Anything, mock.Anything).Return([]*fftypes.ConfigRecord{}, nil)
	or.mdi.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mii.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	ctx, cancelCtx := context.WithCancel(context.Background())
	err := or.Init(ctx, cancelCtx)
	assert.Regexp(t, "FF10110.*wrong", err)
}

func TestBlockchaiInitFail(t *testing.T) {
	or := newTestOrchestrator()
	or.mdi.On("GetConfigRecords", mock.Anything, mock.Anything, mock.Anything).Return([]*fftypes.ConfigRecord{}, nil)
	or.mii.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mdi.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mbi.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("pop"))
	ctx, cancelCtx := context.WithCancel(context.Background())
	err := or.Init(ctx, cancelCtx)
	assert.EqualError(t, err, "pop")
}

func TestBlockchaiInitGetConfigRecordsFail(t *testing.T) {
	or := newTestOrchestrator()
	or.mdi.On("GetConfigRecords", mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("pop"))
	or.mii.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mdi.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mbi.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	ctx, cancelCtx := context.WithCancel(context.Background())
	err := or.Init(ctx, cancelCtx)
	assert.EqualError(t, err, "pop")
}

func TestBlockchaiInitMergeConfigRecordsFail(t *testing.T) {
	or := newTestOrchestrator()
	or.mdi.On("GetConfigRecords", mock.Anything, mock.Anything, mock.Anything).Return([]*fftypes.ConfigRecord{
		{
			Key:   "pizza.toppings",
			Value: []byte("cheese, pepperoni, mushrooms"),
		},
	}, nil)
	or.mii.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mdi.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mbi.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	ctx, cancelCtx := context.WithCancel(context.Background())
	err := or.Init(ctx, cancelCtx)
	assert.EqualError(t, err, "While parsing config: invalid character 'c' looking for beginning of value")
}

func TestBadPublicStoragePlugin(t *testing.T) {
	or := newTestOrchestrator()
	config.Set(config.PublicStorageType, "wrong")
	or.publicstorage = nil
	or.mdi.On("GetConfigRecords", mock.Anything, mock.Anything, mock.Anything).Return([]*fftypes.ConfigRecord{}, nil)
	or.mdi.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mbi.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mii.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mbi.On("VerifyIdentitySyntax", mock.Anything, mock.Anything, mock.Anything).Return("", nil)
	ctx, cancelCtx := context.WithCancel(context.Background())
	err := or.Init(ctx, cancelCtx)
	assert.Regexp(t, "FF10134.*wrong", err)
}

func TestBadPublicStorageInitFail(t *testing.T) {
	or := newTestOrchestrator()
	or.mdi.On("GetConfigRecords", mock.Anything, mock.Anything, mock.Anything).Return([]*fftypes.ConfigRecord{}, nil)
	or.mdi.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mbi.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mii.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mbi.On("VerifyIdentitySyntax", mock.Anything, mock.Anything, mock.Anything).Return("", nil)
	or.mps.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("pop"))
	ctx, cancelCtx := context.WithCancel(context.Background())
	err := or.Init(ctx, cancelCtx)
	assert.EqualError(t, err, "pop")
}

func TestBadDataExchangePlugin(t *testing.T) {
	or := newTestOrchestrator()
	config.Set(config.DataexchangeType, "wrong")
	or.dataexchange = nil
	or.mdi.On("GetConfigRecords", mock.Anything, mock.Anything, mock.Anything).Return([]*fftypes.ConfigRecord{}, nil)
	or.mdi.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mbi.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mii.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mbi.On("VerifyIdentitySyntax", mock.Anything, mock.Anything, mock.Anything).Return("", nil)
	or.mps.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	ctx, cancelCtx := context.WithCancel(context.Background())
	err := or.Init(ctx, cancelCtx)
	assert.Regexp(t, "FF10213.*wrong", err)
}

func TestBadPDataExchangeInitFail(t *testing.T) {
	or := newTestOrchestrator()
	or.mdi.On("GetConfigRecords", mock.Anything, mock.Anything, mock.Anything).Return([]*fftypes.ConfigRecord{}, nil)
	or.mdi.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mbi.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mii.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mbi.On("VerifyIdentitySyntax", mock.Anything, mock.Anything, mock.Anything).Return("", nil)
	or.mps.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mdx.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("pop"))
	ctx, cancelCtx := context.WithCancel(context.Background())
	err := or.Init(ctx, cancelCtx)
	assert.EqualError(t, err, "pop")
}

func TestInitMessagingomponentFail(t *testing.T) {
	or := newTestOrchestrator()
	or.database = nil
	or.messaging = nil
	err := or.initComponents(context.Background())
	assert.Regexp(t, "FF10128", err)
}

func TestInitEventsComponentFail(t *testing.T) {
	or := newTestOrchestrator()
	or.database = nil
	or.events = nil
	err := or.initComponents(context.Background())
	assert.Regexp(t, "FF10128", err)
}

func TestInitNetworkMapComponentFail(t *testing.T) {
	or := newTestOrchestrator()
	or.database = nil
	or.networkmap = nil
	err := or.initComponents(context.Background())
	assert.Regexp(t, "FF10128", err)
}

func TestInitBatchComponentFail(t *testing.T) {
	or := newTestOrchestrator()
	or.database = nil
	or.batch = nil
	err := or.initComponents(context.Background())
	assert.Regexp(t, "FF10128", err)
}

func TestInitBroadcastComponentFail(t *testing.T) {
	or := newTestOrchestrator()
	or.database = nil
	or.broadcast = nil
	err := or.initComponents(context.Background())
	assert.Regexp(t, "FF10128", err)
}

func TestInitDataComponentFail(t *testing.T) {
	or := newTestOrchestrator()
	or.database = nil
	or.data = nil
	err := or.initComponents(context.Background())
	assert.Regexp(t, "FF10128", err)
}

func TestStartBatchFail(t *testing.T) {
	config.Reset()
	or := newTestOrchestrator()
	or.mba.On("Start").Return(fmt.Errorf("pop"))
	or.mbi.On("Start").Return(nil)
	err := or.Start()
	assert.Regexp(t, "pop", err)
}

func TestStartStopOk(t *testing.T) {
	config.Reset()
	or := newTestOrchestrator()
	or.mbi.On("Start").Return(nil)
	or.mba.On("Start").Return(nil)
	or.mem.On("Start").Return(nil)
	or.mbm.On("Start").Return(nil)
	or.mpm.On("Start").Return(nil)
	or.mbi.On("WaitStop").Return(nil)
	or.mba.On("WaitStop").Return(nil)
	or.mem.On("WaitStop").Return(nil)
	or.mbm.On("WaitStop").Return(nil)
	err := or.Start()
	assert.NoError(t, err)
	or.WaitStop()
	or.WaitStop() // swallows dups
}

func TestInitNamespacesBadName(t *testing.T) {
	config.Reset()
	config.Set(config.NamespacesPredefined, fftypes.JSONObjectArray{
		{"name": "!Badness"},
	})
	or := newTestOrchestrator()
	err := or.initNamespaces(context.Background())
	assert.Regexp(t, "FF10131", err)
}

func TestInitNamespacesGetFail(t *testing.T) {
	config.Reset()
	or := newTestOrchestrator()
	or.mdi.On("GetNamespace", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("pop"))
	err := or.initNamespaces(context.Background())
	assert.Regexp(t, "pop", err)
}

func TestInitNamespacesUpsertFail(t *testing.T) {
	config.Reset()
	or := newTestOrchestrator()
	or.mdi.On("GetNamespace", mock.Anything, mock.Anything).Return(nil, nil)
	or.mdi.On("UpsertNamespace", mock.Anything, mock.Anything, true).Return(fmt.Errorf("pop"))
	err := or.initNamespaces(context.Background())
	assert.Regexp(t, "pop", err)
}

func TestInitNamespacesUpsertNotNeeded(t *testing.T) {
	config.Reset()
	or := newTestOrchestrator()
	or.mdi.On("GetNamespace", mock.Anything, mock.Anything).Return(&fftypes.Namespace{
		Type: fftypes.NamespaceTypeBroadcast, // any broadcasted NS will not be updated
	}, nil)
	err := or.initNamespaces(context.Background())
	assert.NoError(t, err)
}

func TestInitNamespacesDefaultMissing(t *testing.T) {
	config.Reset()
	or := newTestOrchestrator()
	config.Set(config.NamespacesPredefined, fftypes.JSONObjectArray{})
	err := or.initNamespaces(context.Background())
	assert.Regexp(t, "FF10166", err)
}

func TestInitNamespacesDupName(t *testing.T) {
	config.Reset()
	or := newTestOrchestrator()
	config.Set(config.NamespacesPredefined, fftypes.JSONObjectArray{
		{"name": "ns1"},
		{"name": "ns2"},
		{"name": "ns2"},
	})
	config.Set(config.NamespacesDefault, "ns1")
	nsList, err := or.getPrefdefinedNamespaces(context.Background())
	assert.NoError(t, err)
	assert.Len(t, nsList, 3)
	assert.Equal(t, fftypes.SystemNamespace, nsList[0].Name)
	assert.Equal(t, "ns1", nsList[1].Name)
	assert.Equal(t, "ns2", nsList[2].Name)
}

func TestInitOK(t *testing.T) {
	or := newTestOrchestrator()
	or.mdi.On("GetConfigRecords", mock.Anything, mock.Anything, mock.Anything).Return([]*fftypes.ConfigRecord{}, nil)
	or.mdi.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mii.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mbi.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mbi.On("VerifyIdentitySyntax", mock.Anything, mock.Anything, mock.Anything).Return("", nil)
	or.mps.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mdx.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	or.mdi.On("GetNamespace", mock.Anything, mock.Anything).Return(nil, nil)
	or.mdi.On("UpsertNamespace", mock.Anything, mock.Anything, true).Return(nil)
	err := config.ReadConfig(configDir + "/firefly.core.yaml")
	assert.NoError(t, err)
	ctx, cancelCtx := context.WithCancel(context.Background())
	err = or.Init(ctx, cancelCtx)
	assert.NoError(t, err)

	assert.Equal(t, or.mbm, or.Broadcast())
	assert.Equal(t, or.mpm, or.PrivateMessaging())
	assert.Equal(t, or.mem, or.Events())
	assert.Equal(t, or.mnm, or.NetworkMap())
	assert.Equal(t, or.mdm, or.Data())
}
