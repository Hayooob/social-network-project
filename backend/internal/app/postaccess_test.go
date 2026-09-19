package app

import (
	"fmt"
	"net/http"
	"testing"

	"social-network/internal/db"
)

// These tests cover post visibility on the per-post endpoints. The feed filters
// posts by privacy, but /api/posts/{id}/comments and /api/posts/{id}/like take
// an arbitrary post ID, so without their own check they are a way around it.

func commentsPath(postID int) string { return fmt.Sprintf("/api/posts/%d/comments", postID) }
func likePath(postID int) string     { return fmt.Sprintf("/api/posts/%d/like", postID) }

// --- reading comments -------------------------------------------------------

func TestStrangerCannotReadCommentsOnPrivatePost(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	mallory := env.registerUser("mallory@example.com", "Mallory")

	post := env.createPost(alice, "private thoughts", "private")

	resp := env.do(mallory, http.MethodGet, commentsPath(post.ID), nil)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusNotFound, "stranger reading comments on a private post")
}

func TestStrangerCannotReadCommentsOnAlmostPrivatePost(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	mallory := env.registerUser("mallory@example.com", "Mallory")

	post := env.createPost(alice, "followers only", "almost-private")

	resp := env.do(mallory, http.MethodGet, commentsPath(post.ID), nil)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusNotFound, "non-follower reading comments on an almost-private post")
}

func TestFollowerCanReadCommentsOnAlmostPrivatePost(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	bob := env.registerUser("bob@example.com", "Bob")
	env.follow(bob, alice, "accepted")

	post := env.createPost(alice, "followers only", "almost-private")

	resp := env.do(bob, http.MethodGet, commentsPath(post.ID), nil)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusOK, "follower reading comments on an almost-private post")
}

func TestPendingFollowerCannotReadAlmostPrivatePost(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	bob := env.registerUser("bob@example.com", "Bob")
	env.setPrivate(alice, true)
	env.follow(bob, alice, "pending")

	post := env.createPost(alice, "followers only", "almost-private")

	resp := env.do(bob, http.MethodGet, commentsPath(post.ID), nil)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusNotFound, "pending follower reading an almost-private post")
}

func TestNamedViewerCanReadPrivatePost(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	bob := env.registerUser("bob@example.com", "Bob")

	post := env.createPost(alice, "just for bob", "private", int(bob.ID))

	resp := env.do(bob, http.MethodGet, commentsPath(post.ID), nil)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusOK, "named viewer reading a private post")
}

func TestAuthorCanAlwaysReadOwnPost(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")

	post := env.createPost(alice, "private thoughts", "private")

	resp := env.do(alice, http.MethodGet, commentsPath(post.ID), nil)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusOK, "author reading comments on their own private post")
}

// --- writing comments -------------------------------------------------------

func TestStrangerCannotCommentOnPrivatePost(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	mallory := env.registerUser("mallory@example.com", "Mallory")

	post := env.createPost(alice, "private thoughts", "private")

	resp := env.do(mallory, http.MethodPost, commentsPath(post.ID), map[string]string{
		"content": "I should not be able to write here",
	})
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusNotFound, "stranger commenting on a private post")

	var count int
	if err := env.Server.DB.QueryRow(
		`SELECT COUNT(*) FROM post_comments WHERE post_id = ?`, post.ID,
	).Scan(&count); err != nil {
		t.Fatalf("count comments: %v", err)
	}
	if count != 0 {
		t.Errorf("%d comment(s) were stored on a post the author did not share", count)
	}
}

// TestCannotCommentOnNonexistentPost guards against orphan rows: before the
// visibility check existed, a comment on any post ID at all was accepted.
func TestCannotCommentOnNonexistentPost(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")

	resp := env.do(alice, http.MethodPost, commentsPath(999999), map[string]string{
		"content": "orphan comment",
	})
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusNotFound, "commenting on a post that does not exist")
}

func TestCommentOnVisiblePostIsStored(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	bob := env.registerUser("bob@example.com", "Bob")

	post := env.createPost(alice, "hello world", "public")

	resp := env.do(bob, http.MethodPost, commentsPath(post.ID), map[string]string{
		"content": "nice post",
	})
	assertStatus(t, resp, http.StatusCreated, "commenting on a public post")

	var comment map[string]any
	decodeJSON(t, resp, &comment)

	if comment["content"] != "nice post" {
		t.Errorf("got content %v, want %q", comment["content"], "nice post")
	}
	if comment["author_name"] != "Bob" {
		t.Errorf("got author %v, want Bob", comment["author_name"])
	}
}

// --- likes ------------------------------------------------------------------

func TestStrangerCannotLikePrivatePost(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	mallory := env.registerUser("mallory@example.com", "Mallory")

	post := env.createPost(alice, "private thoughts", "private")

	resp := env.do(mallory, http.MethodPost, likePath(post.ID), nil)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusNotFound, "stranger liking a private post")

	var count int
	if err := env.Server.DB.QueryRow(
		`SELECT COUNT(*) FROM post_likes WHERE post_id = ?`, post.ID,
	).Scan(&count); err != nil {
		t.Fatalf("count likes: %v", err)
	}
	if count != 0 {
		t.Errorf("%d like(s) were stored on a post the author did not share", count)
	}
}

func TestCannotLikeNonexistentPost(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")

	resp := env.do(alice, http.MethodPost, likePath(999999), nil)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusNotFound, "liking a post that does not exist")
}

func TestLikeTogglesOnAndOff(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	bob := env.registerUser("bob@example.com", "Bob")

	post := env.createPost(alice, "hello world", "public")

	first := env.do(bob, http.MethodPost, likePath(post.ID), nil)
	assertStatus(t, first, http.StatusOK, "first like")

	var body map[string]any
	decodeJSON(t, first, &body)
	if body["liked"] != true {
		t.Error("first like did not report liked=true")
	}
	if body["like_count"].(float64) != 1 {
		t.Errorf("got like_count %v, want 1", body["like_count"])
	}

	second := env.do(bob, http.MethodPost, likePath(post.ID), nil)
	assertStatus(t, second, http.StatusOK, "second like")

	decodeJSON(t, second, &body)
	if body["liked"] != false {
		t.Error("liking twice did not unlike")
	}
	if body["like_count"].(float64) != 0 {
		t.Errorf("got like_count %v, want 0", body["like_count"])
	}
}

// --- CanViewPost directly ---------------------------------------------------

func TestCanViewPostRules(t *testing.T) {
	env := newTestEnv(t)
	author := env.registerUser("author@example.com", "Author")
	follower := env.registerUser("follower@example.com", "Follower")
	stranger := env.registerUser("stranger@example.com", "Stranger")
	named := env.registerUser("named@example.com", "Named Viewer")

	env.follow(follower, author, "accepted")

	publicPost := env.createPost(author, "public", "public")
	almostPost := env.createPost(author, "almost", "almost-private")
	privatePost := env.createPost(author, "private", "private", int(named.ID))

	cases := []struct {
		name   string
		postID int
		viewer *testUser
		want   bool
	}{
		{"author sees own public post", publicPost.ID, author, true},
		{"author sees own almost-private post", almostPost.ID, author, true},
		{"author sees own private post", privatePost.ID, author, true},
		{"stranger sees public post", publicPost.ID, stranger, true},
		{"stranger cannot see almost-private post", almostPost.ID, stranger, false},
		{"stranger cannot see private post", privatePost.ID, stranger, false},
		{"follower sees public post", publicPost.ID, follower, true},
		{"follower sees almost-private post", almostPost.ID, follower, true},
		{"follower cannot see private post they are not named in", privatePost.ID, follower, false},
		{"named viewer sees private post", privatePost.ID, named, true},
		{"nonexistent post is never visible", 999999, author, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := db.CanViewPost(env.Server.DB, tc.postID, int(tc.viewer.ID))
			if err != nil {
				t.Fatalf("CanViewPost: %v", err)
			}
			if got != tc.want {
				t.Errorf("CanViewPost = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestPrivateProfilePublicPostHiddenFromStrangers covers the interaction
// between profile privacy and post privacy: a public post by a private account
// is not public.
func TestPrivateProfilePublicPostHiddenFromStrangers(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	stranger := env.registerUser("stranger@example.com", "Stranger")
	env.setPrivate(alice, true)

	post := env.createPost(alice, "public post, private account", "public")

	visible, err := db.CanViewPost(env.Server.DB, post.ID, int(stranger.ID))
	if err != nil {
		t.Fatalf("CanViewPost: %v", err)
	}
	if visible {
		t.Error("a public post by a private account is visible to a stranger")
	}
}
