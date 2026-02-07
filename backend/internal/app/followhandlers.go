package app

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"social-network/internal/db"
	"social-network/internal/models"
)

func extractUserIDFromPath(path string, prefix string, suffix string) (int, error) {
	trimmed := strings.TrimPrefix(path, prefix)
	trimmed = strings.TrimSuffix(trimmed, suffix)
	return strconv.Atoi(trimmed)
}

func (s *Server) handleUserFollowRoutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Stage 5: GET /api/users/{id}
	// (must be checked BEFORE follow/unfollow, otherwise it will fall into not found)
	trimmed := strings.TrimPrefix(path, "/api/users/")
	if trimmed != "" && !strings.Contains(trimmed, "/") && r.Method == http.MethodGet {
		s.handleGetUserProfile(w, r)
		return
	}

	if strings.HasSuffix(path, "/follow") {
		s.handleFollow(w, r)
		return
	}

	if strings.HasSuffix(path, "/unfollow") {
		s.handleUnfollow(w, r)
		return
	}

	writeError(w, http.StatusNotFound, "not found")
}

func (s *Server) handleFollowRequestRoutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasSuffix(path, "/accept") {
		s.handleAcceptFollow(w, r)
		return
	}

	if strings.HasSuffix(path, "/decline") {
		s.handleDeclineFollow(w, r)
		return
	}

	writeError(w, http.StatusNotFound, "not found")
}

func (s *Server) handleFollow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	currentUser := CurrentUser(r.Context())
	if currentUser == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetID, err := extractUserIDFromPath(r.URL.Path, "/api/users/", "/follow")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	if int(currentUser.ID) == targetID {
		writeError(w, http.StatusBadRequest, "cannot follow yourself")
		return
	}

	existingFollow, err := db.GetFollowStatus(s.DB, int(currentUser.ID), targetID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check follow status")
		return
	}
	if existingFollow != nil {
		writeError(w, http.StatusBadRequest, "already following or request pending")
		return
	}

	targetUser, err := db.GetUserByID(s.DB, int64(targetID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get user")
		return
	}
	if targetUser == nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	status := "accepted"
	if targetUser.IsPrivate {
		status = "pending"
	}

	err = db.CreateFollow(s.DB, int(currentUser.ID), targetID, status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to follow user")
		return
	}

	// Create notification for the target user
	fromUserID := int(currentUser.ID)
	if status == "pending" {
		// Notify target user about follow request
		_, err = db.CreateNotification(s.DB, targetID, models.NotificationTypeFollowRequest, nil, &fromUserID, "sent you a follow request")
		if err != nil {
			log.Println("Failed to create follow request notification:", err)
		}
	} else {
		// Notify target user that someone followed them (public profile)
		_, err = db.CreateNotification(s.DB, targetID, models.NotificationTypeFollowAccept, nil, &fromUserID, "started following you")
		if err != nil {
			log.Println("Failed to create follow notification:", err)
		}
	}

	message := "now following user"
	if status == "pending" {
		message = "follow request sent"
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": message,
		"status":  status,
	})
}

func (s *Server) handleUnfollow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	currentUser := CurrentUser(r.Context())
	if currentUser == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetID, err := extractUserIDFromPath(r.URL.Path, "/api/users/", "/unfollow")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	err = db.DeleteFollow(s.DB, int(currentUser.ID), targetID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to unfollow user")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "unfollowed user"})
}

func (s *Server) handleGetFollowers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	currentUser := CurrentUser(r.Context())
	if currentUser == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	followers, err := db.GetFollowers(s.DB, int(currentUser.ID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get followers")
		return
	}

	if followers == nil {
		followers = []models.Follow{}
	}

	writeJSON(w, http.StatusOK, followers)
}

func (s *Server) handleGetFollowing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	currentUser := CurrentUser(r.Context())
	if currentUser == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	following, err := db.GetFollowing(s.DB, int(currentUser.ID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get following")
		return
	}

	if following == nil {
		following = []models.Follow{}
	}

	writeJSON(w, http.StatusOK, following)
}

func (s *Server) handleGetFollowRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	currentUser := CurrentUser(r.Context())
	if currentUser == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	requests, err := db.GetPendingFollowRequests(s.DB, int(currentUser.ID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get follow requests")
		return
	}

	if requests == nil {
		requests = []models.Follow{}
	}

	writeJSON(w, http.StatusOK, requests)
}

func (s *Server) handleAcceptFollow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	currentUser := CurrentUser(r.Context())
	if currentUser == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	followerID, err := extractUserIDFromPath(r.URL.Path, "/api/follow-requests/", "/accept")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	follow, err := db.GetFollowStatus(s.DB, followerID, int(currentUser.ID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check follow status")
		return
	}
	if follow == nil {
		writeError(w, http.StatusNotFound, "follow request not found")
		return
	}
	if follow.Status != "pending" {
		writeError(w, http.StatusBadRequest, "follow request is not pending")
		return
	}

	err = db.UpdateFollowStatus(s.DB, followerID, int(currentUser.ID), "accepted")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to accept follow request")
		return
	}

	// Create notification for the requester that their request was accepted
	fromUserID := int(currentUser.ID)
	_, err = db.CreateNotification(s.DB, followerID, models.NotificationTypeFollowAccept, nil, &fromUserID, "accepted your follow request")
	if err != nil {
		log.Println("Failed to create follow accept notification:", err)
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "follow request accepted"})
}

func (s *Server) handleDeclineFollow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	currentUser := CurrentUser(r.Context())
	if currentUser == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	followerID, err := extractUserIDFromPath(r.URL.Path, "/api/follow-requests/", "/decline")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	follow, err := db.GetFollowStatus(s.DB, followerID, int(currentUser.ID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check follow status")
		return
	}
	if follow == nil {
		writeError(w, http.StatusNotFound, "follow request not found")
		return
	}

	err = db.DeleteFollow(s.DB, followerID, int(currentUser.ID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to decline follow request")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "follow request declined"})
}

func (s *Server) handleGetFollowCounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	currentUser := CurrentUser(r.Context())
	if currentUser == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	followerCount, err := db.GetFollowerCount(s.DB, int(currentUser.ID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get follower count")
		return
	}

	followingCount, err := db.GetFollowingCount(s.DB, int(currentUser.ID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get following count")
		return
	}

	writeJSON(w, http.StatusOK, map[string]int{
		"followers": followerCount,
		"following": followingCount,
	})
}

// Friends endpoints for messaging
func (s *Server) handleGetFriends(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	friends, err := db.GetMutualFriends(s.DB, int(user.ID))
	if err != nil {
		log.Println("GetMutualFriends error:", err)
		writeError(w, http.StatusInternalServerError, "failed to get friends")
		return
	}

	if friends == nil {
		friends = []models.User{}
	}

	var result []map[string]any
	for _, f := range friends {
		result = append(result, map[string]any{
			"id":         f.ID,
			"uuid":       f.UUID,
			"full_name":  f.FullName,
			"nickname":   f.Nickname,
			"avatar_url": f.AvatarURL,
		})
	}

	if result == nil {
		result = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleCheckMutual(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetID := r.URL.Query().Get("user_id")
	if targetID == "" {
		writeError(w, http.StatusBadRequest, "user_id required")
		return
	}

	id, err := strconv.Atoi(targetID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	isMutual, err := db.AreMutualFriends(s.DB, int(user.ID), id)
	if err != nil {
		log.Println("AreMutualFriends error:", err)
		writeError(w, http.StatusInternalServerError, "failed to check mutual status")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"is_mutual": isMutual})
}