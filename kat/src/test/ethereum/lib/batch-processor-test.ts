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
import sinon, { SinonStub } from 'sinon';
import { promisify } from 'util';
import * as database from '../../../clients/database';
import { BatchProcessor } from '../../../lib/batch-processor';
import { IDBBatch, BatchRecordType } from '../../../lib/interfaces';

const delay = promisify(setTimeout);

export const testBatchProcessor = async () => {

describe('BatchProcessor', () => {

  beforeEach(() => {
    sinon.stub(database, 'retrieveBatches');
    sinon.stub(database, 'upsertBatch');
  });

  afterEach(() => {
    (database.retrieveBatches as SinonStub).restore();
    (database.upsertBatch as SinonStub).restore();
  })

  it('fills a batch full in parallel and dispatches it, then cleans up', async () => {

    const processBatchCallback = sinon.stub();
    const processorCompleteCallback = sinon.stub();
    const p = new BatchProcessor(
      'author1',
      'type1',
      processBatchCallback,
      processorCompleteCallback,
    );

    const scheduleRandomDelayedAdd = async (i: number) => {
      // Introduce some randomness, but with very short delays to keep the test fast
      await delay(Math.ceil(Math.random() * 5));
      // Half and half records vs. properties
      if (i % 2 === 0) {
        await p.add({
          recordType: BatchRecordType.assetInstance,
          id: `test_${i}`,
        });  
      } else {
        await p.add({
          recordType: BatchRecordType.assetProperty,
          key: `test_${i}`,
          value: `value_${i}`,
        });  
      }
    }

    const promises: Promise<void>[] = [];
    for (let i = 0; i < p.config.batchMaxRecords; i++) {
      promises.push(scheduleRandomDelayedAdd(i));
    }
    await Promise.all(promises);

    assert.strictEqual(processBatchCallback.callCount, 1);
    assert.strictEqual(processorCompleteCallback.callCount, 1);
    const batch: IDBBatch = processBatchCallback.getCall(0).args[0];
    for (let i = 0; i < p.config.batchMaxRecords; i++) {
      // Half and half records vs. properties
      if (i % 2 === 0) {
        assert(batch.records.find(r => r.id === `test_${i}`));
      } else {
        assert(batch.records.find(r => r.key === `test_${i}`));
        assert(batch.records.find(r => r.value === `value_${i}`));
      }
    }

  });

  it('takes a batch array on input simulating recovery, and dispatches immediately', async () => {

    const processBatchCallback = sinon.stub();
    const processorCompleteCallback = sinon.stub();
    const p = new BatchProcessor(
      'author1',
      'type1',
      processBatchCallback,
      processorCompleteCallback,
    );

    let batch: IDBBatch = {
      author: 'author1',
      type: 'type1',
      batchID: 'batch1',
      completed: null,
      created: Date.now(),
      records: [],
    };
    for (let i = 0; i < p.config.batchMaxRecords-1; i++) {
      assert(batch.records.push({id: `test_${i}`, recordType: BatchRecordType.assetInstance }));
    }
    await p.init([batch]);

    assert.strictEqual(processBatchCallback.callCount, 1);
    assert.strictEqual(processorCompleteCallback.callCount, 1);
    batch = processBatchCallback.getCall(0).args[0];
    for (let i = 0; i < p.config.batchMaxRecords-1; i++) {
      assert(batch.records.find(r => r.id === `test_${i}`));
    }

  });

  it('times out a batch with arrival, then cleans up once it dispatches', async () => {

    const processBatchCallback = sinon.stub();
    const processorCompleteCallback = sinon.stub();
    const p = new BatchProcessor(
      'author1',
      'type1',
      processBatchCallback,
      processorCompleteCallback,
      {
        batchTimeoutArrivallMS: 10,
      }
    );

    const scheduleRandomDelayedAdd = async (i: number) => {
      // Introduce some randomness, but with very short delays to keep the test fast
      await delay(Math.ceil(Math.random() * 5));
      await p.add({
        recordType: BatchRecordType.assetInstance,
        id: `test_${i}`,
      });
    }

    const before = Date.now();
    const promises: Promise<void>[] = [];
    for (let i = 0; i < (p.config.batchMaxRecords - 1); i++) {
      promises.push(scheduleRandomDelayedAdd(i));
    }
    await Promise.all(promises);

    // Should not be set yet - wait for timeout
    assert.strictEqual(processBatchCallback.callCount, 0);
    assert.strictEqual(processorCompleteCallback.callCount, 0);

    for (let i = 0; i < 100; i++) {
      if (processBatchCallback.callCount === 0) await delay(1);
    }
    const after = Date.now();

    assert(after - before >= 10 /* we must have waited this long */)
    assert.strictEqual(processBatchCallback.callCount, 1);
    assert.strictEqual(processorCompleteCallback.callCount, 1);

    const batch: IDBBatch = processBatchCallback.getCall(0).args[0];
    for (let i = 0; i < (p.config.batchMaxRecords - 1); i++) {
      assert(batch.records.find(r => r.id === `test_${i}`));
    }

  });

  it('times out a batch with an overall timeout, then continues to add to the next batch', async () => {

    let totalReceived = 0;
    let batchCount = 0;
    const processorCompleteCallback = sinon.stub();
    const p = new BatchProcessor(
      'author1',
      'type1',
      async b => {totalReceived += b.records.length; batchCount++},
      processorCompleteCallback,
      {
        batchTimeoutArrivallMS: 10,
        batchTimeoutOverallMS: 20,
      }
    );

    for (let i = 0; i < 50; i++) {
      await delay(1);
      await p.add({
        recordType: BatchRecordType.assetInstance,
        id: `test_${i}`,
      });
    }

    while (totalReceived < 50) await delay(5);

    assert(batchCount > 1);
    assert(processorCompleteCallback.callCount >= 1);

  });

  it('fills a batch with a slow persistence to the DB', async () => {

    const processBatchCallback = sinon.stub();
    const processorCompleteCallback = sinon.stub();
    const p = new BatchProcessor(
      'author1',
      'type1',
      processBatchCallback,
      processorCompleteCallback,
      {
        batchMaxRecords: 10,
      }
    );

    // Make the persistence slow
    const dbUpdateStub = (database.upsertBatch as SinonStub);
    dbUpdateStub.callsFake(() => delay(10))

    // Make the adding fast
    const addImmediate = async (i: number) => {
      await p.add({
        recordType: BatchRecordType.assetInstance,
        id: `test_${i}`,
      });
    }
    const promises: Promise<void>[] = [];
    for (let i = 0; i < p.config.batchMaxRecords; i++) {
      promises.push(addImmediate(i));
    }
    await Promise.all(promises);

    for (let i = 0; i < 100; i++) {
      if (processorCompleteCallback.callCount === 0) await delay(1);
    }

    // We should have exactly three calls
    // - once for the first as the batch started
    // - once with everything else in the batch
    // - once when we completed the batch
    assert.strictEqual(dbUpdateStub.callCount, 3);

    assert.strictEqual(processBatchCallback.callCount, 1);
    assert.strictEqual(processorCompleteCallback.callCount, 1);
    const batch: IDBBatch = processBatchCallback.getCall(0).args[0];
    for (let i = 0; i < p.config.batchMaxRecords; i++) {
      assert(batch.records.find(r => r.id === `test_${i}`));
    }

  });

  it('handles a failure to persist the batch to the DB', async () => {

    const processBatchCallback = sinon.stub();
    const processorCompleteCallback = sinon.stub();
    const p = new BatchProcessor(
      'author1',
      'type1',
      processBatchCallback,
      processorCompleteCallback,
    );

    // Make the persistence fail
    (database.upsertBatch as SinonStub).rejects(new Error('pop'));

    let failed;
    try {
      await p.add({
        recordType: BatchRecordType.assetInstance,
        id: `test`
      });
    }
    catch(err) {
      failed = true;
      assert.strictEqual(err.message, 'pop');
    }
    assert(failed);

  });

  it('times out a requests that are queued too long, when there is a batch in flight, and a batch queued', async () => {

    const processBatchCallback = sinon.stub().callsFake(() => delay(10));
    const processorCompleteCallback = sinon.stub();
    const p = new BatchProcessor(
      'author1',
      'type1',
      processBatchCallback,
      processorCompleteCallback,
      {
        batchMaxRecords: 1, // to trigger a batch immeidately
        addTimeoutMS: 5,
      }
    );

    await p.add({ id: `test-batch1-dispatched`, recordType: BatchRecordType.assetInstance  });

    // Make the persistence fail
    (database.upsertBatch as SinonStub).onSecondCall().callsFake(() => delay(10));

    let failed;
    try {
      await Promise.all([
        p.add({ id: `test-batch2-blocked`, recordType: BatchRecordType.assetInstance }),
        p.add({ id: `test-batch3-timeout`, recordType: BatchRecordType.assetInstance  }),
      ]);
    }
    catch(err) {
      failed = true;
      assert(err.message.includes('Timed out add of record after'));
    }
    assert(failed);

    // Clear everything out
    for (let i = 0; i < 100; i++) {
      if (processorCompleteCallback.callCount === 0) await delay(1);
    }
    // Confirm two batches went through
    assert.strictEqual(processBatchCallback.callCount, 2);

  });

  describe('with test wrapper', () => {

    class TestBatchProcessorWrapper extends BatchProcessor {
      public dispatchBatch() {
        return super.dispatchBatch();
      }
      public processBatch(batch: IDBBatch) {
        return super.processBatch(batch);
      }
      public newBatch(): IDBBatch {
        return super.newBatch();
      }
    }

    it('protects dispatchBatch against duplicate calls', async () => {
      const p = new TestBatchProcessorWrapper(
        'author1',
        'type1',
        sinon.stub(),
        sinon.stub(),
      );
      // p.assemblyBatch is not set, so this is a no-op and can be called many times
      await p.dispatchBatch();
      await p.dispatchBatch();
    });

    it('retries in processBatch, with backoff', async () => {
      const p = new TestBatchProcessorWrapper(
        'author1',
        'type1',
        sinon.stub()
          .onFirstCall().rejects(new Error('try me again'))
          .onSecondCall().rejects(new Error('and one more time with feeling')),
        sinon.stub(),
        {
          retryInitialDelayMS: 1,
        }
      );
      // p.assemblyBatch is not set, so this is a no-op and can be called many times
      await p.processBatch(p.newBatch());
    });

  });

});
};
