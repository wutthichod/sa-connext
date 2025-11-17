// app/api/chats/[chatId]/join/route.ts
import { NextRequest, NextResponse } from "next/server";
import { BACKEND_URL } from "@/app/config";

// POST - Join a group chat
export async function POST(
  request: NextRequest,
  { params }: { params: Promise<{ chatId: string }> | { chatId: string } }
) {
  try {
    // Handle both Promise and direct params (for different Next.js versions)
    const resolvedParams = params instanceof Promise ? await params : params;
    const chatId = resolvedParams.chatId;

    if (!chatId) {
      return NextResponse.json(
        { error: "chatId is required" },
        { status: 400 }
      );
    }

    const token =
      request.headers.get("authorization")?.replace("Bearer ", "") ||
      request.cookies.get("token")?.value;

    if (!token) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    console.log(`[API] Joining group chat: ${chatId}`);

    const response = await fetch(`${BACKEND_URL}/chats/${chatId}/join`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
    });

    console.log(`[API] Backend response status: ${response.status}`);

    const contentType = response.headers.get("content-type") || "";
    let data: any;

    if (contentType.includes("application/json")) {
      data = await response.json();
      console.log(
        `[API] Backend response data:`,
        JSON.stringify(data, null, 2)
      );
    } else {
      const text = await response.text();
      console.error(`[API] Backend returned non-JSON:`, text);
      return NextResponse.json(
        { error: "Invalid response from server" },
        { status: 500 }
      );
    }

    if (!response.ok) {
      console.error(`[API] Backend error:`, data);
      return NextResponse.json(
        { error: data.message || data.error || "Failed to join group chat" },
        { status: response.status }
      );
    }

    console.log(`[API] Successfully joined group chat: ${chatId}`);
    return NextResponse.json(data);
  } catch (error: any) {
    console.error("Join group chat API error:", error);
    return NextResponse.json(
      { error: error.message || "Internal server error" },
      { status: 500 }
    );
  }
}
