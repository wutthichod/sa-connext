// app/api/chats/users/route.ts
import { NextRequest, NextResponse } from "next/server";

const BACKEND_URL = "http://localhost:8080";

// GET - Get all online users
export async function GET(request: NextRequest) {
  try {
    const token =
      request.headers.get("authorization")?.replace("Bearer ", "") ||
      request.cookies.get("token")?.value;

    if (!token) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    const response = await fetch(`${BACKEND_URL}/chats/users`, {
      method: "GET",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
    });

    const contentType = response.headers.get("content-type") || "";
    let data: any;

    if (contentType.includes("application/json")) {
      data = await response.json();
    } else {
      const text = await response.text();
      console.error("[API] Backend returned non-JSON:", text);
      return NextResponse.json(
        { error: "Invalid response from server" },
        { status: 500 }
      );
    }

    if (!response.ok) {
      console.error("[API] Backend error:", data);
      return NextResponse.json(
        { error: data.message || data.error || "Failed to fetch online users" },
        { status: response.status }
      );
    }

    return NextResponse.json(data);
  } catch (error: any) {
    console.error("Get online users API error:", error);
    return NextResponse.json(
      { error: error.message || "Internal server error" },
      { status: 500 }
    );
  }
}
