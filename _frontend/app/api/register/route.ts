// app/api/register/route.ts
import { NextRequest, NextResponse } from "next/server";
import { BACKEND_URL } from "@/app/config";

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();

    // Forward request to your backend gateway
    const backendUrl = BACKEND_URL;
    const response = await fetch(`${backendUrl}/users/register`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
    });

    // Get response data
    const contentType = response.headers.get("content-type");
    let data;

    if (contentType && contentType.includes("application/json")) {
      data = await response.json();
    } else {
      const text = await response.text();
      return NextResponse.json(
        { error: "Invalid response from server" },
        { status: 500 }
      );
    }

    if (!response.ok) {
      return NextResponse.json(
        { error: data.message || data.error || "Registration failed" },
        { status: response.status }
      );
    }

    return NextResponse.json(data);
  } catch (error: any) {
    return NextResponse.json(
      { error: error.message || "Internal server error" },
      { status: 500 }
    );
  }
}
