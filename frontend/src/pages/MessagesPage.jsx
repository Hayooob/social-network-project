import React, { useEffect, useState, useRef } from 'react';
import { useParams } from 'react-router-dom';
import { getConversations, getConversation } from '../api/messages';
import { useAuth } from '../VerifyAuth';
import ConversationCard from '../components/ConversationCard';
import MessageBubble from '../components/MessageBubble';

export default function MessagesPage() {
  const { userId } = useParams();
  const { user } = useAuth();
  const [conversations, setConversations] = useState([]);
  const [messages, setMessages] = useState([]);
  const [activeUser, setActiveUser] = useState(null);
  const [newMessage, setNewMessage] = useState('');
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const messagesEndRef = useRef(null);
  const wsRef = useRef(null);

  // Fetch conversations list
  const fetchConversations = async () => {
    try {
      const data = await getConversations();
      setConversations(data);
    } catch (err) {
      console.error('Error fetching conversations:', err);
    }
  };

  // Fetch messages for active conversation
  const fetchMessages = async (targetUserId) => {
    try {
      const data = await getConversation(targetUserId);
      setMessages(data);
      
      // Find user name from conversations
      const conv = conversations.find(c => c.user_id === parseInt(targetUserId));
      if (conv) {
        setActiveUser({ id: conv.user_id, name: conv.user_name });
      }
    } catch (err) {
      console.error('Error fetching messages:', err);
    }
  };

  // Initial load
  useEffect(() => {
    const init = async () => {
      setLoading(true);
      await fetchConversations();
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
  }, [userId, conversations]);

  // Setup WebSocket connection
  useEffect(() => {
    if (!user) return;

    const ws = new WebSocket('ws://localhost:8080/ws');
    wsRef.current = ws;

    ws.onopen = () => {
      console.log('WebSocket connected');
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        
        if (data.type === 'chat_message') {
          const msg = data.message;
          
          // If this message is part of the active conversation, add it
          if (userId && (msg.sender_id === parseInt(userId) || msg.receiver_id === parseInt(userId))) {
            setMessages(prev => [...prev, msg]);
          }
          
          // Refresh conversations to update last message
          fetchConversations();
        }
      } catch (err) {
        console.error('WebSocket message error:', err);
      }
    };

    ws.onclose = () => {
      console.log('WebSocket disconnected');
    };

    return () => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.close();
      }
    };
  }, [user, userId]);

  // Scroll to bottom when messages change
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  // Send message via WebSocket
  const handleSend = async (e) => {
    e.preventDefault();
    if (!newMessage.trim() || !userId || !wsRef.current) return;

    setSending(true);
    try {
      const ws = wsRef.current;
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
          type: 'chat_message',
          to: parseInt(userId),
          content: newMessage.trim()
        }));
        setNewMessage('');
      }
    } catch (err) {
      console.error('Error sending message:', err);
    } finally {
      setSending(false);
    }
  };

  const getInitial = (name) => {
    return name ? name.charAt(0).toUpperCase() : '?';
  };

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
          <div className="card-header">
            <span className="card-header-title">Messages</span>
            <span className="card-header-star star-spin">✦</span>
          </div>
          
          <div style={{ flex: 1, overflowY: 'auto' }}>
            {loading ? (
              <div style={{ padding: '20px', textAlign: 'center' }}>
                <p style={{ opacity: 0.6, fontFamily: "'Cormorant Garamond', serif" }}>Loading...</p>
              </div>
            ) : conversations.length === 0 ? (
              <div style={{ padding: '20px', textAlign: 'center' }}>
                <p style={{ opacity: 0.6, fontFamily: "'Cormorant Garamond', serif" }}>No conversations yet</p>
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
              <span style={{ fontSize: '48px' }}>💬</span>
              <p style={{ 
                opacity: 0.6, 
                fontFamily: "'Cormorant Garamond', serif",
                fontSize: '16px'
              }}>
                Select a conversation to start chatting
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
                <div style={{
                  width: '40px',
                  height: '40px',
                  borderRadius: '50%',
                  backgroundColor: 'var(--dusk-blue)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  color: 'var(--white)',
                  fontSize: '16px',
                  fontWeight: 600
                }}>
                  {getInitial(activeUser?.name)}
                </div>
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
                >
                  Send
                  <span>↗</span>
                </button>
              </form>
            </>
          )}
        </div>
      </div>
    </div>
  );
}