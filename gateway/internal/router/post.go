package router

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/lnsp/microlog/gateway/pkg/utils"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

type postContext struct {
	Context
	ID          uint
	Author      string
	Title       string
	Content     string
	HTMLContent template.HTML
	Date        string
	Self        bool
	Liked       bool
}

func (router *Router) postRedirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	http.Redirect(w, r, fmt.Sprintf("/%s/%s/", vars["user"], vars["post"]), http.StatusPermanentRedirect)
}

func (router *Router) postDelete(w http.ResponseWriter, r *http.Request) {
	ctx := router.postContext(r)
	if err := router.Data.DeletePost(ctx.UserID, ctx.ID); err != nil {
		log.WithFields(logrus.Fields{
			"id":   ctx.UserID,
			"post": ctx.ID,
			"addr": utils.RemoteHost(r),
		}).WithError(err).Error("failed to delete post")
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (router *Router) postNew(w http.ResponseWriter, r *http.Request) {
	ctx := router.defaultContext(r)
	if !ctx.SignedIn {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}
	postCtx := postContext{Context: *ctx}
	router.render(postEditTemplate, w, postCtx)
}

func (router *Router) postEdit(w http.ResponseWriter, r *http.Request) {
	ctx := router.defaultContext(r)
	if !ctx.SignedIn {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}
	postCtx := router.postContext(r)
	router.render(postEditTemplate, w, postCtx)
}

func (router *Router) postSubmit(w http.ResponseWriter, r *http.Request) {
	var (
		postID  = r.FormValue("id")
		title   = r.FormValue("title")
		content = r.FormValue("content")
	)
	ctx := router.defaultContext(r)
	if !ctx.SignedIn {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}
	user, err := router.Data.User(ctx.UserID)
	if err != nil {
		log.WithFields(logrus.Fields{
			"id":   ctx.UserID,
			"post": postID,
			"addr": utils.RemoteHost(r),
		}).WithError(err).Error("failed to find user")
		router.renderNotFound(w, r, "user")
		return
	}
	if postID != "" {
		id, _ := strconv.ParseUint(postID, 10, 64)
		post, err := router.Data.Post(uint(id))
		if err != nil {
			log.WithFields(logrus.Fields{
				"id":   ctx.UserID,
				"post": id,
				"addr": utils.RemoteHost(r),
			}).WithError(err).Error("failed to find post")
			router.renderNotFound(w, r, "post")
			return
		}
		postCtx := postContext{
			Context: *ctx,
			Author:  user.Name,
			Title:   title,
			Content: content,
		}
		if !router.Data.ValidatePostTitle(title) {
			postCtx.ErrorMessage = "Your title must have at max 80 characters."
		} else if !router.Data.ValidatePostContent(content) {
			postCtx.ErrorMessage = "Your content must have at max 80000 characters."
		}
		if postCtx.ErrorMessage != "" {
			router.render(postEditTemplate, w, postCtx)
			return
		}
		if err := router.Data.UpdatePost(user.ID, post.ID, title, content); err != nil {
			postCtx.ErrorMessage = "Unexpected internal error, please try again."
			log.WithFields(logrus.Fields{
				"id":   user.ID,
				"post": post.ID,
				"addr": utils.RemoteHost(r),
			}).WithError(err).Error("failed to update post")
			router.render(postEditTemplate, w, postCtx)
			return
		}
		log.WithFields(logrus.Fields{
			"id":   user.ID,
			"post": post.ID,
			"addr": utils.RemoteHost(r),
		}).Debug("updated post")
		http.Redirect(w, r, fmt.Sprintf("/%s/%d/", user.Name, id), http.StatusSeeOther)
	} else {
		postCtx := postContext{
			Context: *ctx,
			Author:  user.Name,
			Title:   title,
			Content: content,
		}
		if !router.Data.ValidatePostTitle(title) {
			postCtx.ErrorMessage = "Your title must have at max 80 characters."
		} else if !router.Data.ValidatePostContent(content) {
			postCtx.ErrorMessage = "Your content must have at max 80000 characters."
		}
		if postCtx.ErrorMessage != "" {
			router.render(postEditTemplate, w, postCtx)
			return
		}
		id, err := router.Data.AddPost(user.ID, title, content)
		if err != nil {
			log.WithFields(logrus.Fields{
				"id":   user.ID,
				"addr": utils.RemoteHost(r),
			}).WithError(err).Error("failed to add post")
			postCtx.ErrorMessage = "Unexpected internal error, please try again."
			router.render(postEditTemplate, w, postCtx)
			return
		}
		log.WithFields(logrus.Fields{
			"id":    user.ID,
			"post":  id,
			"addr":  utils.RemoteHost(r),
			"title": title,
		}).Debug("added new post")
		http.Redirect(w, r, fmt.Sprintf("/%s/%d/", user.Name, id), http.StatusSeeOther)
	}
}

func (router *Router) postContextWithID(r *http.Request, username string, id uint) postContext {
	ctx := router.defaultContext(r)
	post, err := router.Data.Post(id)
	if err != nil {
		ctx.ErrorMessage = "Post does not exist."
	}
	user, err := router.Data.UserByName(username)
	if err != nil {
		ctx.ErrorMessage = "User does not exist."
	}
	if ctx.ErrorMessage != "" {
		return postContext{
			Context: *ctx,
		}
	}
	rendered := blackfriday.MarkdownCommon([]byte(post.Content))
	safe := bluemonday.UGCPolicy().SanitizeBytes(rendered)
	return postContext{
		Context:     *ctx,
		Self:        ctx.SignedIn && ctx.UserID == post.UserID,
		Author:      user.Name,
		ID:          post.ID,
		Title:       post.Title,
		Content:     post.Content,
		HTMLContent: template.HTML(safe),
		Date:        post.CreatedAt.Format(timeFormat),
		Liked:       ctx.SignedIn && router.Data.HasLiked(ctx.UserID, post.ID),
	}
}

func (router *Router) postContext(r *http.Request) postContext {
	userName := mux.Vars(r)["user"]
	postID := mux.Vars(r)["post"]
	id, _ := strconv.ParseUint(postID, 10, 64)
	return router.postContextWithID(r, userName, uint(id))
}

func (router *Router) post(w http.ResponseWriter, r *http.Request) {
	router.render(postTemplate, w, router.postContext(r))
}

func (router *Router) report(w http.ResponseWriter, r *http.Request) {
	ctx := router.postContext(r)
	if !ctx.SignedIn {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}
	router.render(reportTemplate, w, ctx)
}

func (router *Router) reportSubmit(w http.ResponseWriter, r *http.Request) {
	ctx := router.postContext(r)
	if !ctx.SignedIn {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}
	reason := r.FormValue("reason")
	if !router.Data.ValidateReportReason(reason) {
		ctx.ErrorMessage = "The reasoning must be at max 240 characters."
		router.render(reportTemplate, w, ctx)
		return
	}
	if err := router.Data.AddReport(ctx.ID, ctx.UserID, reason); err != nil {
		ctx.ErrorMessage = "Unexpected internal error, please try again."
		router.render(reportTemplate, w, ctx)
		return
	}
	log.WithFields(logrus.Fields{
		"id":   ctx.UserID,
		"post": ctx.ID,
		"addr": utils.RemoteHost(r),
	}).Debug("reported post")
	ctx.ErrorMessage = "Thank you for the report. Our team will look into the issue!"
	router.render(postTemplate, w, ctx)
}

func (router *Router) like(w http.ResponseWriter, r *http.Request) {
	ctx := router.defaultContext(r)
	if !ctx.SignedIn {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}
	vars := mux.Vars(r)
	author := vars["user"]
	post := vars["post"]
	postID, _ := strconv.ParseUint(post, 10, 64)
	if err := router.Data.ToggleLike(ctx.UserID, uint(postID)); err != nil {
		log.WithFields(logrus.Fields{
			"id":   ctx.UserID,
			"post": postID,
			"addr": utils.RemoteHost(r),
		}).WithError(err).Error("failed to toggle like")
		ctx.ErrorMessage = "Unexpected internal error, please try again."
	}
	log.WithFields(logrus.Fields{
		"id":   ctx.UserID,
		"post": postID,
		"addr": utils.RemoteHost(r),
	}).Debug("toggled like")
	http.Redirect(w, r, fmt.Sprintf("/%s/%s/", author, post), http.StatusSeeOther)
}
