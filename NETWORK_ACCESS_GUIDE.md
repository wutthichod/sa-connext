# ตรวจสอบและทดสอบการเข้าถึง API Gateway จาก WiFi Network

## ขั้นตอนการตั้งค่า

### 1. ตรวจสอบว่า Backend กำลังรันอยู่

```powershell
kubectl get pods
```

ต้องเห็น `api-gateway-deployment-xxxxx` ที่ status เป็น `Running`

### 2. ตรวจสอบ Service

```powershell
kubectl get svc api-gateway-svc
```

จะเห็น:

```
NAME               TYPE       CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
api-gateway-svc    NodePort   10.x.x.x        <none>        8080:30080/TCP   xxm
```

### 3. ทดสอบจากเครื่องเดียวกัน

```powershell
# ทดสอบด้วย localhost
curl http://localhost:30080/health

# หรือเปิด browser
Start-Process "http://localhost:30080/health"
```

### 4. ทดสอบด้วย IP Address

```powershell
# ทดสอบด้วย IP ของเครื่อง
curl http://172.20.10.14:30080/health
```

### 5. ทดสอบจากเครื่องอื่น (มือถือหรือเครื่องอื่นใน WiFi เดียวกัน)

เปิด browser แล้วเข้า:

```
http://172.20.10.14:30080/health
```

## URLs ที่ใช้

### จากเครื่องเดียวกัน:

- Frontend: `http://localhost:3000`
- Backend: `http://localhost:30080`

### จากเครื่องอื่นใน WiFi เดียวกัน:

- Frontend: `http://172.20.10.14:3000`
- Backend: `http://172.20.10.14:30080`
- WebSocket: `ws://172.20.10.14:30080/chats/ws/`

## Troubleshooting

### ถ้า NodePort ไม่เข้าถึงได้จากเครื่องอื่น

#### Windows Firewall

```powershell
# เปิด port 30080 สำหรับ incoming connections
New-NetFirewallRule -DisplayName "Kubernetes NodePort 30080" -Direction Inbound -Protocol TCP -LocalPort 30080 -Action Allow

# เปิด port 3000 สำหรับ Next.js
New-NetFirewallRule -DisplayName "Next.js Dev Server" -Direction Inbound -Protocol TCP -LocalPort 3000 -Action Allow
```

#### ตรวจสอบว่า Docker Desktop ใช้ network mode อะไร

```powershell
docker network ls
docker network inspect bridge
```

#### ตรวจสอบ Pod logs

```powershell
kubectl logs -f deployment/api-gateway-deployment
```

### ถ้า WebSocket ไม่ทำงาน

```powershell
# ตรวจสอบว่า upgrade protocol ทำงาน
curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" http://localhost:30080/chats/ws/
```

## การเริ่มต้นใช้งาน

### 1. Start Backend

```bash
cd c:\chula\P3_T1\SA\lastetProject\sa-connext
tilt up
```

### 2. Start Frontend

```bash
cd c:\chula\P3_T1\SA\lastetProject\sa-connext\_frontend
npm run dev
```

### 3. ตรวจสอบว่าทุกอย่างทำงาน

- เปิด http://localhost:3000 บนเครื่อง
- Login และทดสอบ features
- เปิด http://172.20.10.14:3000 จากมือถือ
- Login และทดสอบ real-time chat

## หมายเหตุสำคัญ

⚠️ **NodePort vs Tilt Port Forward:**

- Tilt port forward (8080) = เข้าถึงได้แค่ localhost
- NodePort (30080) = เข้าถึงได้จากทุก IP ใน network

⚠️ **Docker Desktop Kubernetes:**

- NodePort จะทำงานผ่าน Docker Desktop
- Port จะเปิดอัตโนมัติ แต่อาจโดน Firewall บล็อก
- ถ้าใช้ไม่ได้ ลอง restart Docker Desktop

⚠️ **IP Address เปลี่ยนได้:**

- เมื่อเชื่อมต่อ WiFi ใหม่
- เมื่อ restart router
- ต้องแก้ไข `.env.local` และ restart frontend ทุกครั้ง
