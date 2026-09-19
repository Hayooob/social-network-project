package app

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"testing"

	"social-network/internal/db"
)

// --- post creation ----------------------------------------------------------

// TestCreatePostReturnsCreatedPost guards a bug where GetPostByID selected
// seven columns but scanned eight, so every POST /api/posts answered
// 500 "failed to create post" even though the row had already been committed.
func TestCreatePostReturnsCreatedPost(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")

	resp := env.do(alice, http.MethodPost, "/api/posts", map[string]any{
		"content": "hello world",
		"privacy": "public",
	})
	assertStatus(t, resp, http.StatusCreated, "POST /api/posts")

	var post map[string]any
	decodeJSON(t, resp, &post)

	if post["content"] != "hello world" {
		t.Errorf("got content %v, want %q", post["content"], "hello world")
	}
	if id, ok := post["id"].(float64); !ok || id == 0 {
		t.Errorf("got post id %v, want a real ID", post["id"])
	}
	if post["author_name"] != "Alice" {
		t.Errorf("got author_name %v, want Alice", post["author_name"])
	}
}

func TestCreatePostRejectsUnknownPrivacyValue(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")

	resp := env.do(alice, http.MethodPost, "/api/posts", map[string]any{
		"content": "hello world",
		"privacy": "sort-of-secret",
	})
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusBadRequest, "POST /api/posts with an invalid privacy value")

	// A post with an unrecognised privacy value matches none of the visibility
	// queries, so it would be invisible to everyone including its author.
	var count int
	if err := env.Server.DB.QueryRow(`SELECT COUNT(*) FROM posts`).Scan(&count); err != nil {
		t.Fatalf("count posts: %v", err)
	}
	if count != 0 {
		t.Errorf("%d post(s) were stored with an invalid privacy value", count)
	}
}

func TestCreatePostRequiresContent(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")

	resp := env.do(alice, http.MethodPost, "/api/posts", map[string]any{
		"content": "   ",
		"privacy": "public",
	})
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusBadRequest, "POST /api/posts with blank content")
}

// TestOwnProfilePostsAreListed guards a bug where GetPostsByUserID selected the
// author avatar three times, producing thirteen columns for an eleven-column
// scan, so viewing your own profile always failed.
func TestOwnProfilePostsAreListed(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")

	env.createPost(alice, "first", "public")
	env.createPost(alice, "second", "private")

	posts, err := db.GetPostsByUserID(env.Server.DB, int(alice.ID))
	if err != nil {
		t.Fatalf("GetPostsByUserID: %v", err)
	}
	if len(posts) != 2 {
		t.Fatalf("got %d posts, want 2", len(posts))
	}
}

// TestViewerProfilePostsAreFiltered covers the same column-count bug in the
// query used for other people's profiles, and checks the filtering it does.
func TestViewerProfilePostsAreFiltered(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	stranger := env.registerUser("stranger@example.com", "Stranger")

	env.createPost(alice, "public", "public")
	env.createPost(alice, "almost", "almost-private")
	env.createPost(alice, "private", "private")

	posts, err := db.GetPostsByUserVisibleForViewer(
		env.Server.DB, int(alice.ID), int(stranger.ID), false,
	)
	if err != nil {
		t.Fatalf("GetPostsByUserVisibleForViewer: %v", err)
	}

	if len(posts) != 1 {
		t.Fatalf("got %d posts, want 1 (only the public one)", len(posts))
	}
	if posts[0].Content != "public" {
		t.Errorf("got post %q, want the public one", posts[0].Content)
	}
}

// --- referential integrity --------------------------------------------------

// TestForeignKeysAreEnforced guards the connection setting that makes SQLite
// honour FOREIGN KEY clauses. Without it every ON DELETE CASCADE in the
// migrations is inert and orphaned rows accumulate silently.
func TestForeignKeysAreEnforced(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")

	_, err := env.Server.DB.Exec(
		`INSERT INTO post_comments (post_id, user_id, content) VALUES (?, ?, ?)`,
		999999, alice.ID, "comment on a post that does not exist",
	)
	if err == nil {
		t.Fatal("inserted a comment referencing a nonexistent post; foreign keys are not enforced")
	}
}

func TestDeletingUserCascadesToTheirPosts(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	env.createPost(alice, "will be removed with the account", "public")

	if _, err := env.Server.DB.Exec(`DELETE FROM users WHERE id = ?`, alice.ID); err != nil {
		t.Fatalf("delete user: %v", err)
	}

	var count int
	if err := env.Server.DB.QueryRow(
		`SELECT COUNT(*) FROM posts WHERE user_id = ?`, alice.ID,
	).Scan(&count); err != nil {
		t.Fatalf("count posts: %v", err)
	}
	if count != 0 {
		t.Errorf("%d post(s) survived the deletion of their author", count)
	}
}

// --- private messages -------------------------------------------------------

// TestCanMessageRules covers the rule the UI enforces in the browser. Without
// the same check on the server, any account can message any other, over the
// REST endpoint or the WebSocket.
func TestCanMessageRules(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	bob := env.registerUser("bob@example.com", "Bob")
	stranger := env.registerUser("stranger@example.com", "Stranger")
	privateUser := env.registerUser("private@example.com", "Private User")
	env.setPrivate(privateUser, true)

	env.follow(alice, bob, "accepted")
	env.follow(bob, alice, "accepted")
	env.follow(alice, privateUser, "accepted")

	cases := []struct {
		name     string
		from, to *testUser
		want     bool
	}{
		{"mutual follows can message", alice, bob, true},
		{"stranger cannot message", stranger, alice, false},
		{"cannot message yourself", alice, alice, false},
		{"following a public profile allows messaging", stranger, bob, false},
		{"following a private profile does not allow messaging", alice, privateUser, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := db.CanMessage(env.Server.DB, int(tc.from.ID), int(tc.to.ID))
			if err != nil {
				t.Fatalf("CanMessage: %v", err)
			}
			if got != tc.want {
				t.Errorf("CanMessage = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSendMessageToStrangerIsRejected(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	stranger := env.registerUser("stranger@example.com", "Stranger")

	resp := env.do(stranger, http.MethodPost, fmt.Sprintf("/api/messages/%d", alice.ID), map[string]string{
		"content": "unsolicited message",
	})
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusForbidden, "messaging someone with no relationship")

	var count int
	if err := env.Server.DB.QueryRow(`SELECT COUNT(*) FROM messages`).Scan(&count); err != nil {
		t.Fatalf("count messages: %v", err)
	}
	if count != 0 {
		t.Errorf("%d message(s) were stored despite the sender not being allowed to message", count)
	}
}

func TestSendMessageBetweenMutualFollowsSucceeds(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	bob := env.registerUser("bob@example.com", "Bob")
	env.follow(alice, bob, "accepted")
	env.follow(bob, alice, "accepted")

	resp := env.do(alice, http.MethodPost, fmt.Sprintf("/api/messages/%d", bob.ID), map[string]string{
		"content": "hey bob",
	})
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusCreated, "messaging a mutual follow")
}

// --- personal data in responses ---------------------------------------------

func TestSuggestionsDoNotExposeOtherUsersEmail(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	env.registerUser("bob@example.com", "Bob")

	resp := env.do(alice, http.MethodGet, "/api/users/suggestions", nil)
	assertStatus(t, resp, http.StatusOK, "GET /api/users/suggestions")

	var suggestions []map[string]any
	decodeJSON(t, resp, &suggestions)

	if len(suggestions) == 0 {
		t.Fatal("no suggestions returned; the test cannot check the response shape")
	}
	for _, s := range suggestions {
		if _, present := s["email"]; present {
			t.Errorf("suggestion for %v exposes an email address", s["full_name"])
		}
	}
}

func TestOtherUserProfileDoesNotExposeEmailOrDateOfBirth(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	bob := env.registerUser("bob@example.com", "Bob")

	resp := env.do(alice, http.MethodGet, fmt.Sprintf("/api/users/%d", bob.ID), nil)
	assertStatus(t, resp, http.StatusOK, "GET /api/users/{id}")

	var body struct {
		User map[string]any `json:"user"`
	}
	decodeJSON(t, resp, &body)

	for _, field := range []string{"email", "date_of_birth"} {
		if _, present := body.User[field]; present {
			t.Errorf("another user's profile exposes %q", field)
		}
	}

	if body.User["full_name"] != "Bob" {
		t.Errorf("got full_name %v, want Bob", body.User["full_name"])
	}
}

func TestOwnProfileStillIncludesEmail(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")

	resp := env.do(alice, http.MethodGet, fmt.Sprintf("/api/users/%d", alice.ID), nil)
	assertStatus(t, resp, http.StatusOK, "GET own /api/users/{id}")

	var body struct {
		User map[string]any `json:"user"`
	}
	decodeJSON(t, resp, &body)

	if body.User["email"] != alice.Email {
		t.Errorf("own profile does not include the account email: got %v", body.User["email"])
	}
}

func TestPrivateProfileHidesPostsFromStrangers(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	stranger := env.registerUser("stranger@example.com", "Stranger")
	env.setPrivate(alice, true)
	env.createPost(alice, "not for strangers", "public")

	resp := env.do(stranger, http.MethodGet, fmt.Sprintf("/api/users/%d", alice.ID), nil)
	assertStatus(t, resp, http.StatusOK, "GET a private user's profile")

	var body struct {
		Posts  []any          `json:"posts"`
		Viewer map[string]any `json:"viewer"`
	}
	decodeJSON(t, resp, &body)

	if len(body.Posts) != 0 {
		t.Errorf("got %d posts from a private profile, want 0", len(body.Posts))
	}
	if body.Viewer["can_view"] != false {
		t.Error("viewer.can_view should be false for a private profile")
	}
}

// --- uploads ----------------------------------------------------------------

// TestPostUploadRejectsNonImageExtension guards against stored XSS: uploads are
// served from /uploads/ on the app's own origin, so an .html or .svg file would
// run as a page in that origin.
func TestPostUploadRejectsNonImageExtension(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")

	for _, filename := range []string{"payload.html", "payload.svg", "payload.js"} {
		t.Run(filename, func(t *testing.T) {
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			if err := writer.WriteField("content", "post with an attachment"); err != nil {
				t.Fatalf("write field: %v", err)
			}
			if err := writer.WriteField("privacy", "public"); err != nil {
				t.Fatalf("write field: %v", err)
			}

			part, err := writer.CreateFormFile("image", filename)
			if err != nil {
				t.Fatalf("create form file: %v", err)
			}
			if _, err := part.Write([]byte(`<script>alert(1)</script>`)); err != nil {
				t.Fatalf("write file contents: %v", err)
			}
			if err := writer.Close(); err != nil {
				t.Fatalf("close writer: %v", err)
			}

			req, err := http.NewRequest(http.MethodPost, env.HTTP.URL+"/api/posts", &buf)
			if err != nil {
				t.Fatalf("new request: %v", err)
			}
			req.Header.Set("Content-Type", writer.FormDataContentType())

			resp, err := alice.Client.Do(req)
			if err != nil {
				t.Fatalf("POST /api/posts: %v", err)
			}
			defer resp.Body.Close()

			assertStatus(t, resp, http.StatusBadRequest, "uploading "+filename)
		})
	}
}

// --- groups -----------------------------------------------------------------

// TestGroupMemberListRequiresMembership covers a gap where posts, events,
// messages and invites all checked membership but the member list did not, so a
// private group's roster was readable by any account.
func TestGroupMemberListRequiresMembership(t *testing.T) {
	env := newTestEnv(t)
	owner := env.registerUser("owner@example.com", "Group Owner")
	outsider := env.registerUser("outsider@example.com", "Outsider")

	created := env.do(owner, http.MethodPost, "/api/groups", map[string]any{
		"title":       "Study Group",
		"name":        "Study Group",
		"description": "Finals revision",
		"is_private":  true,
	})
	assertStatus(t, created, http.StatusCreated, "POST /api/groups")

	var group map[string]any
	decodeJSON(t, created, &group)

	id, ok := group["id"].(float64)
	if !ok {
		t.Fatalf("group response has no numeric id: %v", group)
	}
	groupID := int(id)

	asOutsider := env.do(outsider, http.MethodGet, fmt.Sprintf("/api/groups/%d/members", groupID), nil)
	defer asOutsider.Body.Close()
	assertStatus(t, asOutsider, http.StatusForbidden, "outsider listing group members")

	asOwner := env.do(owner, http.MethodGet, fmt.Sprintf("/api/groups/%d/members", groupID), nil)
	defer asOwner.Body.Close()
	assertStatus(t, asOwner, http.StatusOK, "owner listing group members")
}
