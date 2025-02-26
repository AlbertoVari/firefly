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

package apiserver

import (
	"net/http/httptest"
	"testing"

	"github.com/hyperledger-labs/firefly/mocks/networkmapmocks"
	"github.com/hyperledger-labs/firefly/mocks/orchestratormocks"
	"github.com/hyperledger-labs/firefly/pkg/fftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetOrg(t *testing.T) {
	o := &orchestratormocks.Orchestrator{}
	nmn := &networkmapmocks.Manager{}
	o.On("NetworkMap").Return(nmn)
	r := createMuxRouter(o)
	req := httptest.NewRequest("GET", "/api/v1/network/organizations/org12345", nil)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	res := httptest.NewRecorder()

	nmn.On("GetOrganizationByID", mock.Anything, "org12345").
		Return(&fftypes.Organization{}, nil)
	r.ServeHTTP(res, req)

	assert.Equal(t, 200, res.Result().StatusCode)
}
