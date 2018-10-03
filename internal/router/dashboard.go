package router

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	humanize "github.com/dustin/go-humanize"
	"github.com/lnsp/microlog/internal/utils"
)

type dashboardPost struct {
	Title, Author, ID, Date string
	Likes                   int
}

type dashboardUser struct {
	Name, MemberSince string
}

type dashboardOption struct {
	Name   string
	Active bool
}

type dashboardContext struct {
	Context
	PopularPosts   []dashboardPost
	LatestUsers    []dashboardUser
	PopularOptions []dashboardOption
}

const (
	timeWeek  = time.Hour * 24 * 7
	timeMonth = timeWeek * 4
	timeYear  = timeMonth * 12
)

func (router *Router) dashboardContext(r *http.Request) *dashboardContext {
	ctx := &dashboardContext{
		Context: *router.defaultContext(r),
	}
	var timeInterval time.Duration
	switch r.URL.Query().Get("popular") {
	case "month":
		timeInterval = -timeMonth
		ctx.PopularOptions = []dashboardOption{
			{"week", false}, {"month", true}, {"year", false},
		}
	case "year":
		timeInterval = -timeYear
		ctx.PopularOptions = []dashboardOption{
			{"week", false}, {"month", false}, {"year", true},
		}
	default:
		timeInterval = -timeWeek
		ctx.PopularOptions = []dashboardOption{
			{"week", true}, {"month", false}, {"year", false},
		}
	}
	popularPosts, err := router.Data.PopularPosts(time.Now().Add(timeInterval), dashboardPostsLimit)
	if err != nil {
		log.WithFields(logrus.Fields{
			"id":   ctx.UserID,
			"addr": utils.RemoteHost(r),
		}).WithError(err).Error("failed to fetch popular posts")
		ctx.ErrorMessage = "An internal error occured, please try again."
	}
	recentUsers, err := router.Data.RecentUsers(dashboardUsersLimit)
	if err != nil {
		log.WithFields(logrus.Fields{
			"id":   ctx.UserID,
			"addr": utils.RemoteHost(r),
		}).WithError(err).Error("failed to fetch new users")
		ctx.ErrorMessage = "An internal error occured, please try again."
	}
	j := 0
	ctx.PopularPosts = make([]dashboardPost, len(popularPosts))
	for _, post := range popularPosts {
		user, err := router.Data.User(post.UserID)
		if err != nil {
			log.WithFields(logrus.Fields{
				"id":   ctx.UserID,
				"post": post.ID,
				"addr": utils.RemoteHost(r),
			}).WithError(err).Error("failed to fetch user")
			continue
		}
		likes, err := router.Data.NumberOfLikes(post.ID)
		if err != nil {
			log.WithFields(logrus.Fields{
				"id":   ctx.UserID,
				"post": post.ID,
				"addr": utils.RemoteHost(r),
			}).WithError(err).Error("failed to retrieve number of likes")
			continue
		}
		ctx.PopularPosts[j] = dashboardPost{
			Title:  post.Title,
			Author: user.Name,
			ID:     strconv.FormatUint(uint64(post.ID), 10),
			Date:   humanize.Time(post.CreatedAt),
			Likes:  likes,
		}
		j++
	}
	ctx.LatestUsers = make([]dashboardUser, len(recentUsers))
	for i, user := range recentUsers {
		ctx.LatestUsers[i] = dashboardUser{
			Name:        user.Name,
			MemberSince: humanize.Time(user.CreatedAt),
		}
	}
	return ctx
}

func (router *Router) dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := router.dashboardContext(r)
	router.render(dashboardTemplate, w, ctx)
}
