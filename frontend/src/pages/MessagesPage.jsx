import React, { useEffect, useState, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getConversations, getConversation } from '../api/messages';

const API_BASE = import.meta.env.VITE_API_URL || "http://localhost:8080";
import { getFriends } from '../api/followers';
import { useAuth } from '../VerifyAuth';
import { useWebSocket } from '../WebSocketContext';
import ConversationCard from '../components/ConversationCard';
import MessageBubble from '../components/MessageBubble';
import { MessageCircle, Send, Search, X } from 'lucide-react';

export default function MessagesPage() {
  const { userId } = useParams();
  const navigate = useNavigate();
  const { user } = useAuth();
  const { sendMessage, lastMessage, isConnected } = useWebSocket();
  const [conversations, setConversations] = useState([]);
  const [friends, setFriends] = useState([]);
  const [messages, setMessages] = useState([]);
  const [activeUser, setActiveUser] = useState(null);
  const [newMessage, setNewMessage] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [showSearch, setShowSearch] = useState(false);
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const messagesEndRef = useRef(null);

  // Send page status when entering/leaving messages page or changing chat partner
  useEffect(() => {
    if (isConnected) {
      // Tell backend we're on messages page
      sendMessage({
        type: 'page_status',
        page: 'messages',
        chat_partner_id: userId ? parseInt(userId) : 0
      });
    }

    // Cleanup: tell backend we left messages page
    return () => {
      if (isConnected) {
        sendMessage({
          type: 'page_status',
          page: 'other',
          chat_partner_id: 0
        });
      }
    };
  }, [isConnected, userId, sendMessage]);

  // Fetch conversations list
  const fetchConversations = async () => {
    try {
      const data = await getConversations();
      setConversations(data);
    } catch (err) {
      console.error('Error fetching conversations:', err);
    }
  };

  // Fetch friends list
  const fetchFriends = async () => {
    try {
      const data = await getFriends();
      setFriends(data);
    } catch (err) {
      console.error('Error fetching friends:', err);
    }
  };

  // Fetch messages for active conversation
  const fetchMessages = async (targetUserId) => {
    try {
      const data = await getConversation(targetUserId);
      setMessages(data);

      // Find user name from conversations or friends
      const conv = conversations.find(c => c.user_id === parseInt(targetUserId));
      const friend = friends.find(f => f.id === parseInt(targetUserId));
      if (conv) {
        setActiveUser({ id: conv.user_id, name: conv.user_name, avatar_url: conv.avatar_url });
      } else if (friend) {
        setActiveUser({ id: friend.id, name: friend.full_name, avatar_url: friend.avatar_url });
      }
    } catch (err) {
      console.error('Error fetching messages:', err);
    }
  };

  // Initial load
  useEffect(() => {
    const init = async () => {
      setLoading(true);
      await Promise.all([fetchConversations(), fetchFriends()]);
      setLoading(false);
    };
    init();
  }, []);

  // Load messages when userId changes
  useEffect(() => {
    if (userId) {
      fetchMessages(userId);
    } else {
      setMessages([]);
      setActiveUser(null);
    }
  }, [userId, conversations, friends]);

  // Handle incoming WebSocket messages
  useEffect(() => {
    if (lastMessage && lastMessage.type === 'chat_message') {
      const msg = lastMessage.message;

      // Add message to current conversation if relevant
      if (userId && (msg.sender_id === parseInt(userId) || msg.receiver_id === parseInt(userId))) {
        setMessages(prev => {
          // Avoid duplicates
          if (prev.some(m => m.id === msg.id)) return prev;
          return [...prev, msg];
        });
      }

      // Refresh conversations list
      fetchConversations();
    }
  }, [lastMessage, userId]);

  // Scroll to bottom when messages change
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  // Send message via WebSocket
  const handleSend = async (e) => {
    e.preventDefault();
    if (!newMessage.trim() || !userId) return;

    setSending(true);
    try {
      const sent = sendMessage({
        type: 'chat_message',
        to: parseInt(userId),
        content: newMessage.trim()
      });

      if (sent) {
        setNewMessage('');
      }
    } catch (err) {
      console.error('Error sending message:', err);
    } finally {
      setSending(false);
    }
  };

  const handleStartChat = (friendId) => {
    setShowSearch(false);
    setSearchQuery('');
    navigate(`/messages/${friendId}`);
  };

  const getInitial = (name) => {
    return name ? name.charAt(0).toUpperCase() : '?';
  };

  // Filter friends based on search
  const filteredFriends = friends.filter(f =>
    f.full_name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
    f.nickname?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="main-content">
      <div style={{
        display: 'grid',
        gridTemplateColumns: '320px 1fr',
        gap: '24px',
        height: 'calc(100vh - 280px)',
        minHeight: '500px'
      }}>
        {/* Left Panel - Conversations List */}
        <div className="card" style={{ display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
          <div className="card-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span className="card-header-title">Messages</span>
            <button
              onClick={() => setShowSearch(!showSearch)}
              style={{
                background: 'none',
                border: 'none',
                cursor: 'pointer',
                padding: '4px',
                display: 'flex',
                alignItems: 'center'
              }}
            >
              {showSearch ? (
                <X size={18} color="var(--coffee-bean)" />
              ) : (
                <Search size={18} color="var(--coffee-bean)" />
              )}
            </button>
          </div>

          {/* Search Bar */}
          {showSearch && (
            <div style={{ padding: '12px 16px', borderBottom: '1px solid rgba(23, 3, 18, 0.08)' }}>
              <input
                type="text"
                className="form-input"
                placeholder="Search friends..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                style={{ width: '100%' }}
              />

              {/* Search Results - Friends you can message */}
              {searchQuery && (
                <div style={{ marginTop: '8px' }}>
                  {filteredFriends.length === 0 ? (
                    <p style={{
                      fontSize: '12px',
                      color: 'var(--jet-black)',
                      opacity: 0.6,
                      padding: '8px 0'
                    }}>
                      No friends found. You can only message mutual friends.
                    </p>
                  ) : (
                    filteredFriends.map(friend => (
                      <div
                        key={friend.id}
                        onClick={() => handleStartChat(friend.id)}
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: '10px',
                          padding: '10px 8px',
                          cursor: 'pointer',
                          borderRadius: '8px',
                          transition: 'background-color 0.2s'
                        }}
                        onMouseEnter={(e) => e.currentTarget.style.backgroundColor = 'rgba(45, 81, 149, 0.08)'}
                        onMouseLeave={(e) => e.currentTarget.style.backgroundColor = 'transparent'}
                      >
                        {(() => {
                          const avatarSrc = friend.avatar_url
                            ? (friend.avatar_url.startsWith('http')
                              ? friend.avatar_url
                              : `${API_BASE}${friend.avatar_url}`)
                            : null;
                          return (
                            <div style={{
                              width: '36px',
                              height: '36px',
                              borderRadius: '50%',
                              backgroundColor: avatarSrc ? 'transparent' : 'var(--dusk-blue)',
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center',
                              color: 'var(--white)',
                              fontSize: '14px',
                              fontWeight: 600
                            }}>
                              {avatarSrc ? (
                                <img
                                  src={avatarSrc}
                                  alt="avatar"
                                  style={{ width: '100%', height: '100%', borderRadius: '50%' }}
                                />
                              ) : (
                                getInitial(friend.full_name)
                              )}
                            </div>
                          );
                        })()}
                        <span style={{
                          fontSize: '13px',
                          fontWeight: 500,
                          color: 'var(--coffee-bean)'
                        }}>
                          {friend.full_name}
                        </span>
                      </div>
                    ))
                  )}
                </div>
              )}
            </div>
          )}

          <div style={{ flex: 1, overflowY: 'auto' }}>
            {loading ? (
              <div style={{ padding: '20px', textAlign: 'center' }}>
                <p style={{ opacity: 0.6, fontFamily: "'Cormorant Garamond', serif" }}>Loading...</p>
              </div>
            ) : conversations.length === 0 && !showSearch ? (
              <div style={{ padding: '20px', textAlign: 'center' }}>
                <p style={{ opacity: 0.6, fontFamily: "'Cormorant Garamond', serif", marginBottom: '12px' }}>
                  No conversations yet
                </p>
                <button
                  className="btn btn-outline"
                  onClick={() => setShowSearch(true)}
                  style={{ fontSize: '12px', padding: '8px 16px' }}
                >
                  <Search size={14} style={{ marginRight: '6px' }} />
                  Find friends to chat
                </button>
              </div>
            ) : (
              conversations.map(conv => (
                <ConversationCard key={conv.user_id} conversation={conv} />
              ))
            )}
          </div>
        </div>

        {/* Right Panel - Active Conversation */}
        <div className="card" style={{ display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
          {!userId ? (
            <div style={{
              flex: 1,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              flexDirection: 'column',
              gap: '16px'
            }}>
              <MessageCircle size={48} color="var(--dusk-blue)" strokeWidth={1.5} />
              <p style={{
                opacity: 0.6,
                fontFamily: "'Cormorant Garamond', serif",
                fontSize: '16px'
              }}>
                Select a conversation or search for a friend
              </p>
            </div>
          ) : (
            <>
              {/* Chat Header */}
              <div style={{
                padding: '16px 20px',
                borderBottom: '1px solid rgba(23, 3, 18, 0.1)',
                display: 'flex',
                alignItems: 'center',
                gap: '12px'
              }}>
                {(() => {
                  const avatarSrc = activeUser?.avatar_url
                    ? (activeUser.avatar_url.startsWith('http')
                      ? activeUser.avatar_url
                      : `${API_BASE}${activeUser.avatar_url}`)
                    : null;
                  return (
                    <div style={{
                      width: '40px',
                      height: '40px',
                      borderRadius: '50%',
                      backgroundColor: avatarSrc ? 'transparent' : 'var(--dusk-blue)',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      color: 'var(--white)',
                      fontSize: '16px',
                      fontWeight: 600
                    }}>
                      {avatarSrc ? (
                        <img
                          src={avatarSrc}
                          alt="avatar"
                          style={{ width: '100%', height: '100%', borderRadius: '50%' }}
                        />
                      ) : (
                        getInitial(activeUser?.name)
                      )}
                    </div>
                  );
                })()}
                <div>
                  <h3 style={{
                    margin: 0,
                    fontSize: '14px',
                    fontWeight: 600,
                    color: 'var(--coffee-bean)',
                    fontFamily: "'Montserrat', sans-serif"
                  }}>
                    {activeUser?.name || 'Loading...'}
                  </h3>
                </div>
              </div>

              {/* Messages Area */}
              <div style={{
                flex: 1,
                overflowY: 'auto',
                padding: '20px',
                backgroundColor: 'rgba(245, 243, 239, 0.5)'
              }}>
                {messages.length === 0 ? (
                  <div style={{ textAlign: 'center', paddingTop: '40px' }}>
                    <p style={{
                      opacity: 0.6,
                      fontFamily: "'Cormorant Garamond', serif"
                    }}>
                      No messages yet. Say hello!
                    </p>
                  </div>
                ) : (
                  messages.map(msg => (
                    <MessageBubble
                      key={msg.id}
                      message={msg}
                      isOwn={msg.sender_id === user?.id}
                    />
                  ))
                )}
                <div ref={messagesEndRef} />
              </div>

              {/* Message Input */}
              <form onSubmit={handleSend} style={{
                padding: '16px 20px',
                borderTop: '1px solid rgba(23, 3, 18, 0.1)',
                display: 'flex',
                gap: '12px'
              }}>
                <input
                  type="text"
                  className="form-input"
                  placeholder="Type a message..."
                  value={newMessage}
                  onChange={(e) => setNewMessage(e.target.value)}
                  disabled={sending}
                  style={{ flex: 1 }}
                />
                <button
                  type="submit"
                  className="btn btn-primary"
                  disabled={sending || !newMessage.trim()}
                  style={{ display: 'flex', alignItems: 'center', gap: '8px' }}
                >
                  Send
                  <Send size={16} />
                </button>
              </form>
            </>
          )}
        </div>
      </div>
    </div>
  );
}