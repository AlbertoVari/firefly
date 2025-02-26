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

package difactory

import (
	"github.com/hyperledger-labs/firefly/internal/database/postgres"
	"github.com/hyperledger-labs/firefly/internal/database/ql"
	"github.com/hyperledger-labs/firefly/internal/database/sqlite"
	"github.com/hyperledger-labs/firefly/pkg/database"
)

var plugins = []database.Plugin{
	&postgres.Postgres{},
	&ql.QL{},
	&sqlite.SQLite{}, // we use the modernc implementation of SQLite that does not require CGO
}
