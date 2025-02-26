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

import { promisify } from 'util';
import { readFile } from 'fs';
import Ajv from 'ajv';
import configSchema from '../schemas/config.json';
import * as utils from './utils';
import { IConfig } from './interfaces';
import path from 'path';
import chokidar, { FSWatcher } from 'chokidar';

const log = utils.getLogger('lib/config.ts');

const asyncReadFile = promisify(readFile);
const ajv = new Ajv();
const validateConfig = ajv.compile(configSchema);
const configFilePath = path.join(utils.constants.DATA_DIRECTORY, utils.constants.CONFIG_FILE_NAME);

export let config: IConfig;
let fsWatcher: FSWatcher;
let configChangeListener: () => void;

export const init = async (listener: () => void) => {
  await loadConfigFile();
  watchConfigFile();
  configChangeListener = listener;
};

const loadConfigFile = async () => {
  try {
    const data = JSON.parse(await asyncReadFile(configFilePath, 'utf8'));
    if(validateConfig(data)) {
      config = data;
    } else {
      throw new Error('Invalid configuration file');
    }
  } catch(err) {
    throw new Error(`Failed to read configuration file. ${err}`);
  }
};

const watchConfigFile = () => {
  fsWatcher = chokidar.watch(configFilePath, { ignoreInitial: true }).on('change', async () => {
    try {
      await loadConfigFile();
      log.info('Loaded configuration file changes');
      configChangeListener();
    } catch(err) {
      log.error(`Failed to load configuration file. ${err}`);
    }
  });
};

export const shutDown = () => {
  fsWatcher.close();
};