package router

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

type PostContext struct {
	Context
	ID          uint
	Author      string
	Title       string
	Content     string
	HTMLContent template.HTML
	Date        string
	Self        bool
}

func (router *Router) postRedirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	http.Redirect(w, r, fmt.Sprintf("/%s/%s/", vars["user"], vars["post"]), http.StatusPermanentRedirect)
}

func (router *Router) postDelete(w http.ResponseWriter, r *http.Request) {
	ctx := router.postContext(r)
	if err := router.Data.DeletePost(ctx.UserID, ctx.ID); err != nil {
		log.Errorln("Failed to delete post:", err)
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
	postCtx := PostContext{Context: *ctx}
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
	user, err := router.Data.GetUser(ctx.UserID)
	if err != nil {
		log.Errorln("Failed to find user:", err)
		router.renderNotFound(w, r, "user")
		return
	}
	if postID != "" {
		id, _ := strconv.ParseUint(postID, 10, 64)
		post, err := router.Data.GetPost(uint(id))
		if err != nil {
			log.Debugln("Failed to find post:", err)
			router.renderNotFound(w, r, "post")
			return
		}
		if err := router.Data.UpdatePost(user.ID, post.ID, title, content); err != nil {
			log.Errorln("Failed to update post:", err)
			postCtx := router.postContextWithID(r, user.Name, post.ID)
			postCtx.ErrorMessage = "Unexpected internal error, please try again."
			router.render(postEditTemplate, w, postCtx)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/%s/%d/", user.Name, id), http.StatusSeeOther)
	} else {
		postCtx := PostContext{
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
			log.Errorln("Failed to add post:", err)
		}
		http.Redirect(w, r, fmt.Sprintf("/%s/%d/", user.Name, id), http.StatusSeeOther)
	}
}

func (router *Router) postContextWithID(r *http.Request, username string, id uint) PostContext {
	ctx := router.defaultContext(r)
	post, err := router.Data.GetPost(id)
	if err != nil {
		ctx.ErrorMessage = "Post does not exist."
	}
	user, err := router.Data.GetUserByName(username)
	if err != nil {
		ctx.ErrorMessage = "User does not exist."
	}
	if ctx.ErrorMessage != "" {
		return PostContext{
			Context: *ctx,
		}
	}
	rendered := blackfriday.MarkdownCommon([]byte(post.Content))
	safe := bluemonday.UGCPolicy().SanitizeBytes(rendered)
	return PostContext{
		Context:     *ctx,
		Self:        ctx.SignedIn && ctx.UserID == post.UserID,
		Author:      user.Name,
		ID:          post.ID,
		Title:       post.Title,
		Content:     post.Content,
		HTMLContent: template.HTML(safe),
		Date:        post.CreatedAt.Format(timeFormat),
	}
}

func (router *Router) postContext(r *http.Request) PostContext {
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
	log.Debugf("User %d reported post %d", ctx.UserID, ctx.ID)
	ctx.ErrorMessage = "Thank you for the report. Our team will look into the issue!"
	router.render(postTemplate, w, ctx)
}
