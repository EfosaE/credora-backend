import http from "k6/http";
import { check, sleep } from "k6";
import { SharedArray } from "k6/data";
import { uuidv4 } from "https://jslib.k6.io/k6-utils/1.4.0/index.js";

export const options = {
  stages: [
    { duration: "30s", target: 50 },
    { duration: "2m", target: 100 },
    // { duration: '3m', target: 100 },
    { duration: "1m", target: 200 },
    { duration: "2m", target: 200 },
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
  return allUsers.slice(0, 300);
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
        account_number: user.account_number,
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
        // Extract JWT from cookie
        const jwtToken =
          res.cookies.jwt && res.cookies.jwt[0]
            ? res.cookies.jwt[0].value
            : null;

        if (jwtToken) {
          sessions.push({
            account_number: batch[idx].account_number,
            jwt: jwtToken, // Store the JWT value
          });
        } else {
          console.error(
            `No JWT cookie found for user: ${batch[idx].account_number}`,
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
      },
      cookies: {
        jwt: session.jwt, // Send JWT as HttpOnly cookie
      },
      tags: { name: "InternalTransfer" },
    },
  );

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
