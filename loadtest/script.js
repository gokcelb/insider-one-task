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
  scenarios: {
    sustained_load: {
      executor: 'ramping-arrival-rate',
      startRate: 100,
      timeUnit: '1s',
      preAllocatedVUs: 100,
      maxVUs: 500,
      stages: [
        { duration: '30s', target: 500 },
        { duration: '30s', target: 1000 },
        { duration: '1m', target: 2000 },
        { duration: '2m', target: 2000 },
        { duration: '30s', target: 500 },
      ],
    },
    spike_test: {
      executor: 'ramping-arrival-rate',
      startRate: 2000,
      timeUnit: '1s',
      preAllocatedVUs: 200,
      maxVUs: 1000,
      startTime: '4m30s', // Start after sustained load
      stages: [
        { duration: '10s', target: 10000 },
        { duration: '20s', target: 20000 },
        { duration: '10s', target: 2000 },
      ],
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<100', 'p(99)<200'],
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
    'response has status': (r) => {
      try {
        return r.json('status') === 'accepted';
      } catch {
        return false;
      }
    },
  });

  errorRate.add(!success);
}

export function handleSummary(data) {
  return {
    '/scripts/results.json': JSON.stringify(data, null, 2),
  };
}
