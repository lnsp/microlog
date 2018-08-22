package router

import (
	"net/http"

	"github.com/gorilla/mux"
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
		log.Errorln("Failed to find user:", err)
		http.Redirect(w, r, "/auth/logout", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/"+user.Name, http.StatusSeeOther)
}

func (router *Router) profile(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["user"]
	user, err := router.Data.GetUserByName(name)
	if err != nil {
		log.Debugln("Failed to find user", name)
		router.renderNotFound(w, r, "profile")
		return
	}
	posts, err := router.Data.GetPostsByUser(user.ID)
	if err != nil {
		log.Errorln("Failed to get posts for user", user.ID)
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
		log.Debugln("Failed to find user:", err)
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
		log.Debugln("Failed to find user:", err)
		router.renderNotFound(w, r, "profile")
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
		log.Errorln("Failed to set biography:", err)
	}
	http.Redirect(w, r, "/"+user.Name, http.StatusSeeOther)
}

func (router *Router) profileDelete(w http.ResponseWriter, r *http.Request) {
	ctx := router.defaultContext(r)
	if !ctx.SignedIn {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}
	router.render(profileDeleteTemplate, w, ctx)
}

func (router *Router) profileDeleteSubmit(w http.ResponseWriter, r *http.Request) {
	ctx := router.defaultContext(r)
	if !ctx.SignedIn {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}
	router.Data.DeleteUser(ctx.UserID)
	http.Redirect(w, r, "/auth/logout", http.StatusSeeOther)
}
