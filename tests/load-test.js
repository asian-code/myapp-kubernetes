import http from 'k6/http';
import { check, sleep } from 'k6';

// Basic load profile: ramp up, sustain, ramp down
export const options = {
  stages: [
    { duration: '30s', target: 20 },
    { duration: '90s', target: 100 },
    { duration: '20s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(99)<1000'],
    http_req_failed: ['rate<0.01'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const TOKEN = __ENV.API_TOKEN || '';

export default function () {
  const res = http.get(`${BASE_URL}/api/v1/dashboard`, {
    headers: TOKEN ? { Authorization: `Bearer ${TOKEN}` } : {},
  });

  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 1000ms': (r) => r.timings.duration < 1000,
  });

  sleep(1);
}
