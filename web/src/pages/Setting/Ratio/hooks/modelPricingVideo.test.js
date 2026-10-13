/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import { test } from 'node:test';
import assert from 'node:assert/strict';
import {
  readVideoPricing,
  serializeVideoPricing,
} from './modelPricingVideo.js';

const t = (value) => value;

test('video prices preserve explicit zero and false on save and reload', () => {
  const pricing = {
    price_per_second: 0,
    tiers: [
      { resolution: '1080p', generate_audio: false, price_per_second: 0.1 },
    ],
  };
  assert.deepEqual(
    serializeVideoPricing(readVideoPricing(pricing), t),
    pricing,
  );
});

test('invalid and duplicate video tiers cannot be saved', () => {
  const model = { videoPrice: '0.1', videoTiers: [] };
  for (const value of ['', '-1', 'NaN', 'Infinity']) {
    assert.throws(() =>
      serializeVideoPricing({ ...model, videoPrice: value }, t),
    );
  }
  assert.throws(() =>
    serializeVideoPricing(
      {
        ...model,
        videoTiers: [{ resolution: '', audio: 'any', price: '0.1' }],
      },
      t,
    ),
  );
  assert.throws(() =>
    serializeVideoPricing(
      {
        ...model,
        videoTiers: [
          { resolution: '1080P', audio: 'false', price: '0.1' },
          { resolution: '1080p', audio: 'false', price: '0.2' },
        ],
      },
      t,
    ),
  );
});

test('dimension tiers normalize before saving and duplicate validation', () => {
  const model = {
    videoPrice: '0.1',
    videoTiers: [{ resolution: '1920*1080', audio: 'any', price: '0.2' }],
  };
  assert.equal(serializeVideoPricing(model, t).tiers[0].resolution, '1080p');
  assert.throws(() =>
    serializeVideoPricing(
      {
        ...model,
        videoTiers: [
          ...model.videoTiers,
          { resolution: '1080p', audio: 'any', price: '0.3' },
        ],
      },
      t,
    ),
  );
});
