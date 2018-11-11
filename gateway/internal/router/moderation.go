package router

import (
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (router *Router) Error(w http.ResponseWriter, r *http.Request, msg string, status int) {
	http.Error(w, msg, status)
}

type moderationReport struct {
	ID        uint
	PostTitle string
	PostID    uint
	PostUser  string
	Reporter  string
	Reason    string
	Status    bool
}

type moderationContext struct {
	Context
	Reports []moderationReport
}

func (router *Router) ModerateDelete(w http.ResponseWriter, r *http.Request) {
	ctx := router.defaultContext(r)
	if !ctx.Moderator {
		router.Error(w, r, "Unauthorized", http.StatusUnauthorized)
		return
	}
	reportID, err := strconv.Atoi(mux.Vars(r)["report"])
	if err != nil {
		router.Error(w, r, "report not a number", http.StatusBadRequest)
		return
	}
	report, err := router.Data.Report(uint(reportID))
	if err != nil {
		router.Error(w, r, "report not found", http.StatusNotFound)
		return
	}
	post, err := router.Data.Post(report.PostID)
	if err != nil {
		router.Error(w, r, "post not found", http.StatusNotFound)
		return
	}
	if err := router.Data.DeletePost(post.UserID, post.ID); err != nil {
		router.Error(w, r, "could not delete post", http.StatusInternalServerError)
		return
	}
	if err := router.Data.CloseReport(uint(reportID)); err != nil {
		router.Error(w, r, "could not close report", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/moderate", http.StatusSeeOther)
}

func (router *Router) ModerateClose(w http.ResponseWriter, r *http.Request) {
	ctx := router.defaultContext(r)
	if !ctx.Moderator {
		router.Error(w, r, "Unauthorized", http.StatusUnauthorized)
		return
	}
	reportID, err := strconv.Atoi(mux.Vars(r)["report"])
	if err != nil {
		router.Error(w, r, "report not a number", http.StatusBadRequest)
		return
	}
	if err := router.Data.CloseReport(uint(reportID)); err != nil {
		router.Error(w, r, "could not close report", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/moderate", http.StatusSeeOther)
}

func (router *Router) Moderate(w http.ResponseWriter, r *http.Request) {
	ctx := router.defaultContext(r)
	if !ctx.Moderator {
		router.Error(w, r, "Unauthorized", http.StatusUnauthorized)
		return
	}
	reports, err := router.Data.Reports()
	if err != nil {
		router.Error(w, r, "Internal error", http.StatusInternalServerError)
		return
	}
	modContext := moderationContext{
		Context: *ctx,
		Reports: make([]moderationReport, len(reports)),
	}
	for index, report := range reports {
		reporter, err := router.Data.User(report.ReporterID)
		if err != nil {
			continue
		}
		post, err := router.Data.Post(report.PostID)
		if err != nil {
			continue
		}
		user, err := router.Data.User(post.UserID)
		if err != nil {
			continue
		}
		modContext.Reports[index] = moderationReport{
			ID:        report.ID,
			Reason:    report.Reason,
			Reporter:  reporter.Name,
			PostUser:  user.Name,
			PostID:    post.ID,
			PostTitle: post.Title,
			Status:    report.Open,
		}
	}
	router.render(moderationTemplate, w, modContext)
}
