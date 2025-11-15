'use client'
import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useWebSocket } from '../contexts/WebSocketContext';

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
  const { lastMessage, isConnected } = useWebSocket();

  useEffect(() => {
    const token = localStorage.getItem('token') || sessionStorage.getItem('token');
    if (!token) {
      router.replace('/login');
      return;
    }

    fetchUserAndChats(token);
  }, [router]);

  // Auto-scroll to bottom when messages change
  useEffect(() => {
    if (selectedConversation && selectedConversation.messages.length > 0) {
      setTimeout(() => {
        const messagesEnd = document.getElementById('messages-end');
        if (messagesEnd) {
          messagesEnd.scrollIntoView({ behavior: 'smooth' });
        }
      }, 100);
    }
  }, [selectedConversation?.messages.length]);

  // Handle WebSocket messages from shared connection
  useEffect(() => {
    if (!lastMessage) return;
    
    console.log('[Chat Page] WebSocket message received:', lastMessage);
    
    if (lastMessage.success && lastMessage.data) {
      // Handle new message - data.data contains the message object from MongoDB
      const newMessage = lastMessage.data;
      console.log('[Chat Page] New message received:', newMessage);
      
      // Extract chat_id from the message (it might be an ObjectID string or object)
      const chatId = newMessage.chat_id?.toString() || newMessage.chat_id || '';
      
      // Check if this message belongs to the currently selected conversation
      setSelectedConversation(prev => {
        if (!prev || chatId !== prev.chat_id) return prev;
        // Check if message already exists to avoid duplicates
        const messageId = newMessage._id?.toString() || newMessage.message_id || '';
        const exists = prev.messages.some(m => 
          m.message_id === messageId || 
          (m.message_id === newMessage._id?.toString())
        );
        if (exists) return prev;
        
        const updatedMessages = [...prev.messages, {
          message_id: messageId,
          sender_id: newMessage.sender_id || newMessage.senderID || '',
          message: newMessage.message || newMessage.content || '',
          created_at: newMessage.created_at || new Date().toISOString(),
        }];
        
        // Scroll to bottom after adding new message
        setTimeout(() => {
          const messagesEnd = document.getElementById('messages-end');
          if (messagesEnd) {
            messagesEnd.scrollIntoView({ behavior: 'smooth' });
          }
        }, 100);
        
        return {
          ...prev,
          messages: updatedMessages,
        };
      });
      
      // Update conversations list to show new message and move to top
      setConversations(prev => {
        const updated = prev.map(conv => {
          if (conv.chat_id === chatId) {
            // Update the last message text
            return {
              ...conv,
              lastMessage: newMessage.message || 'New message',
            };
          }
          return conv;
        });
        
        // Move the updated conversation to the top (most recent first)
        const updatedIndex = updated.findIndex(conv => conv.chat_id === chatId);
        if (updatedIndex > 0) {
          const [updatedConv] = updated.splice(updatedIndex, 1);
          updated.unshift(updatedConv);
        }
        
        return updated;
      });
    }
  }, [lastMessage]);

  const fetchUserAndChats = async (token: string) => {
    try {
      setLoading(true);

      // Fetch current user
      const userRes = await fetch('/api/users/me', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
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

      // Fetch chats
      const chatsRes = await fetch('/api/chats', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!chatsRes.ok) {
        throw new Error('Failed to fetch chats');
      }

          const chatsData = await chatsRes.json();
          if (chatsData.success && chatsData.data) {
            const chats: Chat[] = Array.isArray(chatsData.data) ? chatsData.data : [chatsData.data];
            
            // Filter out group chats (only show is_group = false)
            const directChats = chats.filter(chat => !chat.is_group);

            // Convert chats to conversations - use the name field directly
            const convs: Conversation[] = directChats.map(chat => {
              const otherParticipantId = chat.other_participants_id?.[0] || '';
              const participantName = chat.name || 'Unknown';
              
              const initials = participantName
                .split(' ')
                .map((n: string) => n[0])
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

        // Sort by last_message_at or updated_at (most recent first)
        convs.sort((a, b) => {
          const chatA = directChats.find(c => c.chat_id === a.chat_id);
          const chatB = directChats.find(c => c.chat_id === b.chat_id);
          
          const timeA = chatA?.last_message_at || chatA?.updated_at || chatA?.created_at || '';
          const timeB = chatB?.last_message_at || chatB?.updated_at || chatB?.created_at || '';
          
          // Compare timestamps (most recent first)
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
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!res.ok) {
        throw new Error('Failed to fetch messages');
      }

      const data = await res.json();
      if (data.success && data.data) {
        const messages: Message[] = Array.isArray(data.data) ? data.data : [data.data];
        
        // Update the selected conversation with messages
        setSelectedConversation(prev => {
          if (!prev || prev.chat_id !== chatId) return prev;
          return {
            ...prev,
            messages: messages,
          };
        });

        // Update conversations list
        setConversations(prev => prev.map(conv => {
          if (conv.chat_id === chatId) {
            return {
              ...conv,
              messages: messages,
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
    if (conv.messages.length === 0) {
      fetchMessages(conv.chat_id);
    }
  };

  const handleSend = async () => {
    if (!message.trim() || !selectedConversation || sending) return;

    const token = localStorage.getItem('token') || sessionStorage.getItem('token');
    if (!token) {
      router.replace('/login');
      return;
    }

    setSending(true);
    try {
      const res = await fetch('/api/chats/send', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          chat_id: selectedConversation.chat_id,
          message: message.trim(),
        }),
      });

      if (!res.ok) {
        const errorData = await res.json();
        throw new Error(errorData.error || 'Failed to send message');
      }

      // Clear input and refresh messages
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

  if (loading) {
    return (
      <div className="flex h-screen bg-gradient-to-br from-[#5a7568] to-[#4a6558] items-center justify-center">
        <div className="text-center">
          <div className="w-16 h-16 border-4 border-white/30 border-t-white rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-white font-medium">Loading chats...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-screen bg-[#5a7568] overflow-hidden">
      {/* Conversation List */}
      <div className="w-72 bg-white border-r border-gray-200 shadow-sm flex flex-col h-full">
        <div className="p-5 border-b border-gray-200 bg-gradient-to-r from-[#5a7568] to-[#4a6558] flex-shrink-0">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-white/20 backdrop-blur-sm flex items-center justify-center text-sm font-semibold text-white shadow-md">
              {currentUser ? getInitials(currentUser.username) : 'U'}
            </div>
            <div className="flex-1 min-w-0">
              <span className="text-sm font-semibold text-white block truncate">{currentUser?.username || 'User'}</span>
              <span className="text-xs text-white/80">Direct Messages</span>
            </div>
          </div>
        </div>
        
        {/* Conversation Items */}
        <div className="flex-1 overflow-y-auto">
          {conversations.length === 0 ? (
            <div className="flex flex-col items-center justify-center p-8 text-center">
              <div className="w-16 h-16 rounded-full bg-gray-100 flex items-center justify-center mb-4">
                <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" className="text-gray-400">
                  <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
                </svg>
              </div>
              <h3 className="text-sm font-semibold text-gray-900 mb-1">No conversations yet</h3>
              <p className="text-xs text-gray-500">Start a chat from an event to get started!</p>
            </div>
          ) : (
            conversations.map((conv) => (
              <div
                key={conv.chat_id}
                className={`w-full border-b border-gray-100 transition-all duration-150 ${
                  selectedConversation?.chat_id === conv.chat_id 
                    ? 'bg-[#5a7568]/10 border-l-4 border-l-[#5a7568]' 
                    : 'hover:bg-gray-50/50'
                }`}
              >
                <button
                  onClick={() => handleSelectConversation(conv)}
                  className="w-full p-4 transition-colors text-left"
                >
                  <div className="flex items-center gap-3 mb-2">
                    <div className="w-10 h-10 rounded-full bg-gradient-to-br from-[#5a7568] to-[#4a6558] flex items-center justify-center text-xs font-semibold flex-shrink-0 text-white shadow-sm">
                      {conv.initials}
                    </div>
                    <div className="flex-1 min-w-0">
                      <span className="text-sm font-semibold text-gray-900 truncate block">{conv.name}</span>
                    </div>
                  </div>
                  <p className="text-xs text-gray-500 truncate ml-[52px]">{conv.lastMessage || 'No messages yet'}</p>
                </button>
              </div>
            ))
          )}
        </div>
      </div>

      {/* Main Chat Area */}
      {selectedConversation ? (
        <div className="flex-1 flex flex-col bg-white overflow-hidden">
          {/* Chat Header */}
          <div className="h-16 border-b border-gray-200 bg-white shadow-sm flex items-center px-6 flex-shrink-0">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-gradient-to-br from-[#5a7568] to-[#4a6558] flex items-center justify-center text-sm font-semibold text-white shadow-md">
                {selectedConversation.initials}
              </div>
              <div>
                <span className="text-base font-semibold text-gray-900 block">{selectedConversation.name}</span>
              </div>
            </div>
          </div>

          {/* Messages Area */}
          <div className="flex-1 overflow-y-auto p-6 bg-gradient-to-b from-gray-50 to-white relative">
            {selectedConversation.messages.length === 0 ? (
              <div className="flex flex-col items-center justify-center h-full text-center">
                <div className="w-20 h-20 rounded-full bg-gray-100 flex items-center justify-center mb-4">
                  <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" className="text-gray-400">
                    <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
                  </svg>
                </div>
                <h3 className="text-lg font-semibold text-gray-900 mb-1">No messages yet</h3>
                <p className="text-sm text-gray-500">Start the conversation by sending a message!</p>
              </div>
            ) : (
              <div className="space-y-3">
                {selectedConversation.messages.map((msg) => {
                  const isUser = currentUser && msg.sender_id === currentUser.user_id;
                  const messageDate = new Date(msg.created_at);
                  const timeString = messageDate.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
                  
                  return (
                    <div
                      key={msg.message_id}
                      className={`flex flex-col ${isUser ? 'items-end' : 'items-start'} animate-in fade-in slide-in-from-bottom-2 duration-200`}
                    >
                      <div className="flex items-end gap-2 max-w-md">
                        <div
                          className={`px-4 py-2.5 rounded-2xl shadow-sm ${
                            isUser
                              ? 'bg-gradient-to-br from-[#5a7568] to-[#4a6558] text-white rounded-br-md'
                              : 'bg-white text-gray-900 border border-gray-200 rounded-bl-md'
                          }`}
                        >
                          <p className="text-sm leading-relaxed whitespace-pre-wrap break-words">{msg.message}</p>
                          <span className={`text-[10px] mt-1.5 block ${isUser ? 'text-white/70' : 'text-gray-400'}`}>
                            {timeString}
                          </span>
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
            <div id="messages-end" />
          </div>

          {/* Input Area */}
          <div className="border-t border-gray-200 bg-white p-4 shadow-lg flex-shrink-0">
            <div className="flex items-end gap-3 max-w-5xl mx-auto">
              <div className="flex-1 relative">
                <textarea
                  value={message}
                  onChange={(e) => {
                    setMessage(e.target.value);
                    // Auto-resize textarea
                    e.target.style.height = 'auto';
                    e.target.style.height = `${Math.min(e.target.scrollHeight, 120)}px`;
                  }}
                  onKeyPress={handleKeyPress}
                  placeholder={`Message ${selectedConversation.name}...`}
                  className="w-full px-4 py-3 border border-gray-300 rounded-xl resize-none focus:outline-none focus:ring-2 focus:ring-[#5a7568] focus:border-transparent transition-all duration-200 bg-gray-50 focus:bg-white"
                  rows={1}
                  disabled={sending}
                  style={{ maxHeight: '120px', minHeight: '44px' }}
                />
              </div>
              <button
                onClick={handleSend}
                disabled={sending || !message.trim()}
                className="p-3 bg-[#5a7568] text-white rounded-xl hover:bg-[#4a6558] transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed shadow-md hover:shadow-lg transform hover:scale-105 active:scale-95 disabled:transform-none"
                title="Send message"
              >
                {sending ? (
                  <svg className="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                ) : (
                  <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2.5"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  >
                    <line x1="22" y1="2" x2="11" y2="13"></line>
                    <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
                  </svg>
                )}
              </button>
            </div>
          </div>
        </div>
      ) : (
        // Empty State
        <div className="flex-1 flex items-center justify-center bg-white">
          <div className="text-center">
            <div className="w-16 h-16 rounded-full bg-gray-100 flex items-center justify-center mx-auto mb-4">
              <svg
                width="32"
                height="32"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                className="text-gray-400"
              >
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
