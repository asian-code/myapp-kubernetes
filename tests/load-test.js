// Load testing script for k6
// This script tests the three-tier application performance

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
export let errorRate = new Rate('errors');
export let responseTime = new Trend('response_time');

// Test configuration
export let options = {
  stages: [
    { duration: '2m', target: 10 }, // Ramp up
    { duration: '5m', target: 50 }, // Stay at 50 users
    { duration: '2m', target: 100 }, // Ramp up to 100 users
    { duration: '5m', target: 100 }, // Stay at 100 users
    { duration: '2m', target: 0 }, // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests must complete below 500ms
    errors: ['rate<0.1'], // Error rate must be below 10%
  },
};

const BASE_URL = __ENV.APP_URL || 'http://localhost';

export default function () {
  let response;

  // Test frontend (homepage)
  response = http.get(`${BASE_URL}/`);
  check(response, {
    'frontend status is 200': (r) => r.status === 200,
    'frontend response time < 500ms': (r) => r.timings.duration < 500,
  });
  errorRate.add(response.status !== 200);
  responseTime.add(response.timings.duration);

  sleep(1);

  // Test API health check
  response = http.get(`${BASE_URL}/api/health`);
  check(response, {
    'API health status is 200': (r) => r.status === 200,
    'API health response time < 200ms': (r) => r.timings.duration < 200,
  });
  errorRate.add(response.status !== 200);
  responseTime.add(response.timings.duration);

  sleep(1);

  // Test API endpoints
  response = http.get(`${BASE_URL}/api/users`);
  check(response, {
    'API users status is 200': (r) => r.status === 200,
    'API users response time < 1000ms': (r) => r.timings.duration < 1000,
  });
  errorRate.add(response.status !== 200);
  responseTime.add(response.timings.duration);

  sleep(2);

  // Test database connectivity through API
  response = http.get(`${BASE_URL}/api/products`);
  check(response, {
    'API products status is 200': (r) => r.status === 200,
    'API products response time < 1500ms': (r) => r.timings.duration < 1500,
  });
  errorRate.add(response.status !== 200);
  responseTime.add(response.timings.duration);

  sleep(2);
}

// Setup function - runs once before the test starts
export function setup() {
  console.log(`Starting load test against: ${BASE_URL}`);
  
  // Warmup request
  let response = http.get(`${BASE_URL}/api/health`);
  if (response.status !== 200) {
    console.warn(`Warmup request failed with status: ${response.status}`);
  }
}

// Teardown function - runs once after the test ends
export function teardown(data) {
  console.log('Load test completed');
}
