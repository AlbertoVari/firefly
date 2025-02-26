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

package fftypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEvent(t *testing.T) {

	u := NewUUID()
	g := NewRandB32()
	e := NewEvent(EventTypeMessageConfirmed, "ns1", u, g)
	assert.Equal(t, EventTypeMessageConfirmed, e.Type)
	assert.Equal(t, "ns1", e.Namespace)
	assert.Equal(t, *u, *e.Reference)
	assert.Equal(t, *g, *e.Group)

	e.Sequence = 12345
	var ls LocallySequenced = e
	assert.Equal(t, int64(12345), ls.LocalSequence())

}
