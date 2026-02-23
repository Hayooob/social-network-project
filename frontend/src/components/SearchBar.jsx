import React, { useState, useEffect, useRef } from 'react';
import { Link } from 'react-router-dom';
import { searchUsers } from '../api/users';
import { Search, X, Lock } from 'lucide-react';

export default function SearchBar() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [isOpen, setIsOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const wrapperRef = useRef(null);
  const inputRef = useRef(null);
  const debounceRef = useRef(null);

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event) {
      if (wrapperRef.current && !wrapperRef.current.contains(event.target)) {
        setIsOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // Debounced search
  useEffect(() => {
    if (debounceRef.current) {
      clearTimeout(debounceRef.current);
    }

    if (query.length < 1) {
      setResults([]);
      setIsOpen(false);
      return;
    }

    setLoading(true);
    debounceRef.current = setTimeout(async () => {
      try {
        const data = await searchUsers(query);
        setResults(data);
        setIsOpen(true);
      } catch (err) {
        console.error('Search error:', err);
        setResults([]);
      } finally {
        setLoading(false);
      }
    }, 300);

    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
      }
    };
  }, [query]);

  const handleClear = () => {
    setQuery('');
    setResults([]);
    setIsOpen(false);
    inputRef.current?.focus();
  };

  const getInitial = (name) => {
    return name ? name.charAt(0).toUpperCase() : '?';
  };

  return (
    <div ref={wrapperRef} style={{ position: 'relative', marginBottom: '16px' }}>
      {/* Search Input - Square Design */}
      <div style={{
        display: 'flex',
        alignItems: 'center',
        backgroundColor: 'var(--soft-linen)',
        borderRadius: '8px',
        padding: '10px 14px',
        border: '1px solid rgba(23, 3, 18, 0.1)',
      }}>
        <Search size={16} color="var(--jet-black)" style={{ opacity: 0.4, marginRight: '10px', flexShrink: 0 }} />
        <input
          ref={inputRef}
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onFocus={() => query.length >= 1 && results.length > 0 && setIsOpen(true)}
          placeholder="Search users..."
          style={{
            border: 'none',
            background: 'transparent',
            outline: 'none',
            fontSize: '13px',
            fontFamily: "'Montserrat', sans-serif",
            color: 'var(--jet-black)',
            width: '100%'
          }}
        />
        {query && (
          <button
            onClick={handleClear}
            style={{
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              padding: '2px',
              display: 'flex',
              alignItems: 'center',
              flexShrink: 0
            }}
          >
            <X size={14} color="var(--jet-black)" style={{ opacity: 0.4 }} />
          </button>
        )}
      </div>

      {/* Results Dropdown */}
      {isOpen && (
        <div style={{
          position: 'absolute',
          top: '100%',
          left: 0,
          right: 0,
          marginTop: '4px',
          backgroundColor: 'var(--white)',
          borderRadius: '8px',
          boxShadow: '0 4px 20px rgba(23, 3, 18, 0.15)',
          border: '1px solid rgba(23, 3, 18, 0.08)',
          overflow: 'hidden',
          zIndex: 1000,
          maxHeight: '280px',
          overflowY: 'auto'
        }}>
          {loading ? (
            <div style={{ padding: '16px', textAlign: 'center' }}>
              <span style={{ 
                fontSize: '13px', 
                color: 'var(--jet-black)', 
                opacity: 0.6,
                fontFamily: "'Cormorant Garamond', serif"
              }}>
                Searching...
              </span>
            </div>
          ) : results.length === 0 ? (
            <div style={{ padding: '16px', textAlign: 'center' }}>
              <span style={{ 
                fontSize: '13px', 
                color: 'var(--jet-black)', 
                opacity: 0.6,
                fontFamily: "'Cormorant Garamond', serif"
              }}>
                No users found
              </span>
            </div>
          ) : (
            results.map((user, index) => (
              <Link
                key={user.id}
                to={`/users/${user.id}`}
                onClick={() => {
                  setQuery('');
                  setResults([]);
                  setIsOpen(false);
                }}
                style={{ textDecoration: 'none', color: 'inherit' }}
              >
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '10px',
                    padding: '10px 14px',
                    cursor: 'pointer',
                    borderBottom: index < results.length - 1 ? '1px solid rgba(23, 3, 18, 0.05)' : 'none',
                    transition: 'background-color 0.15s ease'
                  }}
                  onMouseEnter={(e) => e.currentTarget.style.backgroundColor = 'rgba(45, 81, 149, 0.06)'}
                  onMouseLeave={(e) => e.currentTarget.style.backgroundColor = 'transparent'}
                >
                  {/* Avatar */}
                  <div style={{
                    width: '32px',
                    height: '32px',
                    borderRadius: '50%',
                    backgroundColor: index % 3 === 0 ? 'var(--dusk-blue)' : index % 3 === 1 ? 'var(--blush-rose)' : 'var(--jet-black)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    color: 'var(--white)',
                    fontSize: '13px',
                    fontWeight: 600,
                    flexShrink: 0
                  }}>
                    {user.avatar_url ? (
                      <img 
                        src={user.avatar_url} 
                        alt="" 
                        style={{ width: '100%', height: '100%', borderRadius: '50%', objectFit: 'cover' }}
                      />
                    ) : (
                      getInitial(user.full_name)
                    )}
                  </div>

                  {/* User Info */}
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ 
                      display: 'flex', 
                      alignItems: 'center', 
                      gap: '4px'
                    }}>
                      <span style={{
                        fontSize: '13px',
                        fontWeight: 600,
                        color: 'var(--coffee-bean)',
                        fontFamily: "'Montserrat', sans-serif",
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        whiteSpace: 'nowrap'
                      }}>
                        {user.full_name}
                      </span>
                      {user.is_private && (
                        <Lock size={10} color="var(--jet-black)" style={{ opacity: 0.4, flexShrink: 0 }} />
                      )}
                    </div>
                    {user.nickname && (
                      <span style={{
                        fontSize: '11px',
                        color: 'var(--jet-black)',
                        opacity: 0.5,
                        fontFamily: "'Montserrat', sans-serif"
                      }}>
                        @{user.nickname}
                      </span>
                    )}
                  </div>
                </div>
              </Link>
            ))
          )}
        </div>
      )}
    </div>
  );
}