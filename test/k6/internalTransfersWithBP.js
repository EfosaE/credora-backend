import http from "k6/http";
import { check, sleep } from "k6";
import { Counter, Trend, Rate } from "k6/metrics";
import { SharedArray } from "k6/data";
import { uuidv4 } from "https://jslib.k6.io/k6-utils/1.4.0/index.js";

// Custom metrics for backpressure monitoring
const backpressureHits = new Counter("backpressure_hits");
const queueRejections = new Counter("queue_rejections");
const queueSize = new Trend("queue_size");
const backpressureRate = new Rate("backpressure_rate");

export const options = {
  scenarios: {
    default_load: {
      executor: "ramping-vus",
      stages: [
        { duration: "30s", target: 20 },
        { duration: "2m", target: 50 },
        { duration: "1m", target: 100 },
        { duration: "2m", target: 100 },
        { duration: "30s", target: 0 },
      ],
      exec: "default", // points to your default function
    },
    backpressure_load: {
      executor: "ramping-arrival-rate",
      startTime: "6.5m",
      startRate: 1,
      timeUnit: "1s",
      preAllocatedVUs: 200,
      maxVUs: 300,
      stages: [
        { duration: "30s", target: 50 },
        { duration: "1m", target: 100 },
        { duration: "1m", target: 200 },
        { duration: "30s", target: 0 },
      ],
      exec: "backpressureTest", // points to your backpressureTest function
    },
  },
  thresholds: {
    http_req_duration: ["p(95)<20000"],
    http_req_failed: ["rate<0.01"],
    checks: ["rate>0.90"],
    backpressure_rate: ["rate<0.15"],
    queue_size: ["p(95)<20"],
  },
  setupTimeout: "5m",
};

const users = new SharedArray("users", () => {
  const allUsers = JSON.parse(open("./k6_users_accounts.json"));
  return allUsers.slice(0, 200);
});

export function setup() {
  console.log(`Logging in ${users.length} users...`);

  const sessions = [];
  const batchSize = 50;

  for (let i = 0; i < users.length; i += batchSize) {
    const batch = users.slice(i, i + batchSize);

    const requests = batch.map((user) => ({
      method: "POST",
      url: "http://localhost:8080/api/v1/auth/login",
      body: JSON.stringify({
        accountNumber: user.account_number,
        password: user.password,
      }),
      params: {
        headers: { "Content-Type": "application/json" },
        tags: { name: "Login" },
      },
    }));

    const responses = http.batch(requests);

    responses.forEach((res, idx) => {
      const loginSuccess = check(res, {
        "login ok": (r) => r.status === 200,
      });

      if (loginSuccess) {
        try {
          const body = JSON.parse(res.body);
          const accessToken = body.data?.accessToken;

          if (accessToken) {
            sessions.push({
              account_number: batch[idx].account_number,
              jwt: accessToken,
            });
          } else {
            console.error(
              `No accessToken in response for user: ${batch[idx].account_number}`,
            );
          }
        } catch (e) {
          console.error(
            `Failed to parse login response for user: ${batch[idx].account_number}`,
          );
        }
      } else {
        console.error(
          `Failed to login user: ${batch[idx].account_number}, status: ${res.status}`,
        );
      }
    });

    console.log(
      `Logged in ${Math.min(i + batchSize, users.length)}/${users.length} users`,
    );
  }

  console.log(`Setup complete: ${sessions.length} sessions created`);
  return sessions;
}

export default function (sessions) {
  const session = sessions[Math.floor(Math.random() * sessions.length)];

  // Check queue health before attempting transfer
  const healthRes = http.get("http://localhost:8080/api/v1/health/readiness", {
    tags: { name: "HealthCheck" },
  });

  if (healthRes.status === 200) {
    try {
      const body = JSON.parse(healthRes.body);
      const queueData = body.data;
      if (queueData) {
        queueSize.add(queueData.pending_jobs);
        if (queueData.status === "degraded") {
          console.warn(
            `⚠️ Queue degraded — pending: ${queueData.pending_jobs}, active: ${queueData.active_jobs}, capacity: ${queueData.queue_capacity}`,
          );
        }
      }
    } catch (e) {
      // Health check parsing failed, continue anyway
    }
  }

  let recipient;
  do {
    recipient = sessions[Math.floor(Math.random() * sessions.length)];
  } while (recipient.account_number === session.account_number);

  const amount = Math.floor(Math.random() * 500) + 50;

  const payload = JSON.stringify({
    toAccount: recipient.account_number,
    amount: amount.toString(),
    currency: "NGN",
    reference: `ref-${uuidv4()}`,
    description: `Load test transfer - ${amount} NGN`,
  });

  const res = http.post(
    "http://localhost:8080/api/v1/transfers/internal",
    payload,
    {
      headers: {
        "Content-Type": "application/json",
        "Idempotency-Key": uuidv4(),
        Authorization: `Bearer ${session.jwt}`, // changed from cookie to Authorization header
      },
      tags: { name: "InternalTransfer" },
    },
  );

  const checks = check(res, {
    "transfer accepted (202)": (r) => r.status === 202,
    "backpressure triggered (503)": (r) => r.status === 503,
    "not unauthorized": (r) => r.status !== 401,
    "has valid response": (r) => r.body && r.body.length > 0,
    "retry-after header present on 503": (r) =>
      r.status !== 503 || r.headers["Retry-After"] !== undefined,
  });
  

  // Track backpressure metrics
  if (res.status === 503) {
    backpressureHits.add(1);
    backpressureRate.add(1);
    queueRejections.add(1);

    const retryAfter = res.headers["Retry-After"] || "not set";
    console.log(`🚦 Backpressure: Queue full - Retry-After: ${retryAfter}s`);

    if (res.headers["Retry-After"]) {
      const waitTime = parseInt(res.headers["Retry-After"]);
      sleep(waitTime);
    } else {
      sleep(5);
    }
  } else if (res.status === 202) {
    backpressureRate.add(0);
  } else {
    console.error(
      `Transfer failed - Status: ${res.status}, Body: ${res.body}, Account: ${session.account_number}`,
    );
  }

  sleep(Math.random() * 8 + 10); // 10-18 seconds between requests
}

export function teardown(sessions) {
  console.log(`\n=== Test Summary ===`);
  console.log(`Total user sessions: ${sessions.length}`);
  console.log(
    `Backpressure was tested at a realistic rate and monitoring metrics collected`,
  );

  const finalHealth = http.get("http://localhost:8080/health/readiness");
  if (finalHealth.status === 200) {
    try {
      const body = JSON.parse(finalHealth.body);
      console.log(`Final queue state: ${JSON.stringify(body.data, null, 2)}`);
    } catch (e) {
      console.log(`Health check response: ${finalHealth.body}`);
    }
  }
}

export function backpressureTest(sessions) {
  const session = sessions[Math.floor(Math.random() * sessions.length)];

  let recipient;
  do {
    recipient = sessions[Math.floor(Math.random() * sessions.length)];
  } while (recipient.account_number === session.account_number);

  const amount = Math.floor(Math.random() * 500) + 50;

  const payload = JSON.stringify({
    toAccount: recipient.account_number,
    amount: amount.toString(),
    currency: "NGN",
    reference: `ref-${uuidv4()}`,
    description: `Backpressure test - ${amount} NGN`,
  });

  const res = http.post(
    "http://localhost:8080/api/v1/transfers/internal",
    payload,
    {
      headers: {
        "Content-Type": "application/json",
        "Idempotency-Key": uuidv4(),
        Authorization: `Bearer ${session.jwt}`, // changed from cookie to Authorization header
      },
      tags: { name: "BackpressureTest" },
    },
  );

  check(res, {
    "transfer accepted or rejected gracefully": (r) =>
      r.status === 202 || r.status === 503,
  });

  if (res.status === 503) {
    backpressureHits.add(1);
    sleep(0.5);
  } else {
    sleep(0.1);
  }
}
