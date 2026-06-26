export const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Thresholds shared across scripts
export const THRESHOLDS = {
  http_req_failed: [{ threshold: 'rate<0.01', abortOnFail: true }],
  http_req_duration: ['p(95)<500', 'p(99)<1000'],
};

// Reusable stage profiles
export const STAGES = {
  smoke: [
    { duration: '10s', target: 5 },
    { duration: '10s', target: 0 },
  ],
  load: [
    { duration: '30s', target: 50 },
    { duration: '1m', target: 50 },
    { duration: '15s', target: 0 },
  ],
  stress: [
    { duration: '30s', target: 100 },
    { duration: '1m', target: 200 },
    { duration: '30s', target: 0 },
  ],
};

export function selectStages() {
  const profile = (__ENV.PROFILE || 'smoke').toLowerCase();
  if (!STAGES[profile]) {
    throw new Error(`Unknown PROFILE="${profile}". Valid: smoke, load, stress`);
  }
  return STAGES[profile];
}
