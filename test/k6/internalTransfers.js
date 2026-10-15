import http from "k6/http";
import { check, sleep } from "k6";
import { SharedArray } from "k6/data";
import { uuidv4 } from "https://jslib.k6.io/k6-utils/1.4.0/index.js";

// Override at run time: k6 run -e BASE_URL=http://169.58.67.36:8080 internalTransfers.js
const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";

export const options = {
  stages: [
    { duration: "30s", target: 50 },
    { duration: "2m", target: 100 },
    { duration: "1m", target: 50 },
    { duration: "30s", target: 0 },
  ],
  thresholds: {
    http_req_duration: ["p(95)<20000"],
    http_req_failed: ["rate<0.01"],
    checks: ["rate>0.95"],
  },
  setupTimeout: "5m",
};

const users = new SharedArray("users", () => {
  const allUsers = JSON.parse(open("./k6_users_accounts.json"));
  return allUsers.slice(0, 100);
});

export function setup() {
  console.log(`Logging in ${users.length} users via ${BASE_URL}...`);

  const sessions = [];
  const batchSize = 50;

  for (let i = 0; i < users.length; i += batchSize) {
    const batch = users.slice(i, i + batchSize);

    const requests = batch.map((user) => ({
      method: "POST",
      url: `${BASE_URL}/api/v1/auth/login`,
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
      const accountNumber = batch[idx].account_number;
      const loginSuccess = check(res, {
        "login ok": (r) => r.status === 200,
      });

      if (!loginSuccess) {
        console.error(
          `Failed to login user: ${accountNumber}, status: ${res.status}, body: ${res.body}`,
        );
        return;
      }

      let accessToken = null;
      try {
        // ASSUMPTION: login returns { data: { accessToken } } (taken from your
        // existing script, not verified against the backend handler).
        accessToken = JSON.parse(res.body)?.data?.accessToken;
      } catch (e) {
        console.error(`Invalid JSON in login response for user: ${accountNumber}`);
        return;
      }

      if (accessToken) {
        sessions.push({ account_number: accountNumber, accessToken });
      } else {
        console.error(`No accessToken in response for user: ${accountNumber}`);
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
  if (!sessions || sessions.length < 2) {
    console.error("Not enough valid sessions to run transfers.");
    sleep(1);
    return;
  }

  const session = sessions[Math.floor(Math.random() * sessions.length)];

  let recipient;
  do {
    recipient = sessions[Math.floor(Math.random() * sessions.length)];
  } while (recipient.account_number === session.account_number);

  const amount = Math.floor(Math.random() * 500) + 50;

  const payload = JSON.stringify({
    toAccount: recipient.account_number,
    amount: amount.toString(), // ASSUMPTION: API expects amount as a string
    currency: "NGN",
    reference: `ref-${uuidv4()}`,
    description: `Load test transfer - ${amount} NGN`,
  });

  const res = http.post(`${BASE_URL}/api/v1/transfers/internal`, payload, {
    headers: {
      "Content-Type": "application/json",
      // ASSUMPTION: Bearer auth. Your old commented version used a "jwt" cookie;
      // I kept the Bearer version because that is the one active in your file.
      Authorization: `Bearer ${session.accessToken}`,
      "Idempotency-Key": uuidv4(),
    },
    tags: { name: "InternalTransfer" },
  });

  check(res, {
    "transfer accepted": (r) => r.status === 202,
    "not unauthorized": (r) => r.status !== 401,
    "has valid response": (r) => r.body && r.body.length > 0,
  });

  if (res.status !== 202) {
    console.error(
      `Transfer failed - Status: ${res.status}, Body: ${res.body}, Account: ${session.account_number}`,
    );
  }

  sleep(Math.random() * 2 + 1);
}

export function teardown(sessions) {
  console.log(`Test completed with ${sessions.length} user sessions`);
}