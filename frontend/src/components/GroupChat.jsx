import React, { useEffect, useState, useRef } from 'react';
import { getGroupMessages, sendGroupMessage } from '../api/groupmessages';
import { useWebSocket } from '../WebSocketContext';
import { useAuth } from '../VerifyAuth';
import { Send, MessageCircle } from 'lucide-react';

export default function GroupChat({ groupId }) {
  const { user } = useAuth();
  const { lastMessage, isConnected } = useWebSocket();
  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState('');
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const messagesEndRef = useRef(null);

  // Fetch messages
  const fetchMessages = async () => {
    try {
      const data = await getGroupMessages(groupId);
      setMessages(data);
    } catch (err) {
      console.error('Error fetching group messages:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchMessages();
  }, [groupId]);

  // Handle incoming WebSocket messages
  useEffect(() => {
    if (lastMessage && lastMessage.type === 'group_message' && lastMessage.group_id === groupId) {
      const msg = lastMessage.message;
      setMessages(prev => {
        if (prev.some(m => m.id === msg.id)) return prev;
        return [...prev, msg];
      });
    }
  }, [lastMessage, groupId]);

  // Scroll to bottom when messages change
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSend = async (e) => {
    e.preventDefault();
    if (!newMessage.trim()) return;

    setSending(true);
    try {
      const msg = await sendGroupMessage(groupId, newMessage.trim());
      setMessages(prev => [...prev, msg]);
      setNewMessage('');
    } catch (err) {
      console.error('Error sending group message:', err);
      alert('Failed to send message');
    } finally {
      setSending(false);
    }
  };

  const getInitial = (name) => name ? name.charAt(0).toUpperCase() : '?';

  const formatTime = (timestamp) => {
    try {
      return new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } catch {
      return '';
    }
  };

  return (
    <div className="card" style={{ display: 'flex', flexDirection: 'column', height: '400px' }}>
      <div className="card-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <span className="card-header-title">
          <MessageCircle size={16} style={{ marginRight: '8px' }} />
          Group Chat
        </span>
        <span style={{
          width: '8px',
          height: '8px',
          borderRadius: '50%',
          backgroundColor: isConnected ? '#22c55e' : '#ef4444'
        }} title={isConnected ? 'Connected' : 'Disconnected'} />
      </div>

      {/* Messages Area */}
      <div style={{
        flex: 1,
        overflowY: 'auto',
        padding: '16px',
        backgroundColor: 'rgba(245, 243, 239, 0.5)'
      }}>
        {loading ? (
          <div style={{ textAlign: 'center', paddingTop: '20px' }}>
            <p style={{ opacity: 0.6 }}>Loading messages...</p>
          </div>
        ) : messages.length === 0 ? (
          <div style={{ textAlign: 'center', paddingTop: '40px' }}>
            <p style={{ opacity: 0.6 }}>No messages yet. Start the conversation!</p>
          </div>
        ) : (
          messages.map((msg, index) => {
            const isOwn = msg.sender_id === user?.id;
            return (
              <div 
                key={msg.id} 
                style={{
                  display: 'flex',
                  flexDirection: isOwn ? 'row-reverse' : 'row',
                  alignItems: 'flex-end',
                  gap: '8px',
                  marginBottom: '12px'
                }}
              >
                {!isOwn && (
                  <div style={{
                    width: '28px',
                    height: '28px',
                    borderRadius: '50%',
                    backgroundColor: index % 3 === 0 ? 'var(--dusk-blue)' : index % 3 === 1 ? 'var(--blush-rose)' : 'var(--jet-black)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    color: 'var(--white)',
                    fontSize: '12px',
                    fontWeight: 600
                  }}>
                    {getInitial(msg.sender_name)}
                  </div>
                )}

                <div style={{
                  maxWidth: '70%',
                  padding: '10px 14px',
                  borderRadius: isOwn ? '16px 16px 4px 16px' : '16px 16px 16px 4px',
                  backgroundColor: isOwn ? 'var(--dusk-blue)' : 'var(--white)',
                  color: isOwn ? 'var(--white)' : 'var(--jet-black)',
                  boxShadow: '0 1px 2px rgba(0,0,0,0.1)'
                }}>
                  {!isOwn && (
                    <div style={{ fontSize: '11px', fontWeight: 600, marginBottom: '4px', color: 'var(--dusk-blue)' }}>
                      {msg.sender_name}
                    </div>
                  )}
                  <p style={{ margin: 0, fontSize: '14px', lineHeight: 1.4, whiteSpace: 'pre-wrap' }}>
                    {msg.content}
                  </p>
                  <div style={{ fontSize: '10px', opacity: 0.6, marginTop: '4px', textAlign: 'right' }}>
                    {formatTime(msg.created_at)}
                  </div>
                </div>
              </div>
            );
          })
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Message Input */}
      <form onSubmit={handleSend} style={{
        padding: '12px 16px',
        borderTop: '1px solid rgba(23, 3, 18, 0.1)',
        display: 'flex',
        gap: '10px'
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
          <Send size={14} />
        </button>
      </form>
    </div>
  );
}