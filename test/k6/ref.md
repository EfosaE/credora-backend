scenarios: {
  spike: {
    executor: "ramping-arrival-rate",
    stages: [
      { target: 1, duration: "30s" },   // warm up at sustainable
      { target: 6, duration: "30s" },   // ramp hard above ceiling
      { target: 1, duration: "60s" },   // recover — does it drain cleanly?
    ],
  }
}
```

This tests whether your system *recovers* after overload, not just whether it rejects. The recovery tail is what actually matters in production — a transfer service that takes 3 minutes to drain after a 30-second spike is a problem even if the 503s were correct.

---

**Phase 3 — Soak test**

Run `rate=1` for `duration: "30m"`. You're looking for:
- Memory creep in your Go process
- Redis queue depth drifting upward over time (means sustainable rate is slightly optimistic)
- p95 latency slowly degrading (goroutine leak, connection pool exhaustion)

This is the test that catches things the 2-minute runs miss entirely.

---

**Phase 4 — Vary the config, not just the rate**

This is where your benchmark tracker pays off. Run the same `rate=2` stress test against different configs and compare rows:

| What to change | What it tells you |
|---|---|
| `safety_factor: 0.70` vs `0.85` | How conservative is the right margin? |
| `backlog_window_seconds: 10` vs `20` vs `40` | What queue depth is actually acceptable? |
| `concurrency_limit: 4` vs `8` vs `12` | Does IO-bound concurrency actually help throughput? |
| `replicas: 1` vs `2` | Do two replicas on the same machine help or just add overhead? |

---

## What to Record Each Run

For each run in your tracker, also note in the Label column:
```
rate=2 | safety=0.85 | concurrency=4 | replicas=2 | stress-phase