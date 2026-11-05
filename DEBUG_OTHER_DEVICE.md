# คำแนะนำการ Debug จากอีกเครื่อง

## ขั้นตอนการตรวจสอบปัญหา "Failed to create chat"

### 1. ตรวจสอบการเชื่อมต่อ Backend

จากเครื่องที่มีปัญหา ให้เปิด Browser Console (กด F12):

```javascript
// ใน Console tab พิมพ์:
console.log("Backend URL:", process.env.NEXT_PUBLIC_BACKEND_URL);

// ลองเรียก API โดยตรง:
fetch("http://172.20.10.14:30080/health")
  .then((r) => r.json())
  .then((d) => console.log("Health check:", d))
  .catch((e) => console.error("Health check failed:", e));
```

### 2. ตรวจสอบ Token

```javascript
// ใน Console tab พิมพ์:
console.log(
  "Token:",
  localStorage.getItem("token") || sessionStorage.getItem("token")
);
```

### 3. ตรวจสอบ Network Request

1. เปิด DevTools (F12)
2. ไปที่ tab **Network**
3. Filter: `XHR` หรือ `Fetch`
4. ลองคลิกสร้าง chat อีกครั้ง
5. ดูที่ request ที่ไปหา `/api/chats`
6. คลิกดู:
   - **Headers**: ดู Request URL, Request Method, Status Code
   - **Response**: ดูข้อความ error
   - **Preview**: ดูข้อมูล response

### 4. ปัญหาที่พบบ่อย

#### ปัญหา: CORS Error

```
Access to fetch at 'http://172.20.10.14:30080/...' from origin 'http://172.20.10.14:3000'
has been blocked by CORS policy
```

**วิธีแก้:**

- ตรวจสอบว่า Backend (API Gateway) restart แล้วหลังจากแก้ CORS
- ดู logs ของ API Gateway: `kubectl logs -f deployment/api-gateway-deployment`

#### ปัญหา: 401 Unauthorized

```
{"error": "Unauthorized"}
```

**วิธีแก้:**

- Login ใหม่บนอีกเครื่อง
- ตรวจสอบว่า token ถูกเก็บใน localStorage/sessionStorage

#### ปัญหา: Connection Refused หรือ Network Error

```
Failed to fetch
net::ERR_CONNECTION_REFUSED
```

**วิธีแก้:**

1. ตรวจสอบว่าเครื่อง host เปิด Firewall แล้ว:

   ```powershell
   # บนเครื่อง host
   Get-NetFirewallRule -DisplayName "*30080*"
   Get-NetFirewallRule -DisplayName "*3000*"
   ```

2. ตรวจสอบว่า Backend ทำงานอยู่:

   ```powershell
   # บนเครื่อง host
   kubectl get pods
   kubectl get svc api-gateway-svc
   ```

3. ทดสอบเข้าถึง Backend จากอีกเครื่อง:
   ```bash
   # บนอีกเครื่อง (terminal หรือ browser)
   curl http://172.20.10.14:30080/health
   ```

#### ปัญหา: Request ไปที่ localhost แทน IP

ตรวจสอบใน Network tab ว่า Request URL เป็น:

- ❌ `http://localhost:30080/...` (ผิด - จะไม่ทำงานบนอีกเครื่อง)
- ✅ `http://172.20.10.14:30080/...` (ถูก)

**ถ้าผิด ให้:**

1. ตรวจสอบ `.env.local`:
   ```
   NEXT_PUBLIC_BACKEND_URL=http://172.20.10.14:30080
   ```
2. Restart Frontend (Ctrl+C และ `npm run dev`)
3. Hard refresh browser (Ctrl+Shift+R)

### 5. Debug Script สำหรับอีกเครื่อง

วาง script นี้ใน Browser Console:

```javascript
// Debug Helper
(async function debugConnection() {
  console.log("=== Connection Debug ===");

  // 1. Check environment
  console.log("1. Environment Variables:");
  console.log("   BACKEND_URL:", window.location.origin);

  // 2. Check token
  const token =
    localStorage.getItem("token") || sessionStorage.getItem("token");
  console.log(
    "2. Token:",
    token ? "Present (" + token.substring(0, 20) + "...)" : "Missing"
  );

  // 3. Test health endpoint
  console.log("3. Testing backend health...");
  try {
    const healthUrl = "http://172.20.10.14:30080/health";
    console.log("   Health URL:", healthUrl);
    const response = await fetch(healthUrl);
    const data = await response.json();
    console.log("   ✅ Health check success:", data);
  } catch (err) {
    console.error("   ❌ Health check failed:", err.message);
  }

  // 4. Test API endpoint
  if (token) {
    console.log("4. Testing chat API...");
    try {
      const apiUrl = "/api/chats";
      console.log("   API URL:", apiUrl);
      const response = await fetch(apiUrl, {
        method: "GET",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });
      console.log("   Status:", response.status);
      const data = await response.json();
      console.log("   ✅ API test success:", data);
    } catch (err) {
      console.error("   ❌ API test failed:", err.message);
    }
  } else {
    console.log("4. Skipping API test (no token)");
  }

  console.log("=== Debug Complete ===");
})();
```

### 6. ตรวจสอบ Backend Logs

บนเครื่อง host:

```powershell
# ดู logs แบบ real-time
kubectl logs -f deployment/api-gateway-deployment

# หรือดู logs ย้อนหลัง
kubectl logs deployment/api-gateway-deployment --tail=100
```

ดูว่ามี request เข้ามาหรือไม่ เมื่อลองสร้าง chat จากอีกเครื่อง

### 7. Quick Fix Checklist

- [ ] Backend (API Gateway) restart แล้วหลังแก้ CORS
- [ ] Firewall เปิด port 3000 และ 30080 แล้ว
- [ ] `.env.local` ใช้ IP ที่ถูกต้อง (172.20.10.14:30080)
- [ ] Frontend restart แล้วหลังแก้ `.env.local`
- [ ] Login ใหม่บนอีกเครื่อง
- [ ] Hard refresh browser (Ctrl+Shift+R)
- [ ] ทั้งสองเครื่องใช้ WiFi เดียวกัน
