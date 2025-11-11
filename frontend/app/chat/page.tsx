'use client'
import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';

interface Message {
  message_id: string;
  sender_id: string;
  message: string;
  created_at: string;
}

interface Chat {
  chat_id: string;
  is_group: boolean;
  name: string; // For direct chats, this is the other participant's username
  other_participants_id: string[];
  last_message_at: string;
  created_at: string;
  updated_at: string;
}

interface Conversation {
  chat_id: string;
  name: string;
  initials: string;
  lastMessage: string;
  otherParticipantId: string;
  messages: Message[];
}

export default function ChatPage() {
  const router = useRouter();
  const [message, setMessage] = useState('');
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [selectedConversation, setSelectedConversation] = useState<Conversation | null>(null);
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const [currentUser, setCurrentUser] = useState<{ username: string; user_id: string } | null>(null);
  const [ws, setWs] = useState<WebSocket | null>(null);

  useEffect(() => {
    const token = localStorage.getItem('token') || sessionStorage.getItem('token');
    if (!token) {
      router.replace('/login');
      return;
    }
    fetchUserAndChats(token);
  }, [router]);

  // WebSocket connection
  useEffect(() => {
    const token = localStorage.getItem('token') || sessionStorage.getItem('token');
    if (!token) return;

    const wsUrl = `ws://localhost:8080/chats/ws/?token=${encodeURIComponent(token)}`;
    const websocket = new WebSocket(wsUrl);

    websocket.onopen = () => console.log('[Chat Page] WebSocket connected');

    websocket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.success && data.data) {
          const newMessage = data.data;
          const chatId = newMessage.chat_id?.toString() || newMessage.chat_id || '';

          // Update selected conversation if it matches
          setSelectedConversation(prev => {
            if (!prev || chatId !== prev.chat_id) return prev;

            const messageId = newMessage._id?.toString() || newMessage.message_id || '';
            const exists = prev.messages.some(m =>
              m.message_id === messageId
            );
            if (exists) return prev;

            return {
              ...prev,
              messages: [
                ...prev.messages,
                {
                  message_id: messageId,
                  sender_id: newMessage.sender_id || newMessage.senderID || '',
                  message: newMessage.message || newMessage.content || '',
                  created_at: newMessage.created_at || new Date().toISOString(),
                },
              ],
            };
          });

          // Update conversations list
          setConversations(prev => {
            const updated = prev.map(conv => {
              if (conv.chat_id === chatId) {
                return {
                  ...conv,
                  lastMessage: newMessage.message || 'New message',
                };
              }
              return conv;
            });

            const updatedIndex = updated.findIndex(conv => conv.chat_id === chatId);
            if (updatedIndex > 0) {
              const [updatedConv] = updated.splice(updatedIndex, 1);
              updated.unshift(updatedConv);
            }

            return updated;
          });
        }
      } catch (err) {
        console.error('[Chat Page] Error parsing WebSocket message:', err);
      }
    };

    websocket.onerror = (error) => console.error('[Chat Page] WebSocket error:', error);
    websocket.onclose = () => console.log('[Chat Page] WebSocket disconnected');

    setWs(websocket);

    return () => {
      if (websocket.readyState === WebSocket.OPEN) {
        websocket.close();
      }
    };
  }, [router]);

  const fetchUserAndChats = async (token: string) => {
    try {
      setLoading(true);

      const userRes = await fetch('/api/users/me', {
        headers: { 'Authorization': `Bearer ${token}` },
      });

      if (userRes.ok) {
        const userData = await userRes.json();
        if (userData.success && userData.data) {
          setCurrentUser({
            username: userData.data.username || 'User',
            user_id: userData.data.user_id?.toString() || '',
          });
        }
      }

      const chatsRes = await fetch('/api/chats', {
        headers: { 'Authorization': `Bearer ${token}` },
      });

      if (!chatsRes.ok) throw new Error('Failed to fetch chats');

      const chatsData = await chatsRes.json();
      if (chatsData.success && chatsData.data) {
        const chats: Chat[] = Array.isArray(chatsData.data) ? chatsData.data : [chatsData.data];
        const directChats = chats.filter(chat => !chat.is_group);

        const convs: Conversation[] = directChats.map(chat => {
          const otherParticipantId = chat.other_participants_id?.[0] || '';
          const participantName = chat.name || 'Unknown';
          const initials = participantName
            .split(' ')
            .map(n => n[0])
            .join('')
            .toUpperCase()
            .slice(0, 2) || 'U';

          return {
            chat_id: chat.chat_id,
            name: participantName,
            initials,
            lastMessage: chat.last_message_at ? 'Last message' : 'No messages yet',
            otherParticipantId,
            messages: [],
          };
        });

        convs.sort((a, b) => {
          const chatA = directChats.find(c => c.chat_id === a.chat_id);
          const chatB = directChats.find(c => c.chat_id === b.chat_id);
          const timeA = chatA?.last_message_at || chatA?.updated_at || chatA?.created_at || '';
          const timeB = chatB?.last_message_at || chatB?.updated_at || chatB?.created_at || '';
          return timeB.localeCompare(timeA);
        });

        setConversations(convs);
      }
    } catch (err: any) {
      console.error('Error fetching chats:', err);
    } finally {
      setLoading(false);
    }
  };

  const fetchMessages = async (chatId: string) => {
    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      if (!token) return;

      const res = await fetch(`/api/chats/${chatId}/messages`, {
        headers: { 'Authorization': `Bearer ${token}` },
      });

      if (!res.ok) throw new Error('Failed to fetch messages');

      const data = await res.json();
      if (data.success && data.data) {
        const messages: Message[] = Array.isArray(data.data) ? data.data : [data.data];

        setSelectedConversation(prev => {
          if (!prev || prev.chat_id !== chatId) return prev;
          return { ...prev, messages };
        });

        setConversations(prev => prev.map(conv => {
          if (conv.chat_id === chatId) {
            return {
              ...conv,
              messages,
              lastMessage: messages.length > 0 ? messages[messages.length - 1].message : 'No messages yet',
            };
          }
          return conv;
        }));
      }
    } catch (err: any) {
      console.error('Error fetching messages:', err);
    }
  };

  const handleSelectConversation = (conv: Conversation) => {
    setSelectedConversation(conv);
    if (conv.messages.length === 0) fetchMessages(conv.chat_id);
  };

  const handleSend = async () => {
    if (!message.trim() || !selectedConversation || sending) return;

    const token = localStorage.getItem('token') || sessionStorage.getItem('token');
    if (!token) return router.replace('/login');

    setSending(true);
    try {
      const res = await fetch('/api/chats/send', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ chat_id: selectedConversation.chat_id, message: message.trim() }),
      });

      if (!res.ok) {
        const errorData = await res.json();
        throw new Error(errorData.error || 'Failed to send message');
      }

      setMessage('');
      await fetchMessages(selectedConversation.chat_id);
    } catch (err: any) {
      console.error('Error sending message:', err);
      alert(err.message || 'Failed to send message');
    } finally {
      setSending(false);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const getInitials = (name: string) => {
    return name
      .split(' ')
      .map(n => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2) || 'U';
  };

  // Scroll to bottom whenever messages change
  useEffect(() => {
    const messagesEnd = document.getElementById('messages-end');
    if (messagesEnd) {
      messagesEnd.scrollIntoView({ behavior: 'smooth' });
    }
  }, [selectedConversation?.messages]);

  if (loading) {
    return (
      <div className="flex h-screen bg-[#5a7568] items-center justify-center">
        <div className="text-white">Loading chats...</div>
      </div>
    );
  }

  return (
    <div className="flex h-screen bg-[#5a7568]">
      {/* Conversation List */}
      <div className="w-52 bg-white border-r border-gray-200">
        <div className="p-4 border-b border-gray-200">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-full bg-gray-300 flex items-center justify-center text-sm font-medium">
              {currentUser ? getInitials(currentUser.username) : 'U'}
            </div>
            <span className="text-sm font-medium">{currentUser?.username || 'User'}</span>
          </div>
        </div>

        <div className="overflow-y-auto" style={{ maxHeight: 'calc(100vh - 80px)' }}>
          {conversations.length === 0 ? (
            <div className="p-4 text-center text-sm text-gray-500">
              No conversations yet. Start a chat from an event!
            </div>
          ) : (
            conversations.map(conv => (
              <button
                key={conv.chat_id}
                onClick={() => handleSelectConversation(conv)}
                className={`w-full p-3 border-b border-gray-100 hover:bg-gray-50 transition-colors text-left ${
                  selectedConversation?.chat_id === conv.chat_id ? 'bg-gray-100' : ''
                }`}
              >
                <div className="flex items-center gap-2 mb-1">
                  <div className="w-8 h-8 rounded-full bg-gray-300 flex items-center justify-center text-xs font-medium flex-shrink-0">
                    {conv.initials}
                  </div>
                  <span className="text-sm font-medium truncate">{conv.name}</span>
                </div>
                <p className="text-xs text-gray-500 truncate ml-10">{conv.lastMessage}</p>
              </button>
            ))
          )}
        </div>
      </div>

      {/* Main Chat Area */}
      {selectedConversation ? (
        <div className="flex-1 flex flex-col bg-white">
          <div className="h-16 border-b border-gray-200 flex items-center px-4">
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-full bg-gray-300 flex items-center justify-center text-sm font-medium">
                {selectedConversation.initials}
              </div>
              <span className="text-sm font-medium">{selectedConversation.name}</span>
            </div>
          </div>

          <div className="flex-1 overflow-y-auto p-4 bg-gray-50">
            {selectedConversation.messages.length === 0 ? (
              <div className="text-center text-gray-500 mt-8">
                No messages yet. Start the conversation!
              </div>
            ) : (
              selectedConversation.messages.map(msg => {
                const isUser = currentUser && msg.sender_id === currentUser.user_id;
                return (
                  <div key={msg.message_id} className={`mb-4 flex ${isUser ? 'justify-end' : 'justify-start'}`}>
                    <div className={`max-w-md px-4 py-2 rounded-lg ${
                      isUser ? 'bg-[#5a7568] text-white' : 'bg-white border border-gray-200'
                    }`}>
                      {msg.message}
                    </div>
                  </div>
                );
              })
            )}
            <div id="messages-end" />
          </div>

          <div className="border-t border-gray-200 p-4">
            <div className="flex items-end gap-2">
              <div className="flex-1 relative">
                <textarea
                  value={message}
                  onChange={e => setMessage(e.target.value)}
                  onKeyPress={handleKeyPress}
                  placeholder={`Message ${selectedConversation.name}...`}
                  className="w-full px-4 py-3 border border-gray-300 rounded-lg resize-none focus:outline-none focus:ring-2 focus:ring-[#5a7568] focus:border-transparent"
                  rows={1}
                  disabled={sending}
                />
              </div>
              <button
                onClick={handleSend}
                disabled={sending || !message.trim()}
                className="p-3 text-[#5a7568] hover:bg-gray-100 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <line x1="22" y1="2" x2="11" y2="13"></line>
                  <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
                </svg>
              </button>
            </div>
          </div>
        </div>
      ) : (
        <div className="flex-1 flex items-center justify-center bg-white">
          <div className="text-center">
            <div className="w-16 h-16 rounded-full bg-gray-100 flex items-center justify-center mx-auto mb-4">
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-gray-400">
                <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
              </svg>
            </div>
            <h3 className="text-lg font-medium text-gray-900 mb-2">No conversation selected</h3>
            <p className="text-sm text-gray-500">Choose a conversation from the list to start chatting</p>
          </div>
        </div>
      )}
    </div>
  );
}
