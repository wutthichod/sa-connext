// app/api/events/route.ts
import { NextRequest, NextResponse } from "next/server";
import { BACKEND_URL } from "@/app/config";

// GET - Get all events or user events
export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const userOnly = searchParams.get("user") === "true";

    const token =
      request.headers.get("authorization")?.replace("Bearer ", "") ||
      request.cookies.get("token")?.value;

    if (!token) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    const endpoint = userOnly
      ? `${BACKEND_URL}/events/user`
      : `${BACKEND_URL}/events`;

    const response = await fetch(endpoint, {
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
      return NextResponse.json(
        { error: "Invalid response from server" },
        { status: 500 }
      );
    }

    if (!response.ok) {
      return NextResponse.json(
        { error: data.message || data.error || "Failed to fetch events" },
        { status: response.status }
      );
    }

    return NextResponse.json(data);
  } catch (error: any) {
    console.error("Events API error:", error);
    return NextResponse.json(
      { error: error.message || "Internal server error" },
      { status: 500 }
    );
  }
}

// POST - Create event
export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    console.log(
      "[API] Create event request body:",
      JSON.stringify(body, null, 2)
    );

    const token =
      request.headers.get("authorization")?.replace("Bearer ", "") ||
      request.cookies.get("token")?.value;

    if (!token) {
      console.error("[API] No token provided");
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    console.log("[API] Calling backend:", `${BACKEND_URL}/events`);
    let response: Response;
    try {
      response = await fetch(`${BACKEND_URL}/events`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(body),
      });
    } catch (fetchError: any) {
      console.error("[API] Fetch error:", fetchError);
      return NextResponse.json(
        {
          error: `Failed to connect to backend: ${
            fetchError.message || "Connection refused"
          }. Make sure the backend is running on ${BACKEND_URL}`,
        },
        { status: 503 }
      );
    }

    console.log("[API] Backend response status:", response.status);

    const contentType = response.headers.get("content-type") || "";
    let data: any;

    if (contentType.includes("application/json")) {
      data = await response.json();
      console.log(
        "[API] Backend response data:",
        JSON.stringify(data, null, 2)
      );
    } else {
      const text = await response.text();
      console.error("[API] Backend returned non-JSON:", text);
      return NextResponse.json(
        { error: "Invalid response from server" },
        { status: 500 }
      );
    }

    if (!response.ok) {
      console.error("[API] Backend error response:", {
        status: response.status,
        data: data,
      });
      return NextResponse.json(
        { error: data.message || data.error || "Failed to create event" },
        { status: response.status }
      );
    }

    console.log("[API] Create event successful");
    return NextResponse.json(data);
  } catch (error: any) {
    console.error("[API] Create event exception:", error);
    return NextResponse.json(
      { error: error.message || "Internal server error" },
      { status: 500 }
    );
  }
}
