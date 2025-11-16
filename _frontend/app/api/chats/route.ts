// app/api/chats/route.ts
import { NextRequest, NextResponse } from 'next/server';

const BACKEND_URL = 'http://localhost:8080';

// POST - Create a new chat
export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    console.log('[API] Create chat request body:', JSON.stringify(body, null, 2));
    
    const token = request.headers.get('authorization')?.replace('Bearer ', '') || 
                  request.cookies.get('token')?.value;
    
    if (!token) {
      console.error('[API] No token provided');
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    if (!body.recipient_id) {
      console.error('[API] Missing recipient_id');
      return NextResponse.json(
        { error: 'recipient_id is required' },
        { status: 400 }
      );
    }

    console.log('[API] Calling backend:', `${BACKEND_URL}/chats`);
    const response = await fetch(`${BACKEND_URL}/chats`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        recipient_id: body.recipient_id,
      }),
    });

    console.log('[API] Backend response status:', response.status);

    const contentType = response.headers.get('content-type') || '';
    let data: any;

    if (contentType.includes('application/json')) {
      data = await response.json();
      console.log('[API] Backend response data:', JSON.stringify(data, null, 2));
    } else {
      const text = await response.text();
      console.error('[API] Backend returned non-JSON:', text);
      return NextResponse.json(
        { error: 'Invalid response from server' },
        { status: 500 }
      );
    }

    if (!response.ok) {
      console.error('[API] Backend error response:', {
        status: response.status,
        data: data
      });
      return NextResponse.json(
        { error: data.message || data.error || 'Failed to create chat' },
        { status: response.status }
      );
    }

    console.log('[API] Create chat successful');
    return NextResponse.json(data);
  } catch (error: any) {
    console.error('Create chat API error:', error);
    return NextResponse.json(
      { error: error.message || 'Internal server error' },
      { status: 500 }
    );
  }
}

// GET - Get all chats for the current user
export async function GET(request: NextRequest) {
  try {
    const token = request.headers.get('authorization')?.replace('Bearer ', '') || 
                  request.cookies.get('token')?.value;
    
    if (!token) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    const response = await fetch(`${BACKEND_URL}/chats`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
    });

    const contentType = response.headers.get('content-type') || '';
    let data: any;

    if (contentType.includes('application/json')) {
      data = await response.json();
    } else {
      const text = await response.text();
      return NextResponse.json(
        { error: 'Invalid response from server' },
        { status: 500 }
      );
    }

    if (!response.ok) {
      return NextResponse.json(
        { error: data.message || data.error || 'Failed to fetch chats' },
        { status: response.status }
      );
    }

    return NextResponse.json(data);
  } catch (error: any) {
    console.error('Get chats API error:', error);
    return NextResponse.json(
      { error: error.message || 'Internal server error' },
      { status: 500 }
    );
  }
}

