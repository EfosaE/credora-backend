import http from "k6/http";
import { check, sleep } from "k6";
import { SharedArray } from "k6/data";
import { uuidv4 } from "https://jslib.k6.io/k6-utils/1.4.0/index.js";

/*
|--------------------------------------------------------------------------
| System Capacity Constants
| Keep these in sync with config.yaml — they define what the system can
| actually handle. All thresholds and rates below are derived from these.
|--------------------------------------------------------------------------
|
| From config.yaml:
|   job.processing_rate_per_sec : 3.3975   (p95 service time T_s in seconds)
|   job.safety_factor           : 0.85
|   job.backlog_window_seconds  : 20
|   machine.cpu_cores           : 4
|   machine.job_bound           : "io"
|
| Derived (mirrors computeJobCapacity in config.go):
|   μ_max (hw ceiling)   = cpu_cores / T_s          = 4 / 3.3975 = 1.1775 jobs/sec
|   sustainable_per_sec  = μ_max * safety_factor     = 1.1775 * 0.85 = 1.0009 jobs/sec
|   rate_limit_per_min   = floor(sustainable * 60)   = 60 req/min = 1 req/sec
|   queue_max_size       = floor(μ_max * backlog_ws) = floor(1.1775 * 20) = 23 jobs
|
*/
const SYSTEM = {
  // The maximum rate the system can sustain without queue buildup.
  // This is your baseline "safe" test rate. Do not run production load above this.
  // Unit: requests per second
  SUSTAINABLE_RPS: 1.0,

  // Hard ceiling — what the hardware can theoretically process.
  // Running at or above this will cause queue buildup and 503s.
  // Use this only to intentionally stress-test backpressure behaviour.
  // Unit: requests per second
  HW_CEILING_RPS: 1.1775,

  // p95 service time — how long one job takes end-to-end.
  // Used to set realistic http_req_duration thresholds.
  // Unit: seconds
  SERVICE_TIME_P95_SEC: 3.3975, // NOTE this has been improved now at 1.4s

  // Maximum jobs that can queue before the system starts rejecting (503).
  // Mirrors job.queue_max_size from config.go.
  QUEUE_MAX_SIZE: 23,

  // How long a client should wait before retrying a 503.
  // Mirrors the Retry-After header value returned by BackpressureMiddleware.
  // Unit: seconds
  RETRY_AFTER_SEC: 20,
};

/*
|--------------------------------------------------------------------------
| Test Scenarios
|
| HOW TO PICK A RATE:
|   - Baseline (normal load)  : rate = SUSTAINABLE_RPS        (~1 rps)
|   - Stress (find the edge)  : rate = HW_CEILING_RPS * 1.5   (~2 rps)
|   - Spike (backpressure)    : rate = HW_CEILING_RPS * 3+    (~4 rps)
|
| CHANGE `rate` BELOW to switch between modes.
| Do NOT change preAllocatedVUs or maxVUs unless you change user pool size.
|--------------------------------------------------------------------------
*/
export const options = {
  scenarios: {
    transfers: {
      executor: "constant-arrival-rate",

      //  TWEAK THIS to change load level. See rates above.
      // 1  = baseline (sustainable, no queue buildup expected)
      // 2  = stress   (at ceiling, some 503s expected)
      // 4+ = spike    (above ceiling, backpressure kicks in, 503s guaranteed)
      rate: 2,

      timeUnit: "1s",

      // ✏️  TWEAK THIS to change how long the test runs.
      // Run for at least 2× the backlog_window_seconds to observe steady state.
      // Minimum recommended: "1m". Use "5m" for soak testing.
      duration: "3m",

      // Keep at 20 — enough VUs to sustain arrival rate without VU starvation.
      // Only increase if rate > 10 and you see "insufficient VUs" warnings.
      preAllocatedVUs: 100,
      maxVUs: 400,
    },
  },

  thresholds: {
    // p95 response time must stay under: backlog_window + T_s + 10% buffer
    // = (20 + 3.3975) * 1.1 * 1000ms ≈ 25,737ms → rounded to 26,000ms
    //  Loosen this (e.g. 40000) if testing spike scenarios where queue
    //     buildup is expected and you only care about backpressure, not latency.
    http_req_duration: [
      `p(95)<${Math.ceil((SYSTEM.RETRY_AFTER_SEC + SYSTEM.SERVICE_TIME_P95_SEC) * 1.1 * 1000)}`,
    ],

    // At sustainable rate: failure rate should be near 0.
    // At spike rate: expect 503s — loosen to e.g. "rate<0.5" when intentionally
    // testing backpressure so the test doesn't fail just because 503s are working.
    http_req_failed: ["rate<0.01"],

    // At least 95% of checks must pass (202 Accepted).
    //  Lower to "rate>0.5" when stress/spike testing backpressure intentionally.
    checks: ["rate>0.95"],
  },
};

/*
|--------------------------------------------------------------------------
| User Pool
| Reads from k6_users_accounts.json. Each entry must have:
|   { "account_number": "...", "password": "..." }
|
| Pool size is capped at 100. Increase the slice limit if your test
| duration is long and you want to avoid repeating the same accounts.
|--------------------------------------------------------------------------
*/
const users = new SharedArray("users", () => {
  const allUsers = JSON.parse(open("./k6_users_accounts.json"));
  //  Increase this cap if you have more seeded users and want more variety.
  return allUsers.slice(0, 500);
});

/*
|--------------------------------------------------------------------------
| Setup — Login all users and cache their tokens
| Runs once before the test. Returns a sessions array passed to default().
|--------------------------------------------------------------------------
*/
export function setup() {
  console.log(`Logging in ${users.length} users...`);

  const sessions = [];

  // Login in batches to avoid hammering the auth endpoint serially.
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
      },
    }));

    const responses = http.batch(requests);

    responses.forEach((res, idx) => {
      if (res.status === 200) {
        const body = JSON.parse(res.body);
        sessions.push({
          account_number: batch[idx].account_number,
          accessToken: body?.data?.accessToken,
        });
      } else {
        console.warn(
          `Login failed for ${batch[idx].account_number} — status: ${res.status}`,
        );
      }
    });
  }

  console.log(
    `Setup complete: ${sessions.length}/${users.length} sessions created`,
  );

  if (sessions.length === 0) {
    throw new Error(
      "No sessions created — check auth endpoint and user seed data",
    );
  }

  return sessions;
}

export default function (sessions) {
  // Pick a random sender
  const senderIndex = Math.floor(Math.random() * sessions.length);
  const sender = sessions[senderIndex];

  // Pick a recipient efficiently (avoid the sender)
  // Trick: pick a random index in 0..n-2, then map it to the real array
  const n = sessions.length;
  const recipientIndex = Math.floor(Math.random() * (n - 1));
  const recipient =
    recipientIndex >= senderIndex
      ? sessions[recipientIndex + 1]
      : sessions[recipientIndex];

  // Random amount (100–549 NGN)
  const amount = Math.floor(Math.random() * 450) + 100;

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
        Authorization: `Bearer ${sender.accessToken}`,
        "Idempotency-Key": uuidv4(),
      },
      timeout: `${(SYSTEM.RETRY_AFTER_SEC + SYSTEM.SERVICE_TIME_P95_SEC + 5) * 1000}ms`,
    },
  );

  const accepted = check(res, {
    "transfer accepted (202)": (r) => r.status === 202,
  });

  if (!accepted) {
    if (res.status === 503) {
      const retryAfter = res.headers["Retry-After"];
      const resetAt = res.headers["X-RateLimit-Reset"];
      console.warn(
        `[BACKPRESSURE] 503 — Retry-After: ${retryAfter}s, X-RateLimit-Reset: ${resetAt} | sender: ${sender.account_number}`,
      );
    } else if (res.status === 429) {
      console.warn(
        `[RATE LIMITED] 429 — arrival rate exceeds ${SYSTEM.SUSTAINABLE_RPS} rps | sender: ${sender.account_number}`,
      );
    } else {
      console.error(
        `[FAILED] status: ${res.status} | body: ${res.body} | sender: ${sender.account_number}`,
      );
    }
  }
}

/*
|--------------------------------------------------------------------------
| Teardown — Runs once after the test completes
|--------------------------------------------------------------------------
*/
export function teardown() {
  console.log(
    `Test complete.\n` +
      `Review results against system capacity:\n` +
      `  Sustainable RPS : ${SYSTEM.SUSTAINABLE_RPS}\n` +
      `  HW Ceiling RPS  : ${SYSTEM.HW_CEILING_RPS}\n` +
      `  Queue Max Size  : ${SYSTEM.QUEUE_MAX_SIZE}\n` +
      `  p95 Service Time: ${SYSTEM.SERVICE_TIME_P95_SEC}s`,
  );
}
