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

package sqlcommon

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hyperledger-labs/firefly/internal/log"
	"github.com/hyperledger-labs/firefly/pkg/database"
	"github.com/hyperledger-labs/firefly/pkg/fftypes"
	"github.com/stretchr/testify/assert"
)

func TestOffsetsE2EWithDB(t *testing.T) {
	log.SetLevel("debug")

	s := newQLTestProvider(t)
	defer s.Close()
	ctx := context.Background()

	// Create a new offset entry
	rand1, _ := rand.Int(rand.Reader, big.NewInt(10000000000000))
	offset := &fftypes.Offset{
		ID:        fftypes.NewUUID(),
		Type:      fftypes.OffsetTypeBatch,
		Namespace: "ns1",
		Name:      "offset1",
		Current:   rand1.Int64(),
	}
	err := s.UpsertOffset(ctx, offset, true)
	assert.NoError(t, err)

	// Check we get the exact same offset back
	offsetRead, err := s.GetOffset(ctx, offset.Type, offset.Namespace, offset.Name)
	assert.NoError(t, err)
	assert.NotNil(t, offsetRead)
	offsetJson, _ := json.Marshal(&offset)
	offsetReadJson, _ := json.Marshal(&offsetRead)
	assert.Equal(t, string(offsetJson), string(offsetReadJson))

	// Update the offset (this is testing what's possible at the database layer,
	// and does not account for the verification that happens at the higher level)
	rand2, _ := rand.Int(rand.Reader, big.NewInt(10000000000000))
	offsetUpdated := &fftypes.Offset{
		ID:        fftypes.NewUUID(),
		Type:      fftypes.OffsetTypeBatch,
		Namespace: "ns1",
		Name:      "offset1",
		Current:   rand2.Int64(),
	}

	// Attempt with wrong ID
	err = s.UpsertOffset(context.Background(), offsetUpdated, true)
	assert.Equal(t, err, database.IDMismatch)

	// Remove ID for upsert and retry
	offsetUpdated.ID = nil
	err = s.UpsertOffset(context.Background(), offsetUpdated, true)
	assert.NoError(t, err)

	// Check we get the exact same data back - note the removal of one of the offset elements
	offsetRead, err = s.GetOffset(ctx, offset.Type, offset.Namespace, offset.Name)
	assert.NoError(t, err)
	offsetJson, _ = json.Marshal(&offsetUpdated)
	offsetReadJson, _ = json.Marshal(&offsetRead)
	assert.Equal(t, string(offsetJson), string(offsetReadJson))

	// Query back the offset
	fb := database.OffsetQueryFactory.NewFilter(ctx)
	filter := fb.And(
		fb.Eq("type", string(offsetUpdated.Type)),
		fb.Eq("namespace", offsetUpdated.Namespace),
		fb.Eq("name", offsetUpdated.Name),
		fb.Gt("current", 0),
	)
	offsetRes, err := s.GetOffsets(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(offsetRes))
	offsetReadJson, _ = json.Marshal(offsetRes[0])
	assert.Equal(t, string(offsetJson), string(offsetReadJson))

	// Update
	rand3, _ := rand.Int(rand.Reader, big.NewInt(10000000000000))
	up := database.OffsetQueryFactory.NewUpdate(ctx).Set("current", rand3.Int64())
	err = s.UpdateOffset(ctx, offsetUpdated.ID, up)
	assert.NoError(t, err)

	// Test find updated value
	filter = fb.And(
		fb.Eq("name", offsetUpdated.Name),
		fb.Eq("current", rand3.Int64()),
	)
	offsets, err := s.GetOffsets(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(offsets))

	// Test delete
	err = s.DeleteOffset(ctx, fftypes.OffsetTypeBatch, offsetUpdated.Namespace, offsetUpdated.Name)
	assert.NoError(t, err)
	offsets, err = s.GetOffsets(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(offsets))

}

func TestUpsertOffsetFailBegin(t *testing.T) {
	s, mock := newMockProvider().init()
	mock.ExpectBegin().WillReturnError(fmt.Errorf("pop"))
	err := s.UpsertOffset(context.Background(), &fftypes.Offset{}, true)
	assert.Regexp(t, "FF10114", err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpsertOffsetFailSelect(t *testing.T) {
	s, mock := newMockProvider().init()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*").WillReturnError(fmt.Errorf("pop"))
	mock.ExpectRollback()
	err := s.UpsertOffset(context.Background(), &fftypes.Offset{Name: "name1"}, true)
	assert.Regexp(t, "FF10115", err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpsertOffsetFailInsert(t *testing.T) {
	s, mock := newMockProvider().init()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*").WillReturnRows(sqlmock.NewRows([]string{}))
	mock.ExpectExec("INSERT .*").WillReturnError(fmt.Errorf("pop"))
	mock.ExpectRollback()
	err := s.UpsertOffset(context.Background(), &fftypes.Offset{Name: "name1"}, true)
	assert.Regexp(t, "FF10116", err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpsertOffsetFailUpdate(t *testing.T) {
	s, mock := newMockProvider().init()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*").WillReturnRows(sqlmock.NewRows([]string{"otype", "namespace", "name"}).
		AddRow(fftypes.OffsetTypeBatch, "ns1", "name1"))
	mock.ExpectExec("UPDATE .*").WillReturnError(fmt.Errorf("pop"))
	mock.ExpectRollback()
	err := s.UpsertOffset(context.Background(), &fftypes.Offset{Name: "name1"}, true)
	assert.Regexp(t, "FF10117", err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpsertOffsetFailCommit(t *testing.T) {
	s, mock := newMockProvider().init()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*").WillReturnRows(sqlmock.NewRows([]string{"otype", "namespace", "name"}))
	mock.ExpectExec("INSERT .*").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit().WillReturnError(fmt.Errorf("pop"))
	err := s.UpsertOffset(context.Background(), &fftypes.Offset{Name: "name1"}, true)
	assert.Regexp(t, "FF10119", err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOffsetByIDSelectFail(t *testing.T) {
	s, mock := newMockProvider().init()
	mock.ExpectQuery("SELECT .*").WillReturnError(fmt.Errorf("pop"))
	_, err := s.GetOffset(context.Background(), fftypes.OffsetTypeBatch, "ns1", "name1")
	assert.Regexp(t, "FF10115", err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOffsetByIDNotFound(t *testing.T) {
	s, mock := newMockProvider().init()
	mock.ExpectQuery("SELECT .*").WillReturnRows(sqlmock.NewRows([]string{"otype", "namespace", "name"}))
	msg, err := s.GetOffset(context.Background(), fftypes.OffsetTypeBatch, "ns1", "name1")
	assert.NoError(t, err)
	assert.Nil(t, msg)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOffsetByIDScanFail(t *testing.T) {
	s, mock := newMockProvider().init()
	mock.ExpectQuery("SELECT .*").WillReturnRows(sqlmock.NewRows([]string{"otype"}).AddRow("only one"))
	_, err := s.GetOffset(context.Background(), fftypes.OffsetTypeBatch, "ns1", "name1")
	assert.Regexp(t, "FF10121", err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOffsetQueryFail(t *testing.T) {
	s, mock := newMockProvider().init()
	mock.ExpectQuery("SELECT .*").WillReturnError(fmt.Errorf("pop"))
	f := database.OffsetQueryFactory.NewFilter(context.Background()).Eq("type", "")
	_, err := s.GetOffsets(context.Background(), f)
	assert.Regexp(t, "FF10115", err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOffsetBuildQueryFail(t *testing.T) {
	s, _ := newMockProvider().init()
	f := database.OffsetQueryFactory.NewFilter(context.Background()).Eq("type", map[bool]bool{true: false})
	_, err := s.GetOffsets(context.Background(), f)
	assert.Regexp(t, "FF10149.*type", err)
}

func TestGetOffsetReadMessageFail(t *testing.T) {
	s, mock := newMockProvider().init()
	mock.ExpectQuery("SELECT .*").WillReturnRows(sqlmock.NewRows([]string{"otype"}).AddRow("only one"))
	f := database.OffsetQueryFactory.NewFilter(context.Background()).Eq("type", "")
	_, err := s.GetOffsets(context.Background(), f)
	assert.Regexp(t, "FF10121", err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOffsetUpdateBeginFail(t *testing.T) {
	s, mock := newMockProvider().init()
	mock.ExpectBegin().WillReturnError(fmt.Errorf("pop"))
	u := database.OffsetQueryFactory.NewUpdate(context.Background()).Set("name", "anything")
	err := s.UpdateOffset(context.Background(), fftypes.NewUUID(), u)
	assert.Regexp(t, "FF10114", err)
}

func TestOffsetUpdateBuildQueryFail(t *testing.T) {
	s, mock := newMockProvider().init()
	mock.ExpectBegin()
	u := database.OffsetQueryFactory.NewUpdate(context.Background()).Set("name", map[bool]bool{true: false})
	err := s.UpdateOffset(context.Background(), fftypes.NewUUID(), u)
	assert.Regexp(t, "FF10149.*name", err)
}

func TestOffsetUpdateFail(t *testing.T) {
	s, mock := newMockProvider().init()
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE .*").WillReturnError(fmt.Errorf("pop"))
	mock.ExpectRollback()
	u := database.OffsetQueryFactory.NewUpdate(context.Background()).Set("name", fftypes.NewUUID())
	err := s.UpdateOffset(context.Background(), fftypes.NewUUID(), u)
	assert.Regexp(t, "FF10117", err)
}

func TestOffsetDeleteBeginFail(t *testing.T) {
	s, mock := newMockProvider().init()
	mock.ExpectBegin().WillReturnError(fmt.Errorf("pop"))
	err := s.DeleteOffset(context.Background(), fftypes.OffsetTypeSubscription, "ns1", "sub1")
	assert.Regexp(t, "FF10114", err)
}

func TestOffsetDeleteFail(t *testing.T) {
	s, mock := newMockProvider().init()
	mock.ExpectBegin()
	mock.ExpectExec("DELETE .*").WillReturnError(fmt.Errorf("pop"))
	mock.ExpectRollback()
	err := s.DeleteOffset(context.Background(), fftypes.OffsetTypeSubscription, "ns1", "sub1")
	assert.Regexp(t, "FF10118", err)
}
