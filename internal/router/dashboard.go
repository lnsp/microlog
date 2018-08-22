package router

import (
	"net/http"
	"strconv"

	humanize "github.com/dustin/go-humanize"
)

type dashboardPost struct {
	Title, Author, ID, Date string
}

type dashboardUser struct {
	Name, MemberSince string
}

type dashboardContext struct {
	Context
	LatestPosts []dashboardPost
	LatestUsers []dashboardUser
}

func (router *Router) dashboardContext(r *http.Request) *dashboardContext {
	ctx := dashboardContext{
		Context: *router.defaultContext(r),
	}
	posts, err := router.Data.GetRecentPosts(dashboardPostsLimit)
	if err != nil {
		log.Errorln("Failed to fetch recent posts:", err)
		ctx.ErrorMessage = "An internal error occured, please try again."
	}
	users, err := router.Data.GetRecentUsers(dashboardUsersLimit)
	if err != nil {
		log.Errorln("Failed to fetch recent signups:", err)
		ctx.ErrorMessage = "An internal error occured, please try again."
	}
	j := 0
	ctx.LatestPosts = make([]dashboardPost, len(posts))
	for _, post := range posts {
		user, err := router.Data.GetUser(post.UserID)
		if err != nil {
			log.Errorln("Failed to fetch user:", err)
			continue
		}
		ctx.LatestPosts[j] = dashboardPost{
			Title:  post.Title,
			Author: user.Name,
			ID:     strconv.FormatUint(uint64(post.ID), 10),
			Date:   humanize.Time(post.CreatedAt),
		}
		j++
	}
	ctx.LatestUsers = make([]dashboardUser, len(users))
	for i, user := range users {
		ctx.LatestUsers[i] = dashboardUser{
			Name:        user.Name,
			MemberSince: humanize.Time(user.CreatedAt),
		}
	}
	return &ctx
}

func (router *Router) dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := router.dashboardContext(r)
	router.render(dashboardTemplate, w, ctx)
}
