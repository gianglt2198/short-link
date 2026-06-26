/**
 * k6 load test — POST /encode
 *
 * Usage:
 *   k6 run tests/k6/encode.js
 *   k6 run -e PROFILE=load tests/k6/encode.js
 *   k6 run -e BASE_URL=http://staging:8080 -e PROFILE=stress tests/k6/encode.js
 */
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';
import { BASE_URL, THRESHOLDS, selectStages } from './config.js';

const encodeErrors = new Counter('encode_errors');
const encodeSuccess = new Rate('encode_success_rate');
const encodeDuration = new Trend('encode_duration', true);

export const options = {
  stages: selectStages(),
  thresholds: {
    ...THRESHOLDS,
    encode_success_rate: ['rate>0.99'],
    encode_duration: ['p(95)<400'],
  },
};

// Unique URL per VU+iteration avoids hitting the idempotency cache so we
// actually exercise the full encode path on every request.
export default function () {
  const url = `https://example.com/load-test/${__VU}-${__ITER}-${Date.now()}`;

  const res = http.post(
    `${BASE_URL}/encode`,
    JSON.stringify({ url }),
    { headers: { 'Content-Type': 'application/json' } },
  );

  const ok = check(res, {
    'status is 200': (r) => r.status === 200,
    'body has short_url': (r) => {
      try {
        return JSON.parse(r.body).short_url !== undefined;
      } catch {
        return false;
      }
    },
    'short_url starts with base url': (r) => {
      try {
        return JSON.parse(r.body).short_url.startsWith(BASE_URL);
      } catch {
        return false;
      }
    },
  });

  encodeDuration.add(res.timings.duration);
  encodeSuccess.add(ok);
  if (!ok) encodeErrors.add(1);

  sleep(0.1);
}
