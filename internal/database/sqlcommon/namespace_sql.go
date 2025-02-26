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
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/hyperledger-labs/firefly/internal/i18n"
	"github.com/hyperledger-labs/firefly/internal/log"
	"github.com/hyperledger-labs/firefly/pkg/database"
	"github.com/hyperledger-labs/firefly/pkg/fftypes"
)

var (
	namespaceColumns = []string{
		"id",
		"message_id",
		"ntype",
		"name",
		"description",
		"created",
	}
	namespaceFilterTypeMap = map[string]string{
		"message": "message_id",
		"type":    "ntype",
	}
)

func (s *SQLCommon) UpsertNamespace(ctx context.Context, namespace *fftypes.Namespace, allowExisting bool) (err error) {
	ctx, tx, autoCommit, err := s.beginOrUseTx(ctx)
	if err != nil {
		return err
	}
	defer s.rollbackTx(ctx, tx, autoCommit)

	existing := false
	if allowExisting {
		// Do a select within the transaction to detemine if the UUID already exists
		namespaceRows, err := s.queryTx(ctx, tx,
			sq.Select("id").
				From("namespaces").
				Where(sq.Eq{"name": namespace.Name}),
		)
		if err != nil {
			return err
		}
		existing = namespaceRows.Next()

		if existing {
			var id fftypes.UUID
			_ = namespaceRows.Scan(&id)
			if namespace.ID != nil {
				if *namespace.ID != id {
					namespaceRows.Close()
					return database.IDMismatch
				}
			}
			namespace.ID = &id // Update on returned object
		}
		namespaceRows.Close()
	}

	if existing {
		// Update the namespace
		if err = s.updateTx(ctx, tx,
			sq.Update("namespaces").
				// Note we do not update ID
				Set("message_id", namespace.Message).
				Set("ntype", string(namespace.Type)).
				Set("name", namespace.Name).
				Set("description", namespace.Description).
				Set("created", namespace.Created).
				Where(sq.Eq{"name": namespace.Name}),
		); err != nil {
			return err
		}
	} else {
		if namespace.ID == nil {
			namespace.ID = fftypes.NewUUID()
		}

		if _, err = s.insertTx(ctx, tx,
			sq.Insert("namespaces").
				Columns(namespaceColumns...).
				Values(
					namespace.ID,
					namespace.Message,
					string(namespace.Type),
					namespace.Name,
					namespace.Description,
					namespace.Created,
				),
		); err != nil {
			return err
		}
	}

	return s.commitTx(ctx, tx, autoCommit)
}

func (s *SQLCommon) namespaceResult(ctx context.Context, row *sql.Rows) (*fftypes.Namespace, error) {
	namespace := fftypes.Namespace{}
	err := row.Scan(
		&namespace.ID,
		&namespace.Message,
		&namespace.Type,
		&namespace.Name,
		&namespace.Description,
		&namespace.Created,
	)
	if err != nil {
		return nil, i18n.WrapError(ctx, err, i18n.MsgDBReadErr, "namespaces")
	}
	return &namespace, nil
}

func (s *SQLCommon) GetNamespace(ctx context.Context, name string) (message *fftypes.Namespace, err error) {

	rows, err := s.query(ctx,
		sq.Select(namespaceColumns...).
			From("namespaces").
			Where(sq.Eq{"name": name}),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		log.L(ctx).Debugf("Namespace '%s' not found", name)
		return nil, nil
	}

	namespace, err := s.namespaceResult(ctx, rows)
	if err != nil {
		return nil, err
	}

	return namespace, nil
}

func (s *SQLCommon) GetNamespaces(ctx context.Context, filter database.Filter) (message []*fftypes.Namespace, err error) {

	query, err := s.filterSelect(ctx, "", sq.Select(namespaceColumns...).From("namespaces"), filter, namespaceFilterTypeMap)
	if err != nil {
		return nil, err
	}

	rows, err := s.query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	namespace := []*fftypes.Namespace{}
	for rows.Next() {
		d, err := s.namespaceResult(ctx, rows)
		if err != nil {
			return nil, err
		}
		namespace = append(namespace, d)
	}

	return namespace, err

}

func (s *SQLCommon) UpdateNamespace(ctx context.Context, id *fftypes.UUID, update database.Update) (err error) {

	ctx, tx, autoCommit, err := s.beginOrUseTx(ctx)
	if err != nil {
		return err
	}
	defer s.rollbackTx(ctx, tx, autoCommit)

	query, err := s.buildUpdate(sq.Update("namespaces"), update, namespaceFilterTypeMap)
	if err != nil {
		return err
	}
	query = query.Where(sq.Eq{"id": id})

	err = s.updateTx(ctx, tx, query)
	if err != nil {
		return err
	}

	return s.commitTx(ctx, tx, autoCommit)
}

func (s *SQLCommon) DeleteNamespace(ctx context.Context, id *fftypes.UUID) (err error) {

	ctx, tx, autoCommit, err := s.beginOrUseTx(ctx)
	if err != nil {
		return err
	}
	defer s.rollbackTx(ctx, tx, autoCommit)

	err = s.deleteTx(ctx, tx, sq.Delete("namespaces").Where(sq.Eq{
		"id": id,
	}))
	if err != nil {
		return err
	}

	return s.commitTx(ctx, tx, autoCommit)
}
