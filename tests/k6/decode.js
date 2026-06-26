/**
 * k6 load test — POST /decode
 *
 * Setup phase encodes a pool of URLs so the decode test hits real short codes.
 *
 * Usage:
 *   k6 run tests/k6/decode.js
 *   k6 run -e PROFILE=load -e POOL_SIZE=200 tests/k6/decode.js
 *   k6 run -e BASE_URL=http://staging:8080 -e PROFILE=stress tests/k6/decode.js
 */
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';
import { BASE_URL, THRESHOLDS, selectStages } from './config.js';

const decodeErrors = new Counter('decode_errors');
const decodeSuccess = new Rate('decode_success_rate');
const decodeDuration = new Trend('decode_duration', true);

const POOL_SIZE = parseInt(__ENV.POOL_SIZE || '50', 10);

export const options = {
  stages: selectStages(),
  thresholds: {
    ...THRESHOLDS,
    decode_success_rate: ['rate>0.99'],
    decode_duration: ['p(95)<400'],
  },
};

// Runs once before VUs start; return value is passed to default() and teardown().
export function setup() {
  const urls = [];
  for (let i = 0; i < POOL_SIZE; i++) {
    const original = `https://example.com/seed/${i}-${Date.now()}`;
    const res = http.post(
      `${BASE_URL}/encode`,
      JSON.stringify({ url: original }),
      { headers: { 'Content-Type': 'application/json' } },
    );
    if (res.status !== 200) {
      console.error(`Seed encode failed for index ${i}: ${res.status} ${res.body}`);
      continue;
    }
    urls.push(JSON.parse(res.body).short_url);
  }
  if (urls.length === 0) {
    throw new Error('Seeding failed: no short URLs were created. Is the server running?');
  }
  console.log(`Seeded ${urls.length} short URLs`);
  return { urls };
}

export default function (data) {
  const shortURL = data.urls[Math.floor(Math.random() * data.urls.length)];

  const res = http.post(
    `${BASE_URL}/decode`,
    JSON.stringify({ short_url: shortURL }),
    { headers: { 'Content-Type': 'application/json' } },
  );

  const ok = check(res, {
    'status is 200': (r) => r.status === 200,
    'body has url': (r) => {
      try {
        return JSON.parse(r.body).url !== undefined;
      } catch {
        return false;
      }
    },
    'url starts with https': (r) => {
      try {
        return JSON.parse(r.body).url.startsWith('https://');
      } catch {
        return false;
      }
    },
  });

  decodeDuration.add(res.timings.duration);
  decodeSuccess.add(ok);
  if (!ok) decodeErrors.add(1);

  sleep(0.1);
}
