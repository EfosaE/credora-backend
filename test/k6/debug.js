// Temporary debug script
import http from "k6/http";

export default function() {
  const res = http.post(
    "http://localhost:8080/api/v1/auth/login",
    JSON.stringify({
      account_number: "3058865182", // Use a real account number
      password: "!*1fTY**Ah6k",
    }),
    { headers: { "Content-Type": "application/json" } }
  );
  
  console.log("Status:", res.status);
  console.log("Body:", res.body);
  console.log("Cookies:", JSON.stringify(res.cookies));
}