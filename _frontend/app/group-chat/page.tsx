'use client'
import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useWebSocket } from '../contexts/WebSocketContext';

interface Message {
  message_id: string;
  sender_id: string;
  message: string;
  created_at: string;
  sender_username?: string; // Username of the sender
}

interface Chat {
  chat_id: string;
  is_group: boolean;
  name: string; // For group chats, this is the group name
  other_participants_id: string[];
  last_message_at: string;
  created_at: string;
  updated_at: string;
}

interface GroupConversation {
  chat_id: string;
  name: string;
  initials: string;
  lastMessage: string;
  participantCount: number;
  messages: Message[];
  isMember: boolean; // Whether the current user is a member of this group
}

export default function GroupChatPage() {
  const router = useRouter();
  const [message, setMessage] = useState('');
  const [conversations, setConversations] = useState<GroupConversation[]>([]);
  const [selectedConversation, setSelectedConversation] = useState<GroupConversation | null>(null);
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const [currentUser, setCurrentUser] = useState<{ username: string; user_id: string } | null>(null);
  const { lastMessage, isConnected } = useWebSocket();
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showJoinModal, setShowJoinModal] = useState(false);
  const [groupName, setGroupName] = useState('');
  const [joinChatId, setJoinChatId] = useState('');
  const [creating, setCreating] = useState(false);
  const [joining, setJoining] = useState(false);
  const [joiningGroupId, setJoiningGroupId] = useState<string | null>(null);
  const [groupChatsData, setGroupChatsData] = useState<Chat[]>([]);
  const [knownMemberships, setKnownMemberships] = useState<Set<string>>(new Set());
  const [usernames, setUsernames] = useState<Map<string, string>>(new Map()); // Map of user_id -> username
  const [groupMembers, setGroupMembers] = useState<Array<{ user_id: string; username: string }>>([]);
  const [showMembersList, setShowMembersList] = useState(false);
  const [myGroupsCollapsed, setMyGroupsCollapsed] = useState(false);
  const [otherGroupsCollapsed, setOtherGroupsCollapsed] = useState(false);

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
    
    console.log('[Group Chat Page] WebSocket message received:', lastMessage);
    
    if (lastMessage.success && lastMessage.data) {
      const newMessage = lastMessage.data;
      console.log('[Group Chat Page] New message received:', newMessage);
      
      const chatId = newMessage.chat_id?.toString() || newMessage.chat_id || '';
      
      // Check if this message belongs to the currently selected conversation
      setSelectedConversation(prev => {
        if (!prev || chatId !== prev.chat_id) return prev;
        const messageId = newMessage._id?.toString() || newMessage.message_id || '';
        const exists = prev.messages.some(m => 
          m.message_id === messageId || 
          (m.message_id === newMessage._id?.toString())
        );
        if (exists) return prev;
        
        // Fetch username for new message sender if not already cached
        const senderId = newMessage.sender_id || newMessage.senderID || '';
        let senderUsername = usernames.get(senderId);
        if (!senderUsername && senderId) {
          // Fetch username asynchronously
          fetchUsername(senderId).then(username => {
            setSelectedConversation(prev => {
              if (!prev) return prev;
              return {
                ...prev,
                messages: prev.messages.map(m => 
                  m.message_id === messageId 
                    ? { ...m, sender_username: username }
                    : m
                ),
              };
            });
          });
        }
        
        const updatedMessages = [...prev.messages, {
          message_id: messageId,
          sender_id: senderId,
          message: newMessage.message || newMessage.content || '',
          created_at: newMessage.created_at || new Date().toISOString(),
          sender_username: senderUsername || senderId,
        }];
        
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
  }, [lastMessage, usernames]);

  const fetchUserAndChats = async (token: string) => {
    try {
      setLoading(true);

      // Fetch current user
      let fetchedUserId = '';
      const userRes = await fetch('/api/users/me', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (userRes.ok) {
        const userData = await userRes.json();
        if (userData.success && userData.data) {
          fetchedUserId = userData.data.user_id?.toString() || '';
          setCurrentUser({
            username: userData.data.username || 'User',
            user_id: fetchedUserId,
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
        
        // Filter for group chats only (is_group = true)
        const groupChats = chats.filter(chat => chat.is_group);
        setGroupChatsData(groupChats); // Store for later reference

        // Convert chats to conversations
        const convs: GroupConversation[] = groupChats.map(chat => {
          const groupName = chat.name || 'Unnamed Group';
          
          const initials = groupName
            .split(' ')
            .map((n: string) => n[0])
            .join('')
            .toUpperCase()
            .slice(0, 2) || 'GC';

          // For groups (is_group: true), backend's other_participants_id contains ALL participants
          // So the participant count is simply the length of other_participants_id
          const otherParticipantIds = chat.other_participants_id || [];
          const participantCount = otherParticipantIds.length;
          
          // Check if current user is in the participants list
          // For groups, other_participants_id contains ALL participants, so we can check directly
          // Use fetchedUserId instead of currentUser state (which might not be updated yet)
          // Also check both string and number formats in case of type mismatch
          let isUserInParticipants = false;
          if (fetchedUserId && otherParticipantIds.length > 0) {
            // Normalize both to strings for comparison
            const normalizedUserId = fetchedUserId.toString().trim();
            // Check exact match (both as strings)
            isUserInParticipants = otherParticipantIds.some(pid => {
              const normalizedPid = pid.toString().trim();
              return normalizedPid === normalizedUserId;
            });
          }
          
          // Check if we already know this user is a member (from knownMemberships set)
          // OR if the user is in the participants list from the backend
          // This preserves membership status after joins/creates and checks backend data
          // IMPORTANT: Only trust knownMemberships if user is actually in participants list
          // This prevents stale knownMemberships from incorrectly marking groups as members
          const isMember = (knownMemberships.has(chat.chat_id) && isUserInParticipants) || isUserInParticipants;
          
          // Debug logging
          if (fetchedUserId) {
            console.log(`[Membership Check] Group: ${chat.chat_id} (${groupName}), User: ${fetchedUserId}, In Participants: ${isUserInParticipants}, Known Member: ${knownMemberships.has(chat.chat_id)}, Is Member: ${isMember}, Participant Count: ${otherParticipantIds.length}, Participants:`, otherParticipantIds);
          }

          return {
            chat_id: chat.chat_id,
            name: groupName,
            initials,
            lastMessage: chat.last_message_at ? 'Last message' : 'No messages yet',
            participantCount: participantCount,
            messages: [],
            isMember: isMember, // Explicitly false for groups not in knownMemberships
          };
        });

        // Sort by last_message_at or updated_at (most recent first)
        convs.sort((a, b) => {
          const chatA = groupChats.find(c => c.chat_id === a.chat_id);
          const chatB = groupChats.find(c => c.chat_id === b.chat_id);
          
          const timeA = chatA?.last_message_at || chatA?.updated_at || chatA?.created_at || '';
          const timeB = chatB?.last_message_at || chatB?.updated_at || chatB?.created_at || '';
          
          return timeB.localeCompare(timeA);
        });

        setConversations(convs);

        // Note: We don't do async membership checks via message fetching because
        // the backend GetMessagesByChatId doesn't check authorization, so it would
        // incorrectly mark all groups as members. Instead, we rely on the
        // other_participants_id check from the initial data fetch, which correctly
        // contains all participants for groups.
      }
    } catch (err: any) {
      console.error('Error fetching group chats:', err);
    } finally {
      setLoading(false);
    }
  };

  const fetchUsername = async (userId: string): Promise<string> => {
    // Check if we already have the username cached
    if (usernames.has(userId)) {
      return usernames.get(userId) || userId;
    }

    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      if (!token) return userId;

      const res = await fetch(`/api/users/${userId}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (res.ok) {
        const data = await res.json();
        if (data.success && data.data && data.data.username) {
          const username = data.data.username;
          setUsernames(prev => new Map(prev).set(userId, username));
          return username;
        }
      }
    } catch (err) {
      console.error(`Error fetching username for user ${userId}:`, err);
    }

    return userId; // Fallback to user ID if we can't fetch username
  };

  const fetchGroupMembers = async (chatId: string) => {
    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      if (!token) return;

      // Get the chat data to find all participants
      const chat = groupChatsData.find(c => c.chat_id === chatId);
      if (!chat) return;

      // For groups (is_group: true), backend's other_participants_id contains ALL participants
      const allParticipantIds = chat.other_participants_id || [];

      // Fetch usernames for all participants
      const memberPromises = allParticipantIds.map(async (userId) => {
        const username = await fetchUsername(userId);
        return { user_id: userId, username };
      });

      const members = await Promise.all(memberPromises);
      setGroupMembers(members);
    } catch (err) {
      console.error('Error fetching group members:', err);
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
        // If we can't fetch messages, user is likely not a member
        if (res.status === 403 || res.status === 404) {
          setConversations(prev => prev.map(conv => {
            if (conv.chat_id === chatId) {
              return {
                ...conv,
                isMember: false,
              };
            }
            return conv;
          }));
        }
        throw new Error('Failed to fetch messages');
      }

      const data = await res.json();
      if (data.success && data.data) {
        const messages: Message[] = Array.isArray(data.data) ? data.data : [data.data];
        
        // Fetch usernames for all unique sender IDs
        const uniqueSenderIds = [...new Set(messages.map(m => m.sender_id))];
        const usernamePromises = uniqueSenderIds.map(async (senderId) => {
          const username = await fetchUsername(senderId);
          return { senderId, username };
        });
        const senderUsernames = await Promise.all(usernamePromises);
        
        // Add usernames to messages
        const messagesWithUsernames = messages.map(msg => {
          const senderInfo = senderUsernames.find(s => s.senderId === msg.sender_id);
          return {
            ...msg,
            sender_username: senderInfo?.username || msg.sender_id,
          };
        });
        
        // If we can fetch messages, user is a member
        setSelectedConversation(prev => {
          if (!prev || prev.chat_id !== chatId) return prev;
          return {
            ...prev,
            messages: messagesWithUsernames,
            isMember: true,
          };
        });

        // Mark as known member
        setKnownMemberships((prev: Set<string>) => new Set(prev).add(chatId));
        
        setConversations(prev => prev.map(conv => {
          if (conv.chat_id === chatId) {
            // Find the original chat data to get accurate participant count
            const originalChat = groupChatsData.find(c => c.chat_id === chatId);
            // For groups, other_participants_id contains ALL participants
            const participantCount = originalChat?.other_participants_id?.length || 0;
            
            return {
              ...conv,
              messages: messagesWithUsernames,
              lastMessage: messagesWithUsernames.length > 0 ? messagesWithUsernames[messagesWithUsernames.length - 1].message : 'No messages yet',
              isMember: true,
              participantCount: participantCount,
            };
          }
          return conv;
        }));

        // Fetch group members when messages are loaded
        await fetchGroupMembers(chatId);
      }
    } catch (err: any) {
      console.error('Error fetching messages:', err);
    }
  };

  const handleSelectConversation = async (conv: GroupConversation) => {
    // If user is not a member, immediately trigger join action
    // Don't try to view messages - they need to join first
    if (!conv.isMember) {
      handleJoinGroupDirectly(conv.chat_id);
      return;
    }
    
    // User is a member, proceed to view the conversation
    setSelectedConversation(conv);
    if (conv.messages.length === 0) {
      fetchMessages(conv.chat_id);
    } else {
      // Fetch group members even if messages are already loaded
      fetchGroupMembers(conv.chat_id);
    }
  };

  const handleJoinGroupFromList = (chatId: string, e: React.MouseEvent) => {
    e.stopPropagation(); // Prevent triggering handleSelectConversation
    handleJoinGroupDirectly(chatId);
  };

  const handleJoinGroupDirectly = async (chatId: string) => {
    if (joining || !chatId) return;

    const token = localStorage.getItem('token') || sessionStorage.getItem('token');
    if (!token) {
      router.replace('/login');
      return;
    }

    setJoining(true);
    try {
      console.log(`[Frontend] Attempting to join group: ${chatId}`);
      const res = await fetch(`/api/chats/${chatId}/join`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      console.log(`[Frontend] Join response status: ${res.status}`);

      const contentType = res.headers.get('content-type') || '';
      let result: any;

      if (contentType.includes('application/json')) {
        result = await res.json();
      } else {
        const text = await res.text();
        result = { error: text || 'Unknown error' };
      }

      if (!res.ok) {
        const errorMessage = result.error || result.message || 'Failed to join group';
        console.error(`[Frontend] Join error:`, result);
        throw new Error(errorMessage);
      }

      console.log(`[Frontend] Successfully joined group:`, result);

      // Mark user as a member in knownMemberships
      setKnownMemberships((prev: Set<string>) => new Set(prev).add(chatId));
      
      // Update membership status immediately
      setConversations(prev => prev.map(conv => {
        if (conv.chat_id === chatId) {
          const originalChat = groupChatsData.find(c => c.chat_id === chatId);
          // For groups, other_participants_id contains ALL participants
          // After joining, refresh will get the updated list, but for now use current count
          const participantCount = originalChat?.other_participants_id?.length || 0;
          
          return {
            ...conv,
            isMember: true,
            participantCount: participantCount,
          };
        }
        return conv;
      }));

      // Refresh the group chat list to get updated participant counts
      const token2 = localStorage.getItem('token') || sessionStorage.getItem('token');
      if (token2) {
        await fetchUserAndChats(token2);
      }
    } catch (err: any) {
      console.error('Error joining group:', err);
      alert(err.message || 'Failed to join group');
    } finally {
      setJoining(false);
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
      .slice(0, 2) || 'GC';
  };

  const handleCreateGroup = async () => {
    if (!groupName.trim() || creating) return;

    const token = localStorage.getItem('token') || sessionStorage.getItem('token');
    if (!token) {
      router.replace('/login');
      return;
    }

    setCreating(true);
    try {
      const res = await fetch('/api/chats/group', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          group_name: groupName.trim(),
        }),
      });

      if (!res.ok) {
        const errorData = await res.json();
        throw new Error(errorData.error || 'Failed to create group');
      }

      const result = await res.json();
      const newChatId = result.data?.chat_id || result.chat_id;
      
      setGroupName('');
      setShowCreateModal(false);
      
      // Refresh the group chat list
      const token2 = localStorage.getItem('token') || sessionStorage.getItem('token');
      if (token2) {
        await fetchUserAndChats(token2);
        
        // After refresh, mark the newly created group as one the user is a member of
        // Creator is automatically a member (backend adds them to participants)
        if (newChatId) {
          // Mark creator as a member
          setKnownMemberships((prev: Set<string>) => new Set(prev).add(newChatId));
          
          // Update the conversation to mark as member
          setConversations(prev => prev.map(conv => {
            if (conv.chat_id === newChatId) {
              const originalChat = groupChatsData.find(c => c.chat_id === newChatId);
              const otherParticipantsCount = originalChat?.other_participants_id?.length || 0;
              return {
                ...conv,
                isMember: true, // Creator is automatically a member
                participantCount: otherParticipantsCount + 1,
              };
            }
            return conv;
          }));
          
          // Fetch messages to load the chat
          await fetchMessages(newChatId);
        }
      }
    } catch (err: any) {
      console.error('Error creating group:', err);
      alert(err.message || 'Failed to create group');
    } finally {
      setCreating(false);
    }
  };

  const handleJoinGroup = async () => {
    const chatIdToJoin = joinChatId.trim() || joiningGroupId;
    if (!chatIdToJoin || joining) return;

    const token = localStorage.getItem('token') || sessionStorage.getItem('token');
    if (!token) {
      router.replace('/login');
      return;
    }

    setJoining(true);
    try {
      console.log(`[Frontend] Attempting to join group: ${chatIdToJoin}`);
      const res = await fetch(`/api/chats/${chatIdToJoin}/join`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      console.log(`[Frontend] Join response status: ${res.status}`);

      const contentType = res.headers.get('content-type') || '';
      let result: any;

      if (contentType.includes('application/json')) {
        result = await res.json();
      } else {
        const text = await res.text();
        result = { error: text || 'Unknown error' };
      }

      if (!res.ok) {
        const errorMessage = result.error || result.message || 'Failed to join group';
        console.error(`[Frontend] Join error:`, result);
        throw new Error(errorMessage);
      }

      console.log(`[Frontend] Successfully joined group:`, result);

      // Update membership status immediately
      setConversations(prev => prev.map(conv => {
        if (conv.chat_id === chatIdToJoin) {
          const originalChat = groupChatsData.find(c => c.chat_id === chatIdToJoin);
          const otherParticipantsCount = originalChat?.other_participants_id?.length || 0;
          return {
            ...conv,
            isMember: true,
            participantCount: otherParticipantsCount + 1, // User is now a member, so +1
          };
        }
        return conv;
      }));

      setJoinChatId('');
      setJoiningGroupId(null);
      setShowJoinModal(false);
      
      // Refresh the group chat list to get updated participant counts
      const token2 = localStorage.getItem('token') || sessionStorage.getItem('token');
      if (token2) {
        await fetchUserAndChats(token2);
      }
    } catch (err: any) {
      console.error('Error joining group:', err);
      alert(err.message || 'Failed to join group');
    } finally {
      setJoining(false);
    }
  };

  if (loading) {
    return (
      <div className="flex h-screen bg-gradient-to-br from-[#5a7568] to-[#4a6558] items-center justify-center">
        <div className="text-center">
          <div className="w-16 h-16 border-4 border-white/30 border-t-white rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-white font-medium">Loading group chats...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-screen bg-[#5a7568] overflow-hidden">
      {/* Group Chat List */}
      <div className="w-72 bg-white border-r border-gray-200 shadow-sm flex flex-col h-full">
        <div className="p-5 border-b border-gray-200 bg-gradient-to-r from-[#5a7568] to-[#4a6558] flex-shrink-0">
          <div className="flex items-center gap-3 mb-4">
            <div className="w-10 h-10 rounded-full bg-white/20 backdrop-blur-sm flex items-center justify-center text-sm font-semibold text-white shadow-md">
              {currentUser ? getInitials(currentUser.username) : 'U'}
            </div>
            <div className="flex-1 min-w-0">
              <span className="text-sm font-semibold text-white block truncate">{currentUser?.username || 'User'}</span>
              <span className="text-xs text-white/80">Group Chats</span>
            </div>
          </div>
          <button
            onClick={() => setShowCreateModal(true)}
            className="w-full px-4 py-2.5 bg-white text-[#5a7568] text-sm font-medium rounded-lg hover:bg-white/90 transition-all duration-200 flex items-center justify-center gap-2 shadow-md hover:shadow-lg transform hover:scale-[1.02]"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
              <line x1="12" y1="5" x2="12" y2="19"></line>
              <line x1="5" y1="12" x2="19" y2="12"></line>
            </svg>
            Create Group
          </button>
        </div>
        
        {/* Group Chat Items */}
        <div className="flex-1 overflow-y-auto">
          {conversations.length === 0 ? (
            <div className="p-4 text-center text-sm text-gray-500">
              No group chats yet. Create or join a group!
            </div>
          ) : (
            <>
              {/* My Groups Section - Groups where user is a member */}
              {conversations.filter(conv => conv.isMember).length > 0 && (
                <div className="border-b border-gray-200">
                  <button
                    onClick={() => setMyGroupsCollapsed(!myGroupsCollapsed)}
                    className="w-full px-5 py-3 bg-gray-50/50 border-b border-gray-100 hover:bg-gray-50 transition-colors flex items-center justify-between"
                  >
                    <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider">My Groups</h3>
                    <svg
                      width="14"
                      height="14"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="2.5"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      className={`text-gray-500 transition-transform duration-200 ${myGroupsCollapsed ? '' : 'rotate-180'}`}
                    >
                      <polyline points="6 9 12 15 18 9"></polyline>
                    </svg>
                  </button>
                  {!myGroupsCollapsed && (
                    <div className="transition-all duration-200">
                      {conversations
                        .filter(conv => conv.isMember)
                        .map((conv) => (
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
                                  <span className="text-xs text-gray-500 mt-0.5 block">{conv.participantCount} {conv.participantCount === 1 ? 'member' : 'members'}</span>
                                </div>
                              </div>
                              <p className="text-xs text-gray-500 truncate ml-[52px]">{conv.lastMessage || 'No messages yet'}</p>
                            </button>
                          </div>
                        ))}
                    </div>
                  )}
                </div>
              )}

              {/* Other Groups Section - Groups where user is not a member */}
              {conversations.filter(conv => !conv.isMember).length > 0 && (
                <div>
                  <button
                    onClick={() => setOtherGroupsCollapsed(!otherGroupsCollapsed)}
                    className="w-full px-5 py-3 bg-gray-50/50 border-t border-gray-200 border-b border-gray-100 hover:bg-gray-50 transition-colors flex items-center justify-between"
                  >
                    <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider">Other Groups</h3>
                    <svg
                      width="14"
                      height="14"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="2.5"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      className={`text-gray-500 transition-transform duration-200 ${otherGroupsCollapsed ? '' : 'rotate-180'}`}
                    >
                      <polyline points="6 9 12 15 18 9"></polyline>
                    </svg>
                  </button>
                  {!otherGroupsCollapsed && (
                    <div className="transition-all duration-200">
                      {conversations
                        .filter(conv => !conv.isMember)
                        .map((conv) => (
                          <div
                            key={conv.chat_id}
                            className="w-full border-b border-gray-100 hover:bg-gray-50/30 transition-colors"
                          >
                            <div className="w-full p-4">
                              <div className="flex items-center gap-3 mb-2">
                                <div className="w-10 h-10 rounded-full bg-gradient-to-br from-gray-400 to-gray-500 flex items-center justify-center text-xs font-semibold flex-shrink-0 text-white shadow-sm">
                                  {conv.initials}
                                </div>
                                <div className="flex-1 min-w-0">
                                  <span className="text-sm font-semibold text-gray-900 truncate block">{conv.name}</span>
                                  <span className="text-xs text-gray-500 mt-0.5 block">{conv.participantCount} {conv.participantCount === 1 ? 'member' : 'members'}</span>
                                </div>
                              </div>
                              <p className="text-xs text-gray-500 truncate ml-[52px] mb-3">{conv.lastMessage || 'No messages yet'}</p>
                              <button
                                onClick={(e) => handleJoinGroupFromList(conv.chat_id, e)}
                                disabled={joining}
                                className="w-full px-4 py-2 bg-[#5a7568] text-white text-sm font-medium rounded-lg hover:bg-[#4a6558] transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed shadow-sm hover:shadow-md transform hover:scale-[1.02] active:scale-[0.98]"
                              >
                                {joining ? (
                                  <span className="flex items-center justify-center gap-2">
                                    <svg className="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                                    </svg>
                                    Joining...
                                  </span>
                                ) : (
                                  'Join Group'
                                )}
                              </button>
                            </div>
                          </div>
                        ))}
                    </div>
                  )}
                </div>
              )}

              {/* Show message if no groups at all */}
              {conversations.filter(conv => conv.isMember).length === 0 && 
               conversations.filter(conv => !conv.isMember).length === 0 && (
                <div className="flex flex-col items-center justify-center p-8 text-center">
                  <div className="w-16 h-16 rounded-full bg-gray-100 flex items-center justify-center mb-4">
                    <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" className="text-gray-400">
                      <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
                      <circle cx="9" cy="7" r="4"></circle>
                      <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
                      <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
                    </svg>
                  </div>
                  <h3 className="text-sm font-semibold text-gray-900 mb-1">No group chats yet</h3>
                  <p className="text-xs text-gray-500 mb-4">Create or join a group to get started!</p>
                  <button
                    onClick={() => setShowCreateModal(true)}
                    className="px-4 py-2 bg-[#5a7568] text-white text-sm font-medium rounded-lg hover:bg-[#4a6558] transition-all duration-200 shadow-sm hover:shadow-md"
                  >
                    Create Your First Group
                  </button>
                </div>
              )}
            </>
          )}
        </div>
      </div>

      {/* Main Chat Area */}
      {selectedConversation ? (
        <div className="flex-1 flex flex-col bg-white overflow-hidden">
          {/* Chat Header */}
          <div className="h-16 border-b border-gray-200 bg-white shadow-sm flex items-center justify-between px-6 flex-shrink-0">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-gradient-to-br from-[#5a7568] to-[#4a6558] flex items-center justify-center text-sm font-semibold text-white shadow-md">
                {selectedConversation.initials}
              </div>
              <div>
                <span className="text-base font-semibold text-gray-900 block">{selectedConversation.name}</span>
                <span className="text-xs text-gray-500">{selectedConversation.participantCount} {selectedConversation.participantCount === 1 ? 'member' : 'members'}</span>
              </div>
            </div>
            <button
              onClick={() => setShowMembersList(!showMembersList)}
              className={`px-4 py-2 text-sm font-medium rounded-lg transition-all duration-200 ${
                showMembersList 
                  ? 'bg-[#5a7568] text-white shadow-md' 
                  : 'text-[#5a7568] hover:bg-gray-100'
              }`}
              title="View members"
            >
              <svg
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
                <circle cx="9" cy="7" r="4"></circle>
                <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
                <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
              </svg>
            </button>
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
                {selectedConversation.messages.map((msg, index) => {
                  const isUser = currentUser && msg.sender_id === currentUser.user_id;
                  const showSenderName = !isUser && (index === 0 || selectedConversation.messages[index - 1].sender_id !== msg.sender_id);
                  const senderName = msg.sender_username || msg.sender_id;
                  const messageDate = new Date(msg.created_at);
                  const timeString = messageDate.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
                  
                  return (
                    <div
                      key={msg.message_id}
                      className={`flex ${isUser ? 'justify-end' : 'justify-start'} mb-4 animate-in fade-in slide-in-from-bottom-2 duration-200`}
                    >
                      {!isUser && (
                        <div className="flex items-end gap-2 max-w-md">
                          {/* Avatar - always shown for non-user messages */}
                          <div className="w-6 h-6 rounded-full bg-gradient-to-br from-gray-400 to-gray-500 flex items-center justify-center text-[10px] font-semibold text-white shadow-sm flex-shrink-0 mb-1">
                            {senderName.charAt(0).toUpperCase()}
                          </div>
                          <div className="flex flex-col">
                            {/* Sender name - only shown when it's a new sender */}
                            {showSenderName && (
                              <span className="text-xs font-medium text-gray-600 mb-1 px-1">{senderName}</span>
                            )}
                            <div
                              className="px-4 py-2.5 rounded-2xl shadow-sm bg-white text-gray-900 border border-gray-200 rounded-bl-md"
                            >
                              <p className="text-sm leading-relaxed whitespace-pre-wrap break-words">{msg.message}</p>
                              <span className="text-[10px] mt-1.5 block text-gray-400">
                                {timeString}
                              </span>
                            </div>
                          </div>
                        </div>
                      )}
                      {isUser && (
                        <div className="flex flex-col items-end max-w-md">
                          <div
                            className="px-4 py-2.5 rounded-2xl shadow-sm bg-gradient-to-br from-[#5a7568] to-[#4a6558] text-white rounded-br-md"
                          >
                            <p className="text-sm leading-relaxed whitespace-pre-wrap break-words">{msg.message}</p>
                            <span className="text-[10px] mt-1.5 block text-white/70">
                              {timeString}
                            </span>
                          </div>
                        </div>
                      )}
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

          {/* Members List Sidebar */}
          {showMembersList && (
            <div className="absolute right-0 top-16 bottom-0 w-72 bg-white border-l border-gray-200 shadow-2xl z-10 animate-in slide-in-from-right duration-200">
              <div className="h-full flex flex-col">
                <div className="p-5 border-b border-gray-200 bg-gradient-to-r from-[#5a7568] to-[#4a6558]">
                  <div className="flex items-center justify-between mb-1">
                    <h3 className="text-sm font-bold text-white">Members</h3>
                    <button
                      onClick={() => setShowMembersList(false)}
                      className="text-white/80 hover:text-white transition-colors p-1 rounded-lg hover:bg-white/10"
                      title="Close"
                    >
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
                        <line x1="18" y1="6" x2="6" y2="18"></line>
                        <line x1="6" y1="6" x2="18" y2="18"></line>
                      </svg>
                    </button>
                  </div>
                  <p className="text-xs text-white/80">{groupMembers.length} {groupMembers.length === 1 ? 'member' : 'members'}</p>
                </div>
                <div className="flex-1 overflow-y-auto p-4 bg-gray-50">
                  {groupMembers.length === 0 ? (
                    <div className="flex flex-col items-center justify-center mt-8">
                      <div className="w-12 h-12 rounded-full bg-gray-200 flex items-center justify-center mb-3">
                        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" className="text-gray-400">
                          <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
                          <circle cx="9" cy="7" r="4"></circle>
                          <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
                          <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
                        </svg>
                      </div>
                      <p className="text-sm text-gray-500">Loading members...</p>
                    </div>
                  ) : (
                    <div className="space-y-2">
                      {groupMembers.map((member) => {
                        const isCurrentUser = currentUser?.user_id === member.user_id;
                        return (
                          <div
                            key={member.user_id}
                            className={`flex items-center gap-3 py-3 px-3 rounded-xl transition-all duration-200 ${
                              isCurrentUser 
                                ? 'bg-[#5a7568]/10 border-2 border-[#5a7568]/20' 
                                : 'bg-white hover:bg-gray-100 border border-gray-200'
                            }`}
                          >
                            <div className={`w-10 h-10 rounded-full flex items-center justify-center text-xs font-semibold text-white shadow-sm ${
                              isCurrentUser 
                                ? 'bg-gradient-to-br from-[#5a7568] to-[#4a6558]' 
                                : 'bg-gradient-to-br from-gray-400 to-gray-500'
                            }`}>
                              {member.username.charAt(0).toUpperCase()}
                            </div>
                            <div className="flex-1 min-w-0">
                              <div className="text-sm font-semibold text-gray-900 truncate">
                                {member.username}
                                {isCurrentUser && (
                                  <span className="text-xs text-[#5a7568] ml-2 font-medium">(You)</span>
                                )}
                              </div>
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  )}
                </div>
              </div>
            </div>
          )}
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
                <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
                <circle cx="9" cy="7" r="4"></circle>
                <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
                <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
              </svg>
            </div>
            <h3 className="text-lg font-medium text-gray-900 mb-2">No group chat selected</h3>
            <p className="text-sm text-gray-500">Choose a group chat from the list to start chatting</p>
          </div>
        </div>
      )}

      {/* Create Group Modal */}
      {showCreateModal && (
        <div 
          className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 animate-in fade-in duration-200"
          onClick={() => {
            if (!creating) {
              setShowCreateModal(false);
              setGroupName('');
            }
          }}
        >
          <div 
            className="bg-white rounded-2xl p-6 w-96 max-w-[90vw] shadow-2xl animate-in zoom-in-95 duration-200"
            onClick={(e) => e.stopPropagation()}
          >
            <h2 className="text-2xl font-bold mb-6 text-gray-900">Create New Group</h2>
            <div className="mb-6">
              <label className="block text-sm font-semibold text-gray-700 mb-2">
                Group Name
              </label>
              <input
                type="text"
                value={groupName}
                onChange={(e) => setGroupName(e.target.value)}
                onKeyPress={(e) => {
                  if (e.key === 'Enter' && !creating && groupName.trim()) {
                    handleCreateGroup();
                  }
                }}
                placeholder="Enter group name"
                className="w-full px-4 py-3 border-2 border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-[#5a7568] focus:border-[#5a7568] transition-all duration-200"
                autoFocus
                disabled={creating}
              />
            </div>
            <div className="flex gap-3 justify-end">
              <button
                onClick={() => {
                  setShowCreateModal(false);
                  setGroupName('');
                }}
                disabled={creating}
                className="px-5 py-2.5 text-gray-700 bg-gray-100 rounded-xl hover:bg-gray-200 transition-all duration-200 disabled:opacity-50 font-medium"
              >
                Cancel
              </button>
              <button
                onClick={handleCreateGroup}
                disabled={creating || !groupName.trim()}
                className="px-5 py-2.5 bg-[#5a7568] text-white rounded-xl hover:bg-[#4a6558] transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed font-medium shadow-md hover:shadow-lg transform hover:scale-105 active:scale-95 disabled:transform-none"
              >
                {creating ? (
                  <span className="flex items-center gap-2">
                    <svg className="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Creating...
                  </span>
                ) : (
                  'Create'
                )}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Join Group Modal */}
      {showJoinModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-96 max-w-[90vw]">
            <h2 className="text-xl font-semibold mb-4">Join Group</h2>
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Group Chat ID
              </label>
              <input
                type="text"
                value={joinChatId}
                onChange={(e) => setJoinChatId(e.target.value)}
                onKeyPress={(e) => {
                  if (e.key === 'Enter' && !joining) {
                    handleJoinGroup();
                  }
                }}
                placeholder="Enter group chat ID"
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#5a7568] focus:border-transparent"
                autoFocus={!joiningGroupId}
                disabled={joining || !!joiningGroupId}
              />
              {joiningGroupId && (
                <p className="text-xs text-[#5a7568] mt-2 font-medium">
                  Joining: {conversations.find(c => c.chat_id === joiningGroupId)?.name || joiningGroupId}
                </p>
              )}
              {!joiningGroupId && (
                <p className="text-xs text-gray-500 mt-2">
                  Ask the group creator for the chat ID
                </p>
              )}
            </div>
            <div className="flex gap-3 justify-end">
              <button
                onClick={() => {
                  setShowJoinModal(false);
                  setJoinChatId('');
                  setJoiningGroupId(null);
                }}
                disabled={joining}
                className="px-4 py-2 text-gray-700 bg-gray-200 rounded-lg hover:bg-gray-300 transition-colors disabled:opacity-50"
              >
                Cancel
              </button>
              <button
                onClick={handleJoinGroup}
                disabled={joining || (!joinChatId.trim() && !joiningGroupId)}
                className="px-4 py-2 bg-[#5a7568] text-white rounded-lg hover:bg-[#4a6558] transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {joining ? 'Joining...' : 'Join'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}


