import http from 'k6/http';
import { check } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

const TARGET_HOST = __ENV.TARGET_HOST || 'localhost:8080';
const BASE_URL = `http://${TARGET_HOST}`;

const channels = ['web', 'mobile', 'api', 'email', 'push'];
const eventNames = [
  'page_view',
  'product_view',
  'add_to_cart',
  'checkout_start',
  'purchase',
  'search',
  'login',
  'signup',
  'share',
  'favorite'
];

export const options = {
    summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)'],
    scenarios: {
    ingestion_test: {
      executor: 'ramping-arrival-rate',
      startRate: 0,
      timeUnit: '1s',
      preAllocatedVUs: 1000,
      maxVUs: 5000,
      stages: [
        { duration: '1m', target: 2000 },
        { duration: '3m', target: 2000 },
        { duration: '15s', target: 20000 },
        { duration: '30s', target: 20000 },
        { duration: '30s', target: 2000 },
        { duration: '2m', target: 2000 },
        { duration: '30s', target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<60', 'p(99)<150'],
    errors: ['rate<0.01'],
  },
};

function generateEvent() {
  const now = Math.floor(Date.now() / 1000);
  const userId = `user_${Math.floor(Math.random() * 100000)}`;
  const eventName = eventNames[Math.floor(Math.random() * eventNames.length)];
  const channel = channels[Math.floor(Math.random() * channels.length)];

  return {
    event_name: eventName,
    channel: channel,
    campaign_id: `campaign_${Math.floor(Math.random() * 100)}`,
    user_id: userId,
    timestamp: now,
    tags: [`source:loadtest`, `env:test`],
    metadata: {
      session_id: `session_${Math.random().toString(36).slice(2, 11)}`,
      page_url: '/products/item-' + Math.floor(Math.random() * 1000),
      referrer: 'https://example.com',
    },
  };
}

const params = {
  headers: {
    'Content-Type': 'application/json',
  },
};

export default function () {
  const event = generateEvent();
  const payload = JSON.stringify(event);

  const res = http.post(`${BASE_URL}/events`, payload, params);

  const success = check(res, {
    'status is 202': (r) => r.status === 202,
  });

  errorRate.add(!success);
}

export function handleSummary(data) {
  const outputPath = __ENV.RESULTS_PATH || 'loadtest/results.json';
  return {
    [outputPath]: JSON.stringify(data, null, 2),
  };
}
