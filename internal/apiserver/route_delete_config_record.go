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
	"net/http"

	"github.com/hyperledger-labs/firefly/internal/i18n"
	"github.com/hyperledger-labs/firefly/internal/oapispec"
)

var deleteConfigRecord = &oapispec.Route{
	Name:   "deleteConfigRecord",
	Path:   "config/{key}",
	Method: http.MethodDelete,
	PathParams: []*oapispec.PathParam{
		{Name: "key", Example: "database", Description: i18n.MsgTBD},
	},
	QueryParams:     nil,
	FilterFactory:   nil,
	Description:     i18n.MsgTBD,
	JSONInputValue:  nil,
	JSONOutputValue: nil,
	JSONOutputCode:  http.StatusNoContent,
	JSONHandler: func(r oapispec.APIRequest) (output interface{}, err error) {
		err = r.Or.DeleteConfigRecord(r.Ctx, r.PP["key"])
		return nil, err
	},
}
