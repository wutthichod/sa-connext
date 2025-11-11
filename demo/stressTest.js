import http from "k6/http";
import ws from "k6/ws";
import { check, sleep } from "k6";
import { Rate, Counter, Trend } from "k6/metrics";
import { textSummary } from "https://jslib.k6.io/k6-summary/0.0.1/index.js";

// Custom metrics
const errorRate = new Rate("errors");
const wsConnections = new Counter("ws_connections");
const wsMessages = new Counter("ws_messages_sent");
const responseTime = new Trend("custom_response_time");

// Breaking point detection metrics
const stage100VUErrors = new Rate("stage_100vu_errors");
const stage200VUErrors = new Rate("stage_200vu_errors");
const stage300VUErrors = new Rate("stage_300vu_errors");
const stage400VUErrors = new Rate("stage_400vu_errors");
const stage500VUErrors = new Rate("stage_500vu_errors");

const stage100VUResponseTime = new Trend("stage_100vu_response_time");
const stage200VUResponseTime = new Trend("stage_200vu_response_time");
const stage300VUResponseTime = new Trend("stage_300vu_response_time");
const stage400VUResponseTime = new Trend("stage_400vu_response_time");
const stage500VUResponseTime = new Trend("stage_500vu_response_time");

// Stress Test Configuration - Development Testing (10-15 minutes)
// Goal: Find the breaking point of the system
export const options = {
  stages: [
    { duration: "2m", target: 100 }, // Baseline - normal load
    { duration: "2m", target: 200 }, // 2x normal load
    { duration: "2m", target: 300 }, // 3x normal load
    { duration: "2m", target: 400 }, // 4x normal load
    { duration: "2m", target: 500 }, // 5x normal load - stress
    { duration: "3m", target: 500 }, // Hold at maximum stress
    { duration: "2m", target: 0 }, // Recovery - ramp down
  ],
  thresholds: {
    // More lenient thresholds for stress testing
    http_req_duration: ["p(95)<2000", "p(99)<5000"], // Allow slower responses under stress
    http_req_failed: ["rate<0.15"], // Allow up to 15% failure rate
    errors: ["rate<0.20"], // Allow up to 20% custom error rate
  },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";

export function setup() {
  console.log("=".repeat(70));
  console.log("🔥 STRESS TEST - Finding System Breaking Point 🔥");
  console.log("=".repeat(70));
  console.log(`Target: ${BASE_URL}`);
  console.log(`Duration: 15 minutes`);
  console.log(`Strategy: Gradually increase load to find system limits`);
  console.log("");
  console.log("Test Stages:");
  console.log("  0-2m:   100 VUs  (Baseline - normal load)");
  console.log("  2-4m:   200 VUs  (2x normal - moderate stress)");
  console.log("  4-6m:   300 VUs  (3x normal - heavy stress)");
  console.log("  6-8m:   400 VUs  (4x normal - likely breaking point)");
  console.log("  8-10m:  500 VUs  (5x normal - extreme stress)");
  console.log("  10-13m: 500 VUs  (Hold at maximum - stability test)");
  console.log("  13-15m: 0 VUs    (Recovery - system cooldown)");
  console.log("");
  console.log("Expected Outcomes:");
  console.log("  ✅ 100-200 VUs: System performs well");
  console.log("  ⚠️  300-400 VUs: Response times increase, some errors");
  console.log("  ❌ 500 VUs: High error rates, system degradation");
  console.log("=".repeat(70));
  console.log("");

  return {
    testId: Date.now(),
    startTime: Date.now(),
  };
}

// VU state
let vuInitialized = false;
let vuEmail = "";
let vuPassword = "StressTest123!";
let authToken = "";
let chatIds = [];
let wsConnected = false;

// Helper function to determine current stage based on elapsed time
function getCurrentStage(startTime) {
  const elapsed = (Date.now() - startTime) / 1000; // seconds
  if (elapsed < 120)
    return {
      name: "100VU",
      metric: stage100VUErrors,
      timingMetric: stage100VUResponseTime,
    };
  if (elapsed < 240)
    return {
      name: "200VU",
      metric: stage200VUErrors,
      timingMetric: stage200VUResponseTime,
    };
  if (elapsed < 360)
    return {
      name: "300VU",
      metric: stage300VUErrors,
      timingMetric: stage300VUResponseTime,
    };
  if (elapsed < 480)
    return {
      name: "400VU",
      metric: stage400VUErrors,
      timingMetric: stage400VUResponseTime,
    };
  if (elapsed < 600)
    return {
      name: "500VU",
      metric: stage500VUErrors,
      timingMetric: stage500VUResponseTime,
    };
  if (elapsed < 780)
    return {
      name: "500VU_HOLD",
      metric: stage500VUErrors,
      timingMetric: stage500VUResponseTime,
    };
  return { name: "RECOVERY", metric: null, timingMetric: null };
}

export default function (data) {
  const headers = {
    "Content-Type": "application/json",
  };

  if (authToken) {
    headers["Cookie"] = `token=${authToken}`;
  }

  // Initialize VU once
  if (!vuInitialized) {
    if (!initializeVU(data, headers)) {
      return;
    }
  }

  // Fetch chats on first iteration if empty
  if (chatIds.length === 0) {
    fetchChats(headers);
  }

  // Aggressive traffic pattern for stress testing
  const rand = Math.random();

  // Prioritize chat creation to build up chat base
  if (chatIds.length < 3) {
    if (rand < 0.5) {
      // 50% chance to create chat when we have less than 3
      testCreateChat(headers);
      return;
    }
  }

  // Heavy traffic distribution
  if (rand < 0.25) {
    // 25% - Heavy read operations
    testGetEventsByUser(headers, data);
  } else if (rand < 0.4) {
    // 15% - Get user profile
    testGetUserProfile(headers, data);
  } else if (rand < 0.55) {
    // 15% - Create events (write load)
    testCreateEvent(headers, data);
  } else if (rand < 0.7) {
    // 15% - Chat creation
    testCreateChat(headers, data);
  } else if (rand < 0.85) {
    // 15% - Message sending
    testSendMessage(headers, data);
  } else if (rand < 0.95) {
    // 10% - Batch operations (heavy load)
    testHeavyBatchOperations(headers, data);
  } else {
    // 5% - WebSocket stress
    if (!wsConnected && chatIds.length > 0) {
      testWebSocketConnect();
    }
  }

  // Reduced sleep time for more aggressive testing
  sleep(Math.random() * 1 + 0.3); // 0.3-1.3 seconds (faster than normal load test)
}

function initializeVU(data, headers) {
  const vuId = `${data.testId}_${__VU}_${__ITER}`;
  vuEmail = `stress_${vuId}@example.com`;

  const registerPayload = JSON.stringify({
    username: `stress_${vuId}`,
    password: vuPassword,
    contact: {
      email: vuEmail,
      phone: `+66${String(__VU).padStart(9, "0")}`,
    },
    education: {
      university: "Stress Test University",
      major: "Performance Engineering",
    },
    jobTitle: "Load Tester",
    interests: ["stress-testing", "performance", "scalability"],
  });

  const registerRes = http.post(`${BASE_URL}/users/register`, registerPayload, {
    headers: headers,
    timeout: "30s", // Longer timeout for stress conditions
  });

  if (registerRes.status === 201) {
    const loginPayload = JSON.stringify({
      email: vuEmail,
      password: vuPassword,
    });

    const loginRes = http.post(`${BASE_URL}/users/login`, loginPayload, {
      headers: headers,
      timeout: "30s",
    });

    if (loginRes.status === 200) {
      const loginData = loginRes.json();
      authToken = loginData.jwtToken;
      vuInitialized = true;

      // Fetch existing chats
      headers["Cookie"] = `token=${authToken}`;
      fetchChats(headers);

      return true;
    }
  }

  errorRate.add(1);
  return false;
}

function fetchChats(headers) {
  const res = http.get(`${BASE_URL}/chats/`, {
    headers: headers,
    timeout: "15s",
  });

  if (res.status === 200) {
    try {
      const body = res.json();
      if (body && body.data && Array.isArray(body.data)) {
        chatIds = body.data
          .filter((chat) => chat.is_group === false)
          .map((chat) => chat.chat_id);
      }
    } catch (e) {
      // Ignore parsing errors
    }
  }
}

function testGetEventsByUser(headers, data) {
  const startTime = Date.now();
  const res = http.get(`${BASE_URL}/events/user`, {
    headers: headers,
    timeout: "15s",
  });
  const duration = Date.now() - startTime;
  responseTime.add(duration);

  const success = check(res, {
    "Get events: status 200": (r) => r.status === 200,
  });

  // Only check timing if request succeeded
  if (res.status === 200 && res.timings) {
    check(res, {
      "Get events: response time < 2000ms": (r) => r.timings.duration < 2000,
    });
  }

  errorRate.add(!success);

  // Track stage-specific metrics
  const stage = getCurrentStage(data.startTime);
  if (stage.metric) {
    stage.metric.add(!success);
    if (stage.timingMetric && res.timings) {
      stage.timingMetric.add(res.timings.duration);
    }
  }
}

function testGetUserProfile(headers, data) {
  const startTime = Date.now();
  const res = http.get(`${BASE_URL}/users/me`, {
    headers: headers,
    timeout: "15s",
  });
  const duration = Date.now() - startTime;
  responseTime.add(duration);

  const success = check(res, {
    "Get profile: status 200": (r) => r.status === 200,
  });

  // Only check timing if request succeeded
  if (res.status === 200 && res.timings) {
    check(res, {
      "Get profile: response time < 2000ms": (r) => r.timings.duration < 2000,
    });
  }

  errorRate.add(!success);

  // Track stage-specific metrics
  const stage = getCurrentStage(data.startTime);
  if (stage.metric) {
    stage.metric.add(!success);
    if (stage.timingMetric && res.timings) {
      stage.timingMetric.add(res.timings.duration);
    }
  }
}

function testCreateEvent(headers, data) {
  const payload = JSON.stringify({
    name: `Stress Event ${Date.now()}_VU${__VU}`,
    detail: "Testing system under high stress load",
    location: `StressTest_Location_${__VU}`,
    date: new Date(Date.now() + 86400000 * 7).toISOString(), // 1 week from now
    joining_code: `STRESS${Math.floor(Math.random() * 1000000)}`,
    organizer_id: "1",
  });

  const startTime = Date.now();
  const res = http.post(`${BASE_URL}/events/`, payload, {
    headers: headers,
    timeout: "20s",
  });
  const duration = Date.now() - startTime;
  responseTime.add(duration);

  const success = check(res, {
    "Create event: status 201": (r) => r.status === 201,
  });

  // Only check timing if request succeeded
  if (res.status === 201 && res.timings) {
    check(res, {
      "Create event: response time < 2000ms": (r) => r.timings.duration < 2000,
    });
  }

  errorRate.add(!success);

  // Track stage-specific metrics
  const stage = getCurrentStage(data.startTime);
  if (stage.metric) {
    stage.metric.add(!success);
    if (stage.timingMetric && res.timings) {
      stage.timingMetric.add(res.timings.duration);
    }
  }
}

function testCreateChat(headers, data) {
  // Use limited recipient pool (1-20) for higher success rate
  const recipientId = Math.floor(Math.random() * 20) + 1;
  const payload = JSON.stringify({
    recipient_id: recipientId.toString(),
  });

  const startTime = Date.now();
  const res = http.post(`${BASE_URL}/chats`, payload, {
    headers: headers,
    timeout: "20s",
  });
  const duration = Date.now() - startTime;
  responseTime.add(duration);

  const success = check(res, {
    "Create chat: status 200 or 201": (r) =>
      r.status === 200 || r.status === 201,
  });

  // Only check timing if request succeeded
  if ((res.status === 200 || res.status === 201) && res.timings) {
    check(res, {
      "Create chat: response time < 2000ms": (r) => r.timings.duration < 2000,
    });
  }

  errorRate.add(!success);

  // Track stage-specific metrics
  const stage = getCurrentStage(data.startTime);
  if (stage.metric) {
    stage.metric.add(!success);
    if (stage.timingMetric && res.timings) {
      stage.timingMetric.add(res.timings.duration);
    }
  }

  // Refresh chat list after successful creation
  if (success && res.status === 201) {
    const chatsRes = http.get(`${BASE_URL}/chats/`, {
      headers: headers,
      timeout: "15s",
    });

    if (chatsRes.status === 200) {
      try {
        const body = chatsRes.json();
        if (body && body.data && Array.isArray(body.data)) {
          chatIds = body.data
            .filter((chat) => chat.is_group === false)
            .map((chat) => chat.chat_id);
        }
      } catch (e) {
        // Ignore errors
      }
    }
  }
}

function testSendMessage(headers, data) {
  if (chatIds.length === 0) {
    // Skip if no chats available
    return;
  }

  const chatId = chatIds[Math.floor(Math.random() * chatIds.length)];
  const payload = JSON.stringify({
    chat_id: chatId,
    message: `Stress test message ${Date.now()} from VU ${__VU}`,
  });

  const startTime = Date.now();
  const res = http.post(`${BASE_URL}/chats/send`, payload, {
    headers: headers,
    timeout: "20s",
  });
  const duration = Date.now() - startTime;
  responseTime.add(duration);

  const success = check(res, {
    "Send message: status 200 or 201": (r) =>
      r.status === 200 || r.status === 201,
  });

  // Only check timing if request succeeded
  if ((res.status === 200 || res.status === 201) && res.timings) {
    check(res, {
      "Send message: response time < 2000ms": (r) => r.timings.duration < 2000,
    });
  }

  errorRate.add(!success);

  // Track stage-specific metrics
  const stage = getCurrentStage(data.startTime);
  if (stage.metric) {
    stage.metric.add(!success);
    if (stage.timingMetric && res.timings) {
      stage.timingMetric.add(res.timings.duration);
    }
  }
}

function testHeavyBatchOperations(headers, data) {
  // Fire multiple requests simultaneously to stress the system
  const startTime = Date.now();
  const batch = http.batch([
    [
      "GET",
      `${BASE_URL}/events/user`,
      null,
      { headers: headers, timeout: "15s" },
    ],
    ["GET", `${BASE_URL}/users/me`, null, { headers: headers, timeout: "15s" }],
    ["GET", `${BASE_URL}/chats/`, null, { headers: headers, timeout: "15s" }],
    ["GET", `${BASE_URL}/events/`, null, { headers: headers, timeout: "15s" }],
  ]);
  responseTime.add(Date.now() - startTime);

  const success = check(batch, {
    "Batch: all completed": (responses) => responses.length === 4,
    "Batch: at least 50% successful": (responses) => {
      const successCount = responses.filter(
        (r) => r.status >= 200 && r.status < 300
      ).length;
      return successCount >= 2;
    },
  });

  // Only check timing for successful requests
  const successfulResponses = batch.filter(
    (r) => r.status >= 200 && r.status < 300
  );
  if (successfulResponses.length > 0) {
    check(batch, {
      "Batch: successful requests < 2000ms": (responses) => {
        const successful = responses.filter(
          (r) => r.status >= 200 && r.status < 300
        );
        return successful.every((r) => r.timings && r.timings.duration < 2000);
      },
    });
  }

  errorRate.add(!success);
}

function testWebSocketConnect() {
  if (chatIds.length === 0) {
    return;
  }

  const chatId = chatIds[0];
  const wsUrl = `ws://localhost:8080/chats/ws`;
  const params = {
    headers: {
      Cookie: `token=${authToken}`,
    },
  };

  const res = ws.connect(wsUrl, params, function (socket) {
    wsConnections.add(1);
    wsConnected = true;

    socket.on("open", () => {
      // Send initial burst of messages
      for (let i = 0; i < 3; i++) {
        const message = JSON.stringify({
          type: "message",
          content: `Stress burst ${i} from VU ${__VU}`,
          chat_id: chatId,
        });
        socket.send(message);
        wsMessages.add(1);
        sleep(0.2);
      }
    });

    socket.on("message", (data) => {
      check(data, {
        "WS: received message": (d) => d !== null && d.length > 0,
      });
    });

    socket.on("error", (e) => {
      if (e.error) {
        const errorMsg = e.error();
        if (!errorMsg.includes("close sent")) {
          errorRate.add(1);
        }
      }
    });

    socket.on("close", () => {
      wsConnected = false;
    });

    // Keep connection open for shorter time under stress
    sleep(5);
    socket.close();
  });

  check(res, {
    "WS: connection successful": (r) => r && r.status === 101,
  });
}

export function handleSummary(data) {
  // Let k6 generate the default summary first
  return {
    stdout: textSummary(data, { indent: " ", enableColors: true }),
  };
}

export function teardown(data) {
  console.log("");
  console.log("=".repeat(70));
  console.log("🏁 STRESS TEST COMPLETE - BREAKING POINT ANALYSIS 🏁");
  console.log("=".repeat(70));
  console.log("");

  // Analyze each stage
  console.log("📊 PERFORMANCE BY LOAD STAGE:");
  console.log("");

  const stages = [
    {
      name: "100 VU (Baseline)",
      errorKey: "stage_100vu_errors",
      timeKey: "stage_100vu_response_time",
    },
    {
      name: "200 VU (2x Load)",
      errorKey: "stage_200vu_errors",
      timeKey: "stage_200vu_response_time",
    },
    {
      name: "300 VU (3x Load)",
      errorKey: "stage_300vu_errors",
      timeKey: "stage_300vu_response_time",
    },
    {
      name: "400 VU (4x Load)",
      errorKey: "stage_400vu_errors",
      timeKey: "stage_400vu_response_time",
    },
    {
      name: "500 VU (5x Load)",
      errorKey: "stage_500vu_errors",
      timeKey: "stage_500vu_response_time",
    },
  ];

  let breakingPoint = "Not reached - system stable";
  let recommendedMax = "500+ VU";

  stages.forEach((stage, index) => {
    const errorMetric = data.metrics[stage.errorKey];
    const timeMetric = data.metrics[stage.timeKey];

    if (!errorMetric || !timeMetric) {
      console.log(`${stage.name}: No data collected`);
      console.log("");
      return;
    }

    const errorRate = (errorMetric.values.rate * 100).toFixed(2);
    const p95 = timeMetric.values["p(95)"]?.toFixed(0) || "N/A";
    const p99 = timeMetric.values["p(99)"]?.toFixed(0) || "N/A";
    const avg = timeMetric.values.avg?.toFixed(0) || "N/A";

    let status = "✅ HEALTHY";
    let details = "System performing well";

    if (parseFloat(errorRate) > 15) {
      status = "❌ FAILING";
      details = "High error rate - system overloaded";
      if (breakingPoint === "Not reached - system stable") {
        breakingPoint = stage.name;
        recommendedMax = index > 0 ? stages[index - 1].name : "< 100 VU";
      }
    } else if (parseFloat(errorRate) > 5 || parseFloat(p95) > 2000) {
      status = "⚠️  DEGRADED";
      details = "Errors or slow responses detected";
      if (breakingPoint === "Not reached - system stable") {
        breakingPoint = `Approaching at ${stage.name}`;
        recommendedMax = stage.name;
      }
    } else if (parseFloat(p95) > 1000) {
      status = "⚡ STRESSED";
      details = "Response times elevated but acceptable";
    }

    console.log(`${stage.name}:`);
    console.log(`  Status:      ${status}`);
    console.log(`  Error Rate:  ${errorRate}%`);
    console.log(`  Avg Time:    ${avg}ms`);
    console.log(`  P95 Time:    ${p95}ms`);
    console.log(`  P99 Time:    ${p99}ms`);
    console.log(`  Assessment:  ${details}`);
    console.log("");
  });

  console.log("=".repeat(70));
  console.log("🎯 CONCLUSIONS:");
  console.log("=".repeat(70));
  console.log(`Breaking Point:        ${breakingPoint}`);
  console.log(`Recommended Max Load:  ${recommendedMax}`);
  console.log("");

  // Overall metrics
  const overallErrors = data.metrics.errors?.values.rate || 0;
  const overallP95 = data.metrics.http_req_duration?.values["p(95)"] || 0;

  console.log("💡 RECOMMENDATIONS:");
  if (overallErrors < 0.05 && overallP95 < 1000) {
    console.log("  ✅ System handled stress test excellently");
    console.log("  ✅ Consider testing with even higher loads");
    console.log("  ✅ Current infrastructure appears well-sized");
  } else if (overallErrors < 0.15 && overallP95 < 2000) {
    console.log("  ⚠️  System shows signs of stress at peak load");
    console.log("  ⚠️  Consider scaling before reaching production peak");
    console.log("  ⚠️  Monitor error rates and response times closely");
  } else {
    console.log("  ❌ System struggled under stress test conditions");
    console.log("  ❌ Immediate scaling/optimization recommended");
    console.log("  ❌ Review bottlenecks: database, CPU, memory, network");
    console.log("  ❌ Consider: connection pooling, caching, load balancing");
  }

  console.log("");
  console.log("=".repeat(70));
  console.log(
    `Test Duration: ${((Date.now() - data.startTime) / 1000 / 60).toFixed(
      1
    )} minutes`
  );
  console.log(`Test Completed: ${new Date().toISOString()}`);
  console.log("=".repeat(70));
}
