import React, { useState } from 'react';
import { getGroupPostComments, addGroupPostComment } from '../api/grouppostcomments';
import { MessageCircle } from 'lucide-react';

export default function GroupPostCard({ post, groupId, userId }) {
    const [showComments, setShowComments] = useState(false);
    const [comments, setComments] = useState([]);
    const [newComment, setNewComment] = useState('');
    const [loading, setLoading] = useState(false);
    const [commentsLoading, setCommentsLoading] = useState(false);
    const [commentCount, setCommentCount] = useState(post.comment_count || 0);

    const fetchComments = async () => {
        setCommentsLoading(true);
        try {
            const data = await getGroupPostComments(groupId, post.id);
            setComments(data || []);
            setCommentCount(data?.length || 0);
        } catch (err) {
            console.error('Error fetching comments:', err);
        } finally {
            setCommentsLoading(false);
        }
    };

    const handleShowComments = async () => {
        if (!showComments) {
            await fetchComments();
        }
        setShowComments(!showComments);
    };

    const handleAddComment = async (e) => {
        e.preventDefault();
        if (!newComment.trim()) return;

        setLoading(true);
        try {
            const comment = await addGroupPostComment(groupId, post.id, newComment.trim());
            setComments(prev => [...prev, comment]);
            setCommentCount(prev => prev + 1);
            setNewComment('');
        } catch (err) {
            console.error('Error adding comment:', err);
            alert('Failed to add comment');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="card" style={{ marginBottom: 12 }}>
            <div className="card-body">
                <div style={{ fontSize: 12, opacity: 0.75, marginBottom: 6 }}>
                    {post.author_name} • {post.created_at ? new Date(post.created_at).toLocaleString() : ""}
                </div>
                <div style={{ marginBottom: 16 }}>{post.content}</div>

                {/* Comments toggle button */}
                <button
                    onClick={handleShowComments}
                    style={{
                        background: 'none',
                        border: 'none',
                        cursor: 'pointer',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '6px',
                        color: 'var(--dusk-blue)',
                        fontSize: '14px',
                        padding: '0',
                        marginBottom: showComments ? '12px' : '0'
                    }}
                >
                    <MessageCircle size={16} />
                    Comments ({commentCount})
                </button>

                {/* Load and display comments */}
                {showComments && (
                    <div style={{ marginTop: 12, borderTop: '1px solid rgba(23, 3, 18, 0.1)', paddingTop: 12 }}>
                        {commentsLoading ? (
                            <p style={{ opacity: 0.6, fontSize: 13 }}>Loading comments...</p>
                        ) : comments.length === 0 ? (
                            <p style={{ opacity: 0.6, fontSize: 13 }}>No comments yet</p>
                        ) : (
                            comments.map((comment) => (
                                <div
                                    key={comment.id}
                                    style={{
                                        marginBottom: 12,
                                        padding: '10px 12px',
                                        backgroundColor: 'rgba(245, 243, 239, 0.3)',
                                        borderRadius: '6px'
                                    }}
                                >
                                    <div style={{ fontSize: 11, fontWeight: 600, marginBottom: 4, color: 'var(--dusk-blue)' }}>
                                        {comment.author_name}
                                    </div>
                                    {comment.image_path && (
                                        <img
                                            src={comment.image_path}
                                            alt="comment attachment"
                                            style={{
                                                maxWidth: '100%',
                                                maxHeight: '120px',
                                                borderRadius: '4px',
                                                marginBottom: comment.content ? '6px' : 0,
                                                display: 'block'
                                            }}
                                        />
                                    )}
                                    {comment.content && (
                                        <p style={{ margin: 0, fontSize: 13, lineHeight: 1.4 }}>
                                            {comment.content}
                                        </p>
                                    )}
                                    <div style={{ fontSize: 10, opacity: 0.5, marginTop: 4 }}>
                                        {new Date(comment.created_at).toLocaleString()}
                                    </div>
                                </div>
                            ))
                        )}

                        {/* Comment form */}
                        <form onSubmit={handleAddComment} style={{ marginTop: 12, display: 'flex', gap: '8px' }}>
                            <input
                                type="text"
                                className="form-input"
                                placeholder="Add a comment..."
                                value={newComment}
                                onChange={(e) => setNewComment(e.target.value)}
                                disabled={loading}
                                style={{ flex: 1, fontSize: '13px' }}
                            />
                            <button
                                type="submit"
                                className="btn btn-primary"
                                disabled={loading || !newComment.trim()}
                                style={{ fontSize: '13px', padding: '6px 12px' }}
                            >
                                {loading ? '...' : 'Reply'}
                            </button>
                        </form>
                    </div>
                )}
            </div>
        </div>
    );
}
