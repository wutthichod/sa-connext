import { NextRequest, NextResponse } from 'next/server';

const BACKEND_URL = 'http://localhost:8080';

export async function POST(request: NextRequest) {
  try {
    // 1️⃣ Read incoming body
    const body = await request.json();
    console.log('[api/join] incoming body:', body);

    // 2️⃣ Get token from Authorization header or cookie
    const authHeader = request.headers.get('authorization');
    const token = authHeader?.replace('Bearer ', '') || 
                  request.cookies.get('token')?.value;
    
    if (!token) {
      console.error('[api/join] missing token');
      return NextResponse.json({ error: 'Missing token' }, { status: 401 });
    }

    // 3️⃣ Map client body to backend expected format (backend will add user_id from JWT)
    const backendBody = {
      joining_code: body.joining_code 
    };

    console.log('[api/join] proxying to backend with body:', backendBody);
    const backendRes = await fetch(`${BACKEND_URL}/events/join`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}` // forward token
      },
      body: JSON.stringify(backendBody)
    });

    console.log('[api/join] Backend response status:', backendRes.status);

    const contentType = backendRes.headers.get('content-type') || '';
    let data: any;

    if (contentType.includes('application/json')) {
      data = await backendRes.json();
      console.log('[api/join] Backend response data:', JSON.stringify(data, null, 2));
    } else {
      const text = await backendRes.text();
      console.error('[api/join] backend returned non-JSON:', text);
      return NextResponse.json(
        { error: text || 'Backend returned non-JSON response' },
        { status: backendRes.status || 500 }
      );
    }

    if (!backendRes.ok) {
      console.error('[api/join] backend error:', {
        status: backendRes.status,
        data: data
      });
      return NextResponse.json(
        { error: data.message || data.error || 'Join failed' },
        { status: backendRes.status }
      );
    }

    console.log('[api/join] Join successful');
    return NextResponse.json(data);
  } catch (err: any) {
    console.error('[api/join] internal error:', err);
    return NextResponse.json(
      { error: err.message || 'Internal server error' },
      { status: 500 }
    );
  }
}