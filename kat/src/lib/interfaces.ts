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

// CONFIG INTERFACE

export interface IConfig {
  port: number
  assetTrailInstanceID: string
  protocol: 'ethereum' | 'corda'
  apiGateway: {
    apiEndpoint: string
    auth?: {
      user: string
      password: string
    }
  }
  eventStreams: {
    wsEndpoint: string
    topic: string
    auth?: {
      user: string
      password: string
    }
    skipSetup?: boolean
    config?: {
      batchSize: number
      batchTimeoutMS: number
      blockedRetryDelaySec: number,
      errorHandling: 'skip' | 'block'
    }
  }
  ipfs: {
    apiEndpoint: string
    gatewayEndpoint: string
  }
  app2app: {
    socketIOEndpoint: string
    destinations: {
      kat: string
      client: string
    }
  }
  docExchange: {
    apiEndpoint: string
    socketIOEndpoint: string
    destination: string
  }
  appCredentials: {
    user: string
    password: string
  }
  mongodb: {
    connectionUrl: string
    databaseName: string
  }
}

// SETTINGS

export interface ISettings {
  clientEvents: string[]
}

// STORED SUBSCRIPTION CONFIG

export interface IStoredSubscriptions {
  [sub: string]: string;
}

// API GATEWAY INTERFACES

export interface IAPIGatewayAsyncResponse {
  type: 'async'
  id: string
  msg: string
  sent: boolean
}

export interface IAPIGatewaySyncResponse {
  type: 'sync'
  transactionHash: string
}

export interface IAPIGatewaySyncResponseEthereum extends IAPIGatewaySyncResponse {
  protocol: 'ethereum'
  blockHash: string
  blockNumber: string
  cumulativeGasUsed: string
  from: string
  gasUsed: string
  headers: {
    id: string
    type: 'string',
    timeReceived: 'string',
    timeElapsed: number
    requestOffset: string
  }
  nonce: string
  status: string
  to: string
  transactionIndex: string
}

export interface IEventStream {
  id: string
}

export interface IEventStreamSubscription {
  name: string
  stream: string
}

// IPFS INTERFACES

export interface IIPFSAssetDefinition {
  assetDefinitionID: string
  name: string
  isContentPrivate: boolean
  isContentUnique: boolean
  descriptionSchema?: object
  contentSchema?: object
  indexes?: indexes
}

// IPFS INTERFACES

export interface IIPFSAssetDefinition {
  assetDefinitionID: string
  name: string
  isContentPrivate: boolean
  isContentUnique: boolean
  descriptionSchema?: object
  contentSchema?: object
  indexes?: indexes
}

// REQUEST INTERFACES

export interface IRequestMultiPartContent {
  author?: string
  assetDefinitionID?: string
  description?: Promise<string>
  contentStream: NodeJS.ReadableStream
  contentFileName: string
  isContentPrivate?: boolean
}

export interface IAssetDefinitionRequest {
  assetDefinitionID: string
  name: string
  author?: string
  isContentPrivate: boolean
  isContentUnique: boolean
  indexes?: indexes
  descriptionSchema?: object
  contentSchema?: object
}

// EVENT STREAM INTERFACES


interface IStateRefCorda {
  txhash: string,
  index: number
}

interface IStateCorda {
  data: object
}

export interface IEventStreamRawMessageCorda {
  data: IStateCorda,
  subId: string,
  signature: string,
  stateRef: IStateRefCorda,
  recordedTime: string,
  consumedTime: string
}

export interface IEventStreamMessage {
  address?: string
  blockNumber?: string
  transactionIndex?: string
  transactionHash: string
  data: object
  subId: string
  signature: string
  logIndex?: string
}

export interface IEventMemberRegistered {
  member: string
  name: string
  assetTrailInstanceID: string
  app2appDestination: string
  docExchangeDestination: string
  timestamp: number
}

export interface IEventAssetDefinitionCreated {
  author: string
  assetDefinitionHash: string
  timestamp: string
}

export interface IEventPaymentDefinitionCreated {
  paymentDefinitionID: string
  author: string
  name: string
  descriptionSchemaHash?: string
  timestamp: string
}

export interface IEventAssetInstanceCreated {
  assetInstanceID: string
  assetDefinitionID: string
  author: string
  descriptionHash?: string
  contentHash: string
  timestamp: string
  isContentPrivate: boolean
  participants?: string[]
}
export interface IEventAssetInstanceBatchCreated {
  batchHash: string;
  author: string
  timestamp: string
  participants?: string[]
}

export interface IEventPaymentInstanceCreated {
  paymentInstanceID: string
  paymentDefinitionID: string
  author: string
  member: string
  descriptionHash?: string
  amount: string
  timestamp: string
  participants?: string[]
}

export interface IEventAssetInstancePropertySet extends IAssetInstancePropertySet {
  timestamp: string
  participants?: string[]
}

// DATABASE INTERFACES

//TODO: figure out how to handle variable asset-instance collection names
export type databaseCollectionName = 'members' | 'asset-definitions' | 'payment-definitions' | 'payment-instances' | 'batches' | customCollectionName
export type customCollectionName = string

export type indexes = {fields: string[], unique?: boolean}[];

export interface IDatabaseProvider {
  init: () => Promise<void>
  createCollection: (collectionName: string, indexes: indexes) => Promise<void>
  count: (collectionName: databaseCollectionName, query: object) => Promise<number>
  find: <T>(collectionName: databaseCollectionName, query: object, sort: object, skip: number, limit: number) => Promise<T[]>
  findOne: <T>(collectionName: databaseCollectionName, query: object) => Promise<T | null>
  updateOne: (collectionName: databaseCollectionName, query: object, value: object, upsert: boolean) => Promise<void>
  shutDown: () => void
}

export interface IDBBlockchainData {
  blockNumber?: number
  transactionHash: string
  participants?: string[]
}

export interface IDBBlockchainPinned extends Partial<IDBBlockchainData> {
  submitted?: number
  timestamp?: number
  receipt?: string
}

export interface IDBMember extends IDBBlockchainPinned {
  _id?: string
  address: string
  name: string
  assetTrailInstanceID: string
  app2appDestination: string
  docExchangeDestination: string
}

export interface IDBAssetDefinition extends IIPFSAssetDefinition, IDBBlockchainPinned {
  _id?: string
  author: string
  assetDefinitionHash: string
  conflict?: boolean
}

export interface IDBPaymentDefinition extends IDBBlockchainPinned {
  _id?: string
  paymentDefinitionID: string
  author: string
  name: string
  descriptionSchema?: object
  descriptionSchemaHash?: string
  conflict?: boolean
}

export interface IAssetInstance {
  assetInstanceID: string
  assetDefinitionID: string
  author: string
  descriptionHash?: string
  description?: object
  content?: object
  contentHash?: string
  isContentPrivate: boolean
  conflict?: boolean
  filename?: string
  properties?: {
    [author: string]: {
      [key: string]: {
        value: string
        submitted?: number
        receipt?: string
        history?: {
          [timestamp: string]: {
            value: string
            timestamp: number
            blockNumber: number
            transactionHash: string
          }
        }
      } | undefined
    } | undefined
  }
}

export interface IDBAssetInstance extends IAssetInstance, IDBBlockchainPinned {
  _id?: string
  batchID?: string;
}

export interface IAssetInstancePropertySet {
  assetInstanceID: string
  assetDefinitionID: string
  author: string
  key: string
  value: string
}

export interface IDBPaymentInstance extends IDBBlockchainPinned {
  _id?: string
  paymentInstanceID: string
  paymentDefinitionID: string
  author: string
  member: string
  amount: number
  descriptionHash?: string
  description?: object
}

export enum BatchRecordType {
  assetInstance = 'instance',
  assetProperty = 'property',
}

export interface IBatchRecord {
  recordType: BatchRecordType,
  [x: string]: any, 
}

export interface IPinnedBatch {
  type: string;
  author: string;
  created: number;
  completed: number | null;
  batchID: string,
  records: IBatchRecord[];
}

export interface IDBBatch extends IPinnedBatch, IDBBlockchainPinned {
  _id?: string;
  batchHash?: string,
}

// APP2APP INTERFACES

export interface IApp2AppMessageHeader {
  from: string
  to: string
}

export interface IApp2AppMessage {
  headers: IApp2AppMessageHeader
  content: string
}

export interface IApp2AppMessageListener {
  (header: IApp2AppMessageHeader, content: AssetTradeMessage): void
}

// DOCUMENT EXCHANGE INTERFACES

export interface IDocExchangeDocumentDetails {
  name: string
  is_directory: boolean
  size: number
  hash: string
}

export interface IDocExchangeTransferData {
  transferId: string
  transferHash: string
  hash: string
  from: string
  to: string
  senderSignature: string
  memberSignature: string
  document: string
  timestamp: string
  status: 'sent' | 'received' | 'failed'
}

export interface IDocExchangeListener {
  (transferData: IDocExchangeTransferData): void
}

// ASSET TRADE INTERFACES

export type AssetTradeMessage =
  IAssetTradePrivateAssetInstanceRequest
  | IAssetTradePrivateAssetInstanceResponse
  | IAssetTradePrivateAssetInstancePush
  | IAssetTradePrivateAssetInstanceAuthorizationResponse

export interface IAssetTradePrivateAssetInstanceRequest {
  type: 'private-asset-instance-request'
  tradeID: string
  assetInstanceID: string
  assetDefinitionID: string
  requester: {
    assetTrailInstanceID: string
    address: string
  }
  metadata?: object
}

export interface IAssetTradePrivateAssetInstanceResponse {
  type: 'private-asset-instance-response'
  tradeID: string
  assetInstanceID: string
  rejection?: string
  content?: object
  filename?: string
}

export interface IAssetTradePrivateAssetInstancePush {
  type: 'private-asset-instance-push'
  assetInstanceID: string
  assetDefinitionID: string
  content?: object
  filename?: string
}

export interface IAssetTradePrivateAssetInstanceAuthorizationRequest {
  type: 'private-asset-instance-authorization-request'
  authorizationID: string
  assetInstance: IDBAssetInstance
  requester: IDBMember
  metadata?: object
}

export interface IAssetTradePrivateAssetInstanceAuthorizationResponse {
  type: 'private-asset-instance-authorization-response'
  authorizationID: string
  authorized: boolean
}

// CLIENT EVENT INTERFACES

export type ClientEventType =
  'member-registered'
  | 'asset-definition-submitted'
  | 'asset-definition-created'
  | 'asset-definition-name-conflict'
  | 'payment-definition-submitted'
  | 'payment-definition-created'
  | 'payment-definition-name-conflict'
  | 'asset-instance-submitted'
  | 'asset-instance-created'
  | 'asset-instance-content-conflict'
  | 'payment-instance-submitted'
  | 'payment-instance-created'
  | 'private-asset-instance-content-stored'
  | 'asset-instance-property-submitted'
  | 'asset-instance-property-set'

export interface IClientEventListener {
  (eventType: ClientEventType, content: object): void
}

export interface IPendingAssetInstancePrivateContentDelivery {
  assetInstanceID: string
  fromDestination: string
  content?: object
  filename?: string
}