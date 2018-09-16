package router

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/lnsp/microlog/internal/utils"
)

type profilePost struct {
	Title  string
	Date   string
	Author string
	ID     uint
}

type profileContext struct {
	Context
	Name        string
	Biography   string
	MemberSince string
	PostCount   int
	Self        bool
	Posts       []profilePost
}

func (router *Router) profileRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := router.defaultContext(r)
	if !ctx.SignedIn {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}
	user, err := router.Data.GetUser(ctx.UserID)
	if err != nil {
		log.WithFields(logrus.Fields{
			"id":   ctx.UserID,
			"addr": utils.RemoteHost(r),
		}).WithError(err).Error("failed to find user")
		http.Redirect(w, r, "/auth/logout", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/"+user.Name, http.StatusSeeOther)
}

func (router *Router) profile(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["user"]
	user, err := router.Data.GetUserByName(name)
	if err != nil {
		log.WithFields(logrus.Fields{
			"name": name,
			"addr": utils.RemoteHost(r),
		}).WithError(err).Debug("failed to find user")
		router.renderNotFound(w, r, "profile")
		return
	}
	posts, err := router.Data.GetPostsByUser(user.ID)
	if err != nil {
		log.WithFields(logrus.Fields{
			"id":   user.ID,
			"name": user.Name,
			"addr": utils.RemoteHost(r),
		}).WithError(err).Error("failed to get posts")
		router.renderNotFound(w, r, "profile")
		return
	}
	// TODO: Sort by specific topic order (date etc.)
	ctx := router.defaultContext(r)
	profileCtx := profileContext{
		Context:     *ctx,
		Name:        user.Name,
		Biography:   user.Biography,
		MemberSince: user.CreatedAt.Format(timeFormat),
		PostCount:   len(posts),
		Self:        ctx.SignedIn && ctx.UserID == user.ID,
		Posts:       make([]profilePost, len(posts)),
	}
	for i := range posts {
		profileCtx.Posts[i] = profilePost{
			Title:  posts[i].Title,
			Date:   posts[i].CreatedAt.Format(timeFormat),
			Author: user.Name,
			ID:     posts[i].ID,
		}
	}
	router.render(profileTemplate, w, profileCtx)
}

func (router *Router) profileEdit(w http.ResponseWriter, r *http.Request) {
	ctx := router.defaultContext(r)
	if !ctx.SignedIn {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}
	user, err := router.Data.GetUser(ctx.UserID)
	if err != nil {
		log.WithFields(logrus.Fields{
			"id":   ctx.UserID,
			"addr": utils.RemoteHost(r),
		}).WithError(err).Error("failed to find user")
		router.renderNotFound(w, r, "profile")
		return
	}
	profileCtx := profileContext{
		Context:   *ctx,
		Name:      user.Name,
		Biography: user.Biography,
	}
	router.render(profileEditTemplate, w, profileCtx)
}

func (router *Router) profileEditSubmit(w http.ResponseWriter, r *http.Request) {
	ctx := router.defaultContext(r)
	if !ctx.SignedIn {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}
	user, err := router.Data.GetUser(ctx.UserID)
	if err != nil {
		log.WithFields(logrus.Fields{
			"id":   ctx.UserID,
			"addr": utils.RemoteHost(r),
		}).WithError(err).Error("failed to find user")
		router.renderNotFound(w, r, "profile")
		return
	}
	biography := r.FormValue("biography")
	profileCtx := profileContext{
		Context:   *ctx,
		Name:      user.Name,
		Biography: biography,
	}
	if !router.Data.ValidateBiography(biography) {
		profileCtx.ErrorMessage = "Your biography must have at max 240 characters."
		router.render(profileEditTemplate, w, profileCtx)
		return
	}
	if err := router.Data.SetBiography(ctx.UserID, biography); err != nil {
		log.WithFields(logrus.Fields{
			"id":   user.ID,
			"name": user.Name,
			"addr": utils.RemoteHost(r),
		}).WithError(err).Error("failed to update biography")
		profileCtx.ErrorMessage = "Unexpected internal error, please try again."
		router.render(profileEditTemplate, w, profileCtx)
		return
	}
	log.WithFields(logrus.Fields{
		"id":   user.ID,
		"name": user.Name,
		"addr": utils.RemoteHost(r),
	}).Debug("updated user profile")
	http.Redirect(w, r, "/"+user.Name, http.StatusSeeOther)
}
