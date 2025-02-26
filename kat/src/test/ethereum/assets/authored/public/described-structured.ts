// Copyright © 2021 Kaleido, Inc.
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

import assert from 'assert';
import { createHash, randomBytes } from 'crypto';
import nock from 'nock';
import request from 'supertest';
import { promisify } from 'util';
import { IDBAssetDefinition, IDBAssetInstance, IEventAssetDefinitionCreated, IEventAssetInstanceBatchCreated } from '../../../../../lib/interfaces';
import * as utils from '../../../../../lib/utils';
import { app, mockEventStreamWebSocket } from '../../../../common';
import { testContent, testDescription } from '../../../../samples';
const delay = promisify(setTimeout);

export const testAuthoredPublicDescribedStructured = () => {

describe('Assets: authored - public - described - structured', async () => {

  let assetDefinitionID: string;
  const assetDefinitionName = 'authored - public - described - structured';
  const timestamp = utils.getTimestamp();
  const batchHashSha256 = '0x' + createHash('sha256').update(randomBytes(10)).digest().toString('hex');
  const batchHashIPFSMulti = utils.sha256ToIPFSHash(batchHashSha256);

  let batchMaxRecordsToRestore: number;
  beforeEach(() => {
    nock.cleanAll();
    // Force batches to close immediately
    batchMaxRecordsToRestore = utils.constants.BATCH_MAX_RECORDS;
    utils.constants.BATCH_MAX_RECORDS = 1;
  });

  afterEach(() => {
    assert.deepStrictEqual(nock.pendingMocks(), []);
    utils.constants.BATCH_MAX_RECORDS = batchMaxRecordsToRestore;
  });

  describe('Create asset definition', () => {

    it('Checks that the asset definition can be added', async () => {

      nock('https://apigateway.kaleido.io')
        .post('/createAssetDefinition?kld-from=0x0000000000000000000000000000000000000001&kld-sync=false')
        .reply(200, { id: 'my-receipt-id' });

      nock('https://ipfs.kaleido.io')
        .post('/api/v0/add')
        .reply(200, { Hash: 'QmPsTQxwhQjJsCFh3hBijCC4gGReQpEz5VjcCK7gu6sYXX' })

      const result = await request(app)
        .post('/api/v1/assets/definitions')
        .send({
          name: assetDefinitionName,
          author: '0x0000000000000000000000000000000000000001',
          isContentPrivate: false,
          isContentUnique: true,
          descriptionSchema: testDescription.schema.object,
          contentSchema: testContent.schema.object
        })
        .expect(200);
      assert.deepStrictEqual(result.body.status, 'submitted');
      assetDefinitionID = result.body.assetDefinitionID;

      const getAssetDefinitionsResponse = await request(app)
        .get('/api/v1/assets/definitions')
        .expect(200);
      const assetDefinition = getAssetDefinitionsResponse.body.find((assetDefinition: IDBAssetDefinition) => assetDefinition.name === 'authored - public - described - structured');
      assert.strictEqual(assetDefinition.assetDefinitionID, assetDefinitionID);
      assert.strictEqual(assetDefinition.author, '0x0000000000000000000000000000000000000001');
      assert.strictEqual(assetDefinition.isContentPrivate, false);
      assert.strictEqual(assetDefinition.isContentUnique, true);
      assert.deepStrictEqual(assetDefinition.descriptionSchema, testDescription.schema.object);
      assert.deepStrictEqual(assetDefinition.contentSchema, testContent.schema.object);
      assert.strictEqual(assetDefinition.name, 'authored - public - described - structured');
      assert.strictEqual(assetDefinition.receipt, 'my-receipt-id');
      assert.strictEqual(typeof assetDefinition.submitted, 'number');
    });

    it('Checks that the event stream notification for confirming the asset definition creation is handled', async () => {
      const eventPromise = new Promise<void>((resolve) => {
        mockEventStreamWebSocket.once('send', message => {
          assert.strictEqual(message, '{"type":"ack","topic":"dev"}');
          resolve();
        })
      });
      nock('https://ipfs.kaleido.io')
        .get('/ipfs/QmPsTQxwhQjJsCFh3hBijCC4gGReQpEz5VjcCK7gu6sYXX')
        .reply(200, {
          assetDefinitionID: assetDefinitionID,
          name: assetDefinitionName,
          isContentPrivate: false,
          isContentUnique: true,
          descriptionSchema: testDescription.schema.object,
          contentSchema: testContent.schema.object
        });
      const data: IEventAssetDefinitionCreated = {
        author: '0x0000000000000000000000000000000000000001',
        assetDefinitionHash: '0x16bfeb8bb4befe63cc90578fe6fa0f4ef56955c462cd57965279effc69df1ea6',
        timestamp: timestamp.toString()
      };
      mockEventStreamWebSocket.emit('message', JSON.stringify([{
        signature: utils.contractEventSignatures.ASSET_DEFINITION_CREATED,
        data,
        blockNumber: '123',
        transactionHash: '0x0000000000000000000000000000000000000000000000000000000000000000'
      }]));
      await eventPromise;
    });

    it('Checks that the asset definition is confirmed', async () => {
      const getAssetDefinitionsResponse = await request(app)
        .get('/api/v1/assets/definitions')
        .expect(200);
      const assetDefinition = getAssetDefinitionsResponse.body.find((assetDefinition: IDBAssetDefinition) => assetDefinition.name === 'authored - public - described - structured');
      assert.strictEqual(assetDefinition.assetDefinitionID, assetDefinitionID);
      assert.strictEqual(assetDefinition.author, '0x0000000000000000000000000000000000000001');
      assert.strictEqual(assetDefinition.isContentPrivate, false);
      assert.strictEqual(assetDefinition.isContentUnique, true);
      assert.deepStrictEqual(assetDefinition.descriptionSchema, testDescription.schema.object);
      assert.deepStrictEqual(assetDefinition.contentSchema, testContent.schema.object);
      assert.strictEqual(assetDefinition.name, 'authored - public - described - structured');
      assert.strictEqual(typeof assetDefinition.submitted, 'number');
      assert.strictEqual(assetDefinition.timestamp, timestamp);
      assert.strictEqual(assetDefinition.receipt, 'my-receipt-id');
      assert.strictEqual(assetDefinition.blockNumber, 123);
      assert.strictEqual(assetDefinition.transactionHash, '0x0000000000000000000000000000000000000000000000000000000000000000');

      const getAssetDefinitionResponse = await request(app)
        .get(`/api/v1/assets/definitions/${assetDefinitionID}`)
        .expect(200);
      assert.deepStrictEqual(assetDefinition, getAssetDefinitionResponse.body);
    });

  });

  describe('Asset instances', async () => {

    let assetInstanceID: string;

    it('Checks that an asset instance can be created', async () => {

      nock('https://apigateway.kaleido.io')
        .post('/createAssetInstanceBatch?kld-from=0x0000000000000000000000000000000000000001&kld-sync=false')
        .reply(200, { id: 'my-receipt-id' });

      nock('https://ipfs.kaleido.io')
        .post('/api/v0/add')
        .reply(200, { Hash: batchHashIPFSMulti })

      const result = await request(app)
        .post(`/api/v1/assets/${assetDefinitionID}`)
        .send({
          author: '0x0000000000000000000000000000000000000001',
          description: testDescription.sample.object,
          content: testContent.sample.object
        })
        .expect(200);
      assert.deepStrictEqual(result.body.status, 'submitted');
      assetInstanceID = result.body.assetInstanceID;

      const getAssetInstancesResponse = await request(app)
        .get(`/api/v1/assets/${assetDefinitionID}`)
        .expect(200);
      const assetInstance = getAssetInstancesResponse.body.find((assetInstance: IDBAssetInstance) => assetInstance.assetInstanceID === assetInstanceID);
      assert.strictEqual(assetInstance.author, '0x0000000000000000000000000000000000000001');
      assert.strictEqual(assetInstance.assetDefinitionID, assetDefinitionID);
      assert.strictEqual(assetInstance.descriptionHash, '0x' + utils.getSha256(JSON.stringify(testDescription.sample.object)));
      assert.deepStrictEqual(assetInstance.description, testDescription.sample.object);
      assert.strictEqual(assetInstance.contentHash,'0x' + utils.getSha256(JSON.stringify(testContent.sample.object)));
      assert.deepStrictEqual(assetInstance.content, testContent.sample.object);
      assert.strictEqual(assetInstance.receipt, undefined); // As this has been batched
      assert.strictEqual(typeof assetInstance.submitted, 'number');
      assert.strictEqual(typeof assetInstance.batchID, 'string');

      // Expect the batch to have been submitted
      let getBatchResponse: any;
      for (let i = 0; i < 10; i++) {
        getBatchResponse = await request(app)
          .get(`/api/v1/batches/${assetInstance.batchID}`)
          .expect(200);
        if (getBatchResponse.body.completed) break;
        await delay(1);
      }
      assert.strictEqual(typeof getBatchResponse.body.completed, 'number');
      assert.strictEqual(typeof getBatchResponse.body.batchHash, 'string');
      assert.strictEqual(getBatchResponse.body.receipt, 'my-receipt-id');
      assert.strictEqual(getBatchResponse.body.batchHash, batchHashSha256);
      // As this is a public asset, the full content will have been written to IPFS in the batch
      assert.deepStrictEqual(getBatchResponse.body.records[0].content, testContent.sample.object);
      // The full description payload should be in the batch data written to IPFS in the batch
      assert.deepStrictEqual(getBatchResponse.body.records[0].description, testDescription.sample.object);

      const getAssetInstanceResponse = await request(app)
        .get(`/api/v1/assets/${assetDefinitionID}/${assetInstanceID}`)
        .expect(200);
      assert.deepStrictEqual(assetInstance, getAssetInstanceResponse.body);

    });

    it('Checks that the event stream notification for confirming the asset instance creation is handled', async () => {
      const eventPromise = new Promise<void>((resolve) => {
        mockEventStreamWebSocket.once('send', message => {
          assert.strictEqual(message, '{"type":"ack","topic":"dev"}');
          resolve();
        })
      });
      const data: IEventAssetInstanceBatchCreated = {
        author: '0x0000000000000000000000000000000000000001',
        batchHash: batchHashSha256,
        timestamp: timestamp.toString()
      };

      mockEventStreamWebSocket.emit('message', JSON.stringify([{
        signature: utils.contractEventSignatures.ASSET_INSTANCE_BATCH_CREATED,
        data,
        blockNumber: '123',
        transactionHash: '0x0000000000000000000000000000000000000000000000000000000000000000'
      }]));
      await eventPromise;
    });

    it('Checks that the asset instance is confirmed', async () => {
      const getAssetInstancesResponse = await request(app)
        .get(`/api/v1/assets/${assetDefinitionID}`)
        .expect(200);
      const assetInstance = getAssetInstancesResponse.body.find((assetInstance: IDBAssetInstance) => assetInstance.assetInstanceID === assetInstanceID);
      assert.strictEqual(assetInstance.author, '0x0000000000000000000000000000000000000001');
      assert.strictEqual(assetInstance.assetDefinitionID, assetDefinitionID);
      assert.strictEqual(assetInstance.descriptionHash, '0x' + utils.getSha256(JSON.stringify(testDescription.sample.object)));
      assert.deepStrictEqual(assetInstance.description, testDescription.sample.object);
      assert.strictEqual(assetInstance.contentHash, '0x' + utils.getSha256(JSON.stringify(testContent.sample.object)));
      assert.deepStrictEqual(assetInstance.content, testContent.sample.object);
      assert.strictEqual(assetInstance.receipt, undefined); // the batch has the receipt
      assert.strictEqual(typeof assetInstance.timestamp, 'number');
      assert.strictEqual(assetInstance.timestamp, timestamp);
      assert.strictEqual(assetInstance.blockNumber, 123);
      assert.strictEqual(assetInstance.transactionHash, '0x0000000000000000000000000000000000000000000000000000000000000000');

      const getAssetInstanceResponse = await request(app)
        .get(`/api/v1/assets/${assetDefinitionID}/${assetInstanceID}`)
        .expect(200);
      assert.deepStrictEqual(assetInstance, getAssetInstanceResponse.body);
    });

  });

});
};
