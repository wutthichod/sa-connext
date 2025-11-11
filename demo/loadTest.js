import http from "k6/http";
import ws from "k6/ws";
import { check, sleep } from "k6";
import { Rate, Counter } from "k6/metrics";

// Custom metrics
const errorRate = new Rate("errors");
const wsConnections = new Counter("ws_connections");
const wsMessages = new Counter("ws_messages_sent");

// Test configuration for 100 requests/second
export const options = {
  stages: [
    { duration: "30s", target: 20 }, // Ramp up to 20 VUs
    { duration: "1m", target: 50 }, // Ramp up to 50 VUs
    { duration: "3m", target: 100 }, // Reach 100 VUs (≈100 req/s)
    { duration: "2m", target: 100 }, // Hold at 100 VUs
    { duration: "30s", target: 0 }, // Ramp down to 0
  ],
  thresholds: {
    http_req_duration: ["p(95)<500", "p(99)<1000"], // 95% < 500ms, 99% < 1s
    http_req_failed: ["rate<0.05"], // Error rate < 5%
    errors: ["rate<0.05"], // Custom error rate < 5%
  },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";

export function setup() {
  console.log(`Starting load test against ${BASE_URL}`);
  console.log("Target: 100 requests/second");
  console.log("Authentication: Cookie-based - Each VU creates its own user");

  // Just return the test timestamp - each VU will create its own user
  return {
    testId: Date.now(),
  };
}

// Each VU initializes once and creates its own user
let vuInitialized = false;
let vuEmail = "";
let vuPassword = "LoadTest123!";
let authToken = "";
let wsSocket = null; // WebSocket connection for this VU

export default function (data) {
  // Use auth token in Cookie header (mimicking browser behavior)
  const headers = {
    "Content-Type": "application/json",
  };

  if (authToken) {
    headers["Cookie"] = `token=${authToken}`;
  }

  // Each VU creates and logs in with its own unique user (only once)
  if (!vuInitialized) {
    // Create unique user for this VU
    const vuId = `${data.testId}_${__VU}_${__ITER}`;
    vuEmail = `loadtest_${vuId}@example.com`;

    const registerPayload = JSON.stringify({
      username: `loadtest_${vuId}`,
      password: vuPassword,
      contact: {
        email: vuEmail,
        phone: `+123456${__VU}`,
      },
      education: {
        university: "Test University",
        major: "Computer Science",
      },
      jobTitle: "QA Engineer",
      interests: ["testing", "performance"],
    });

    const registerRes = http.post(
      `${BASE_URL}/users/register`,
      registerPayload,
      {
        headers: headers,
      }
    );

    if (registerRes.status === 201) {
      // Registration sets cookie, but with Secure: true it won't work on HTTP
      // So we need to login separately (login has Secure: false)
      const loginPayload = JSON.stringify({
        email: vuEmail,
        password: vuPassword,
      });

      const loginRes = http.post(`${BASE_URL}/users/login`, loginPayload, {
        headers: headers,
      });

      if (loginRes.status === 200) {
        // Extract JWT token from response and use it in Cookie header
        const loginData = loginRes.json();
        authToken = loginData.jwtToken;
        vuInitialized = true;
      } else {
        console.error(
          `VU ${__VU}: Login failed - Status: ${
            loginRes.status
          }, Body: ${loginRes.body.substring(0, 100)}`
        );
        return;
      }
    } else {
      console.error(
        `VU ${__VU}: Registration failed - Status: ${
          registerRes.status
        }, Body: ${registerRes.body.substring(0, 100)}`
      );
      return;
    }
  } // Randomly choose an endpoint to test (realistic traffic distribution)
  const rand = Math.random();

  if (rand < 0.25) {
    // 25% - Get events by user (most common read operation)
    testGetEventsByUser(headers);
  } else if (rand < 0.45) {
    // 20% - Get user profile
    testGetUserProfile(headers);
  } else if (rand < 0.6) {
    // 15% - Get chats
    testGetChats(headers);
  } else if (rand < 0.75) {
    // 15% - Create event
    testCreateEvent(headers);
  } else if (rand < 0.85) {
    // 10% - Send message
    testSendMessage(headers);
  } else if (rand < 0.95) {
    // 10% - Mixed operations
    testMixedOperations(headers);
  } else {
    // 5% - WebSocket chat connection
    testWebSocketChat();
  }

  // Think time between requests (realistic user behavior)
  sleep(Math.random() * 2 + 0.5); // 0.5-2.5 seconds
}

function testGetEventsByUser(headers) {
  const res = http.get(`${BASE_URL}/events/user`, {
    headers: headers,
  });

  const success = check(res, {
    "Get events: status 200": (r) => r.status === 200,
    "Get events: response time < 500ms": (r) => r.timings.duration < 500,
  });

  errorRate.add(!success);
}

function testGetUserProfile(headers) {
  const res = http.get(`${BASE_URL}/users/me`, {
    headers: headers,
  });

  const success = check(res, {
    "Get profile: status 200": (r) => r.status === 200,
    "Get profile: has data": (r) => r.json("username") !== undefined,
  });

  errorRate.add(!success);
}

function testGetChats(headers) {
  const res = http.get(`${BASE_URL}/chats/`, {
    headers: headers,
  });

  const success = check(res, {
    "Get chats: status 200 or 404": (r) => r.status === 200 || r.status === 404,
    "Get chats: response time < 500ms": (r) => r.timings.duration < 500,
  });

  errorRate.add(!success);
}

function testCreateEvent(headers) {
  const payload = JSON.stringify({
    name: `Load Test Event ${Date.now()}`,
    detail: "Automated load testing event",
    location: "Virtual",
    date: new Date(Date.now() + 86400000).toISOString(), // Tomorrow
    joining_code: `TEST${Math.floor(Math.random() * 10000)}`,
    organizer_id: "1",
  });

  const res = http.post(`${BASE_URL}/events/`, payload, {
    headers: headers,
  });

  const success = check(res, {
    "Create event: status 201": (r) => r.status === 201,
    "Create event: has event ID": (r) => {
      const body = r.json();
      return body && body.data && body.data.event_id !== undefined;
    },
  });

  errorRate.add(!success);
}

function testSendMessage(headers) {
  const payload = JSON.stringify({
    recipient_id: "2",
    message: `Load test message at ${Date.now()}`,
  });

  const res = http.post(`${BASE_URL}/chats/send`, payload, {
    headers: headers,
  });

  const success = check(res, {
    "Send message: status 200 or 201": (r) =>
      r.status === 200 || r.status === 201,
  });

  errorRate.add(!success);
}

function testMixedOperations(headers) {
  // Batch of read operations - cookies automatically sent
  const batch = http.batch([
    ["GET", `${BASE_URL}/events/user`, null, { headers: headers }],
    ["GET", `${BASE_URL}/users/me`, null, { headers: headers }],
    ["GET", `${BASE_URL}/chats/`, null, { headers: headers }],
  ]);

  const success = check(batch, {
    "Batch: all requests completed": (responses) => responses.length === 3,
  });

  errorRate.add(!success);
}

function testWebSocketChat() {
  // Simulate a real WebSocket chat session
  const wsUrl = `ws://localhost:8080/chats/ws`;
  const params = {
    headers: {
      Cookie: `token=${authToken}`,
    },
  };

  const res = ws.connect(wsUrl, params, function (socket) {
    wsConnections.add(1);

    socket.on("open", () => {
      console.log(`VU ${__VU}: WebSocket connected`);

      // Send a chat message immediately after connecting
      const message = JSON.stringify({
        type: "message",
        content: `Load test chat from VU ${__VU} at ${Date.now()}`,
        recipient_id: "2",
      });

      socket.send(message);
      wsMessages.add(1);
    });

    socket.on("message", (data) => {
      check(data, {
        "WS: received message": (d) => d !== null && d.length > 0,
      });
    });

    socket.on("error", (e) => {
      if (e.error) {
        console.error(`VU ${__VU}: WebSocket error: ${e.error()}`);
        errorRate.add(1);
      }
    });

    socket.on("close", () => {
      console.log(`VU ${__VU}: WebSocket closed`);
    });

    // Stay connected for a short time (simulating real user behavior)
    // Send 2-3 messages during this time
    sleep(1);
    socket.send(
      JSON.stringify({
        type: "message",
        content: `Follow-up message from VU ${__VU}`,
        recipient_id: "2",
      })
    );
    wsMessages.add(1);

    sleep(1);
    socket.send(
      JSON.stringify({
        type: "typing",
        recipient_id: "2",
      })
    );

    sleep(1);
    // Close connection gracefully
    socket.close();
  });

  check(res, {
    "WS: connection successful": (r) => r && r.status === 101,
  });
}
export function teardown(data) {
  console.log("Load test completed!");
  console.log("Check the summary below for detailed metrics.");
}
