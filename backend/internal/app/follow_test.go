package app

import (
	"fmt"
	"net/http"
	"testing"

	"social-network/internal/db"
)

// followPath builds the follow/unfollow URL for a user.
func followPath(u *testUser, action string) string {
	return fmt.Sprintf("/api/users/%d/%s", u.ID, action)
}

// requestPath builds the accept/decline URL for a pending request from u.
func requestPath(u *testUser, action string) string {
	return fmt.Sprintf("/api/follow-requests/%d/%s", u.ID, action)
}

// --- following a public profile ---------------------------------------------

// TestFollowPublicUserIsAcceptedImmediately covers the core rule: a public
// profile has no gate, so the relationship is live as soon as it is created.
func TestFollowPublicUserIsAcceptedImmediately(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	bob := env.registerUser("bob@example.com", "Bob")

	resp := env.do(alice, http.MethodPost, followPath(bob, "follow"), nil)
	assertStatus(t, resp, http.StatusOK, "follow a public user")

	var body map[string]string
	decodeJSON(t, resp, &body)

	if body["status"] != "accepted" {
		t.Errorf("got status %q, want %q", body["status"], "accepted")
	}

	following, err := db.IsFollowing(env.Server.DB, int(alice.ID), int(bob.ID))
	if err != nil {
		t.Fatalf("IsFollowing: %v", err)
	}
	if !following {
		t.Error("alice does not follow bob after a successful follow")
	}
}

func TestFollowPrivateUserCreatesPendingRequest(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	bob := env.registerUser("bob@example.com", "Bob")
	env.setPrivate(bob, true)

	resp := env.do(alice, http.MethodPost, followPath(bob, "follow"), nil)
	assertStatus(t, resp, http.StatusOK, "follow a private user")

	var body map[string]string
	decodeJSON(t, resp, &body)

	if body["status"] != "pending" {
		t.Errorf("got status %q, want %q", body["status"], "pending")
	}

	// Pending is not following: it must not grant access yet.
	following, err := db.IsFollowing(env.Server.DB, int(alice.ID), int(bob.ID))
	if err != nil {
		t.Fatalf("IsFollowing: %v", err)
	}
	if following {
		t.Error("a pending request is being treated as an accepted follow")
	}
}

func TestFollowRequiresAuthentication(t *testing.T) {
	env := newTestEnv(t)
	bob := env.registerUser("bob@example.com", "Bob")

	resp := env.do(nil, http.MethodPost, followPath(bob, "follow"), nil)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusUnauthorized, "follow without a session")
}

func TestCannotFollowYourself(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")

	resp := env.do(alice, http.MethodPost, followPath(alice, "follow"), nil)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusBadRequest, "follow yourself")
}

func TestCannotFollowTwice(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	bob := env.registerUser("bob@example.com", "Bob")

	first := env.do(alice, http.MethodPost, followPath(bob, "follow"), nil)
	first.Body.Close()
	assertStatus(t, first, http.StatusOK, "first follow")

	second := env.do(alice, http.MethodPost, followPath(bob, "follow"), nil)
	defer second.Body.Close()
	assertStatus(t, second, http.StatusBadRequest, "second follow of the same user")
}

func TestFollowNonexistentUserReturnsNotFound(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")

	resp := env.do(alice, http.MethodPost, "/api/users/999999/follow", nil)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusNotFound, "follow a user that does not exist")
}

// --- unfollowing ------------------------------------------------------------

func TestUnfollowRemovesRelationship(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	bob := env.registerUser("bob@example.com", "Bob")
	env.follow(alice, bob, "accepted")

	resp := env.do(alice, http.MethodPost, followPath(bob, "unfollow"), nil)
	resp.Body.Close()
	assertStatus(t, resp, http.StatusOK, "unfollow")

	following, err := db.IsFollowing(env.Server.DB, int(alice.ID), int(bob.ID))
	if err != nil {
		t.Fatalf("IsFollowing: %v", err)
	}
	if following {
		t.Error("alice still follows bob after unfollowing")
	}
}

// TestUnfollowIsOneDirectional makes sure unfollowing does not tear down the
// reverse relationship, which is a separate row.
func TestUnfollowIsOneDirectional(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	bob := env.registerUser("bob@example.com", "Bob")
	env.follow(alice, bob, "accepted")
	env.follow(bob, alice, "accepted")

	resp := env.do(alice, http.MethodPost, followPath(bob, "unfollow"), nil)
	resp.Body.Close()
	assertStatus(t, resp, http.StatusOK, "unfollow")

	stillFollowing, err := db.IsFollowing(env.Server.DB, int(bob.ID), int(alice.ID))
	if err != nil {
		t.Fatalf("IsFollowing: %v", err)
	}
	if !stillFollowing {
		t.Error("bob's follow of alice was removed when alice unfollowed bob")
	}
}

// --- follow requests --------------------------------------------------------

func TestAcceptFollowRequestGrantsAccess(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	bob := env.registerUser("bob@example.com", "Bob")
	env.setPrivate(bob, true)
	env.follow(alice, bob, "pending")

	// bob accepts alice's request
	resp := env.do(bob, http.MethodPost, requestPath(alice, "accept"), nil)
	resp.Body.Close()
	assertStatus(t, resp, http.StatusOK, "accept follow request")

	following, err := db.IsFollowing(env.Server.DB, int(alice.ID), int(bob.ID))
	if err != nil {
		t.Fatalf("IsFollowing: %v", err)
	}
	if !following {
		t.Error("alice does not follow bob after bob accepted the request")
	}
}

func TestDeclineFollowRequestRemovesIt(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	bob := env.registerUser("bob@example.com", "Bob")
	env.setPrivate(bob, true)
	env.follow(alice, bob, "pending")

	resp := env.do(bob, http.MethodPost, requestPath(alice, "decline"), nil)
	resp.Body.Close()
	assertStatus(t, resp, http.StatusOK, "decline follow request")

	status, err := db.GetFollowStatus(env.Server.DB, int(alice.ID), int(bob.ID))
	if err != nil {
		t.Fatalf("GetFollowStatus: %v", err)
	}
	if status != nil {
		t.Errorf("declined request is still stored with status %q", status.Status)
	}
}

// TestCannotAcceptSomeoneElsesFollowRequest checks that accept is scoped to the
// logged-in user: mallory must not be able to accept a request addressed to bob
// and so put herself (or anyone) inside a private account.
func TestCannotAcceptSomeoneElsesFollowRequest(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	bob := env.registerUser("bob@example.com", "Bob")
	mallory := env.registerUser("mallory@example.com", "Mallory")
	env.setPrivate(bob, true)
	env.follow(alice, bob, "pending")

	resp := env.do(mallory, http.MethodPost, requestPath(alice, "accept"), nil)
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Fatal("mallory accepted a follow request that was addressed to bob")
	}

	following, err := db.IsFollowing(env.Server.DB, int(alice.ID), int(bob.ID))
	if err != nil {
		t.Fatalf("IsFollowing: %v", err)
	}
	if following {
		t.Error("alice's pending request to bob was accepted by a third party")
	}
}

func TestAcceptingANonPendingRequestFails(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	bob := env.registerUser("bob@example.com", "Bob")
	env.follow(alice, bob, "accepted")

	resp := env.do(bob, http.MethodPost, requestPath(alice, "accept"), nil)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusBadRequest, "accept an already-accepted follow")
}

func TestPendingRequestsAreListedForTheTarget(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	bob := env.registerUser("bob@example.com", "Bob")
	carol := env.registerUser("carol@example.com", "Carol")
	env.setPrivate(bob, true)
	env.follow(alice, bob, "pending")
	env.follow(carol, bob, "accepted")

	resp := env.do(bob, http.MethodGet, "/api/me/follow-requests", nil)
	assertStatus(t, resp, http.StatusOK, "GET /api/me/follow-requests")

	var requests []map[string]any
	decodeJSON(t, resp, &requests)

	if len(requests) != 1 {
		t.Fatalf("got %d pending requests, want 1 (accepted follows must not appear)", len(requests))
	}
	if got := int64(requests[0]["follower_id"].(float64)); got != alice.ID {
		t.Errorf("pending request is from user %d, want %d", got, alice.ID)
	}
}

// --- counts and mutual relationships ----------------------------------------

// TestFollowCountsOnlyCountAcceptedFollows guards against pending requests
// inflating the numbers shown on a profile.
func TestFollowCountsOnlyCountAcceptedFollows(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	bob := env.registerUser("bob@example.com", "Bob")
	carol := env.registerUser("carol@example.com", "Carol")
	dave := env.registerUser("dave@example.com", "Dave")

	env.follow(bob, alice, "accepted")
	env.follow(carol, alice, "pending")
	env.follow(alice, dave, "accepted")

	resp := env.do(alice, http.MethodGet, "/api/me/follow-counts", nil)
	assertStatus(t, resp, http.StatusOK, "GET /api/me/follow-counts")

	var counts map[string]int
	decodeJSON(t, resp, &counts)

	if counts["followers"] != 1 {
		t.Errorf("got %d followers, want 1 (the pending request must not count)", counts["followers"])
	}
	if counts["following"] != 1 {
		t.Errorf("got %d following, want 1", counts["following"])
	}
}

func TestAreMutualFriendsRequiresBothDirections(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	bob := env.registerUser("bob@example.com", "Bob")

	env.follow(alice, bob, "accepted")

	mutual, err := db.AreMutualFriends(env.Server.DB, int(alice.ID), int(bob.ID))
	if err != nil {
		t.Fatalf("AreMutualFriends: %v", err)
	}
	if mutual {
		t.Fatal("a one-way follow was reported as mutual")
	}

	env.follow(bob, alice, "accepted")

	mutual, err = db.AreMutualFriends(env.Server.DB, int(alice.ID), int(bob.ID))
	if err != nil {
		t.Fatalf("AreMutualFriends: %v", err)
	}
	if !mutual {
		t.Error("two accepted follows were not reported as mutual")
	}
}

func TestPendingFollowIsNotMutual(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	bob := env.registerUser("bob@example.com", "Bob")

	env.follow(alice, bob, "accepted")
	env.follow(bob, alice, "pending")

	mutual, err := db.AreMutualFriends(env.Server.DB, int(alice.ID), int(bob.ID))
	if err != nil {
		t.Fatalf("AreMutualFriends: %v", err)
	}
	if mutual {
		t.Error("a pending reverse follow was counted towards mutual friendship")
	}
}

func TestFollowersListExcludesPendingRequests(t *testing.T) {
	env := newTestEnv(t)
	alice := env.registerUser("alice@example.com", "Alice")
	bob := env.registerUser("bob@example.com", "Bob")
	carol := env.registerUser("carol@example.com", "Carol")

	env.follow(bob, alice, "accepted")
	env.follow(carol, alice, "pending")

	resp := env.do(alice, http.MethodGet, "/api/me/followers", nil)
	assertStatus(t, resp, http.StatusOK, "GET /api/me/followers")

	var followers []map[string]any
	decodeJSON(t, resp, &followers)

	if len(followers) != 1 {
		t.Fatalf("got %d followers, want 1", len(followers))
	}
	if followers[0]["follower_name"] != "Bob" {
		t.Errorf("got follower %v, want Bob", followers[0]["follower_name"])
	}
}
