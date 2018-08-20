package router

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/Sirupsen/logrus"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/lnsp/microlog/internal/models"
)

const (
	timeFormat          = "Monday, 2. January at 15:04"
	sessionCookieName   = "session_token"
	dashboardPostsLimit = 10
	dashboardUsersLimit = 10
)

var log = &logrus.Logger{
	Out:       os.Stderr,
	Formatter: new(logrus.TextFormatter),
	Hooks:     make(logrus.LevelHooks),
	Level:     logrus.DebugLevel,
}

var (
	signupSuccessTemplate  = template.Must(template.ParseFiles("./web/templates/base.html", "./web/templates/signupSuccess.html"))
	dashboardTemplate      = template.Must(template.ParseFiles("./web/templates/base.html", "./web/templates/dashboard.html"))
	loginTemplate          = template.Must(template.ParseFiles("./web/templates/base.html", "./web/templates/login.html"))
	signupTemplate         = template.Must(template.ParseFiles("./web/templates/base.html", "./web/templates/signup.html"))
	profileTemplate        = template.Must(template.ParseFiles("./web/templates/base.html", "./web/templates/profile.html"))
	profileEditTemplate    = template.Must(template.ParseFiles("./web/templates/base.html", "./web/templates/profileEdit.html"))
	profileDeleteTemplate  = template.Must(template.ParseFiles("./web/templates/base.html", "./web/templates/profileDelete.html"))
	postTemplate           = template.Must(template.ParseFiles("./web/templates/base.html", "./web/templates/post.html"))
	postEditTemplate       = template.Must(template.ParseFiles("./web/templates/base.html", "./web/templates/postEdit.html"))
	notFoundTemplate       = template.Must(template.ParseFiles("./web/templates/base.html", "./web/templates/notfound.html"))
	termsOfServiceTemplate = template.Must(template.ParseFiles("./web/templates/base.html", "./web/templates/legal/terms-of-service.html"))
	privacyPolicyTemplate  = template.Must(template.ParseFiles("./web/templates/base.html", "./web/templates/legal/privacy-policy.html"))
)

func New(secret string, data *models.DataSource) http.Handler {
	router := &Router{Data: data, SessionSecret: secret}
	serveMux := mux.NewRouter()
	serveMux.HandleFunc("/auth/login", router.login).Methods("GET")
	serveMux.HandleFunc("/auth/login", router.loginSubmit).Methods("POST")
	serveMux.HandleFunc("/auth/signup", router.signup).Methods("GET")
	serveMux.HandleFunc("/auth/signup", router.signupSubmit).Methods("POST")
	serveMux.HandleFunc("/auth/logout", router.logout).Methods("GET")
	serveMux.HandleFunc("/", router.dashboard).Methods("GET")
	serveMux.HandleFunc("/profile", router.profileRedirect).Methods("GET")
	serveMux.HandleFunc("/profile/edit", router.profileEdit).Methods("GET")
	serveMux.HandleFunc("/profile/edit", router.profileEditSubmit).Methods("POST")
	serveMux.HandleFunc("/profile/delete", router.profileDelete).Methods("GET")
	serveMux.HandleFunc("/profile/delete", router.profileDeleteSubmit).Methods("POST")
	serveMux.HandleFunc("/post", router.postNew).Methods("GET")
	serveMux.HandleFunc("/post", router.postSubmit).Methods("POST")
	serveMux.HandleFunc("/legal/privacy-policy", router.privacyPolicy).Methods("GET")
	serveMux.HandleFunc("/legal/terms-of-service", router.termsOfService).Methods("GET")
	serveMux.HandleFunc("/{user}", router.profile).Methods("GET")
	serveMux.HandleFunc("/{user}/{post}", router.postRedirect).Methods("GET")
	serveMux.HandleFunc("/{user}/{post}/", router.post).Methods("GET")
	serveMux.HandleFunc("/{user}/{post}/edit", router.postEdit).Methods("GET")
	serveMux.HandleFunc("/{user}/{post}/delete", router.postDelete).Methods("GET")
	return serveMux
}

type Context struct {
	ErrorMessage string
	HeadControls bool
	SignedIn     bool
	UserID       uint
}

type DashboardPost struct {
	Title, Author, ID, Date string
}

type DashboardUser struct {
	Name, MemberSince string
}

type DashboardContext struct {
	Context
	LatestPosts []DashboardPost
	LatestUsers []DashboardUser
}

type Router struct {
	Data          *models.DataSource
	SessionSecret string
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

func (router *Router) postRedirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	http.Redirect(w, r, fmt.Sprintf("/%s/%s/", vars["user"], vars["post"]), http.StatusPermanentRedirect)
}

func (router *Router) renderNotFound(w http.ResponseWriter, r *http.Request, topic string) {
	ctx := struct {
		Context
		Topic string
	}{Context: *router.defaultContext(r), Topic: topic}
	if err := notFoundTemplate.Execute(w, ctx); err != nil {
		log.Errorln("Failed to render NotFound:", err)
	}
}

func (router *Router) defaultContext(r *http.Request) *Context {
	sessionCookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return &Context{
			SignedIn:     false,
			HeadControls: true,
		}
	}
	claims := tokenClaims{}
	if _, err := jwt.ParseWithClaims(sessionCookie.Value, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(router.SessionSecret), nil
	}); err != nil {
		log.Errorln("Received invalid token:", err)
		return &Context{
			SignedIn:     false,
			HeadControls: true,
		}
	}
	return &Context{
		SignedIn:     true,
		UserID:       claims.ID,
		HeadControls: true,
	}
}

func (router *Router) login(w http.ResponseWriter, r *http.Request) {
	ctx := router.defaultContext(r)
	ctx.HeadControls = false
	loginTemplate.Execute(w, ctx)
}

type tokenClaims struct {
	jwt.StandardClaims
	Username string
	ID       uint
}

func (router *Router) loginSubmit(w http.ResponseWriter, r *http.Request) {
	var (
		email    = r.FormValue("email")
		password = r.FormValue("password")
	)
	id, err := router.Data.HasUser(email, []byte(password))
	if err != nil {
		ctx := router.defaultContext(r)
		ctx.ErrorMessage = "User identity does not exist."
		loginTemplate.Execute(w, ctx)
		return
	}
	claims := tokenClaims{
		ID: id,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	tokenString, err := token.SignedString([]byte(router.SessionSecret))
	if err != nil {
		log.Errorln("Could not sign token", err)
		ctx := router.defaultContext(r)
		ctx.ErrorMessage = "Unexpected internal error, please try again."
		loginTemplate.Execute(w, ctx)
		return
	}
	expiresAt := time.Now().Add(time.Hour)
	cookie := http.Cookie{Path: "/", Name: sessionCookieName, Value: tokenString, Expires: expiresAt}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (router *Router) signup(w http.ResponseWriter, r *http.Request) {
	ctx := router.defaultContext(r)
	ctx.HeadControls = false
	signupTemplate.Execute(w, ctx)
}

func (router *Router) dashboardContext(r *http.Request) *DashboardContext {
	ctx := DashboardContext{
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
	ctx.LatestPosts = make([]DashboardPost, len(posts))
	for _, post := range posts {
		user, err := router.Data.GetUser(post.UserID)
		if err != nil {
			log.Errorln("Failed to fetch user:", err)
			continue
		}
		ctx.LatestPosts[j] = DashboardPost{
			Title:  post.Title,
			Author: user.Name,
			ID:     strconv.FormatUint(uint64(post.ID), 10),
			Date:   humanize.Time(post.CreatedAt),
		}
		j++
	}
	ctx.LatestUsers = make([]DashboardUser, len(users))
	for i, user := range users {
		ctx.LatestUsers[i] = DashboardUser{
			Name:        user.Name,
			MemberSince: humanize.Time(user.CreatedAt),
		}
	}
	return &ctx
}

func (router *Router) dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := router.dashboardContext(r)
	if err := dashboardTemplate.Execute(w, ctx); err != nil {
		log.Errorln("Failed to render dashboard:", err)
	}
}

func (router *Router) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Path: "/", Value: "", Name: sessionCookieName, Expires: time.Now()})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (router *Router) signupSubmit(w http.ResponseWriter, r *http.Request) {
	var (
		name            = r.FormValue("username")
		email           = r.FormValue("email")
		password        = r.FormValue("password")
		passwordConfirm = r.FormValue("password_confirm")
		acceptTOS       = r.FormValue("accept_tos")
	)

	var (
		ctx        = router.defaultContext(r)
		errMessage string
	)
	if password != passwordConfirm {
		errMessage = "Passwords do not match."
	} else if acceptTOS != "on" {
		errMessage = "You have to accept the Terms of Service and Privacy Policy."
	} else if !router.Data.ValidateEmail(email) {
		errMessage = "Email must be an egligible email address."
	} else if !router.Data.ValidatePassword(password) {
		errMessage = "Password must have a minimum length of 8 characters."
	} else if !router.Data.ValidateName(name) {
		errMessage = "Username must only consist of alphanumerics."
	} else if router.Data.EmailExists(email) {
		errMessage = "Email already exists."
	} else if router.Data.NameExists(name) {
		errMessage = "Name already exists."
	}

	if errMessage != "" {
		ctx.ErrorMessage = errMessage
		signupTemplate.Execute(w, ctx)
		return
	}

	if err := router.Data.AddUser(name, email, []byte(password)); err != nil {
		ctx.ErrorMessage = "Unknown error occured, please try again"
		signupTemplate.Execute(w, ctx)
		return
	}
	log.Debugln("User", name, "just signed up")
	if err := signupSuccessTemplate.Execute(w, ctx); err != nil {
		log.Errorln("Failed to render signup success:", err)
	}
}

type ProfilePost struct {
	Title  string
	Date   string
	Author string
	ID     uint
}

type ProfileContext struct {
	Context
	Name        string
	Biography   string
	MemberSince string
	PostCount   int
	Self        bool
	Posts       []ProfilePost
}

func (router *Router) profile(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["user"]
	user, err := router.Data.GetUserByName(name)
	if err != nil {
		log.Debugln("Failed to find user", name)
		router.renderNotFound(w, r, "profile")
		return
	}
	posts, err := router.Data.GetPostsDesc(user.ID)
	if err != nil {
		log.Errorln("Failed to get posts for user", user.ID)
		router.renderNotFound(w, r, "profile")
		return
	}
	ctx := router.defaultContext(r)
	profileCtx := ProfileContext{
		Context:     *ctx,
		Name:        user.Name,
		Biography:   user.Biography,
		MemberSince: user.CreatedAt.Format(timeFormat),
		PostCount:   len(posts),
		Self:        ctx.SignedIn && ctx.UserID == user.ID,
		Posts:       make([]ProfilePost, len(posts)),
	}
	for i := range posts {
		profileCtx.Posts[i] = ProfilePost{
			Title:  posts[i].Title,
			Date:   posts[i].CreatedAt.Format(timeFormat),
			Author: user.Name,
			ID:     posts[i].ID,
		}
	}
	if err := profileTemplate.Execute(w, profileCtx); err != nil {
		log.Errorln("Failed to render profile:", err)
	}
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
	profileCtx := ProfileContext{
		Context:   *ctx,
		Name:      user.Name,
		Biography: user.Biography,
	}
	if err := profileEditTemplate.Execute(w, profileCtx); err != nil {
		log.Errorln("Failed to render profile edit:", err)
	}
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
	profileCtx := ProfileContext{
		Context:   *ctx,
		Name:      user.Name,
		Biography: biography,
	}
	if !router.Data.ValidateBiography(biography) {
		profileCtx.ErrorMessage = "Your biography must have at max 240 characters."
		profileEditTemplate.Execute(w, profileCtx)
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
	if err := profileDeleteTemplate.Execute(w, ctx); err != nil {
		log.Errorln("Failed to render profile delete:", err)
	}
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

func (router *Router) postDelete(w http.ResponseWriter, r *http.Request) {
	ctx := router.postContext(r)
	if err := router.Data.DeletePost(ctx.UserID, ctx.ID); err != nil {
		log.Errorln("Failed to delete post:", err)
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type PostContext struct {
	Context
	ID      uint
	Author  string
	Title   string
	Content string
	Date    string
	Self    bool
}

func (router *Router) postNew(w http.ResponseWriter, r *http.Request) {
	ctx := router.defaultContext(r)
	if !ctx.SignedIn {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}
	postCtx := PostContext{Context: *ctx}
	if err := postEditTemplate.Execute(w, postCtx); err != nil {
		log.Errorln("Failed to render post edit:", err)
	}
}

func (router *Router) postEdit(w http.ResponseWriter, r *http.Request) {
	ctx := router.defaultContext(r)
	if !ctx.SignedIn {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}
	postCtx := router.postContext(r)
	if err := postEditTemplate.Execute(w, postCtx); err != nil {
		log.Errorln("Failed to render post edit:", err)
	}
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
			if err := postEditTemplate.Execute(w, postCtx); err != nil {
				log.Errorln("Failed to render post edit:", err)
			}
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
			if err := postEditTemplate.Execute(w, postCtx); err != nil {
				log.Errorln("Failed to render post edit:", err)
			}
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
	return PostContext{
		Context: *ctx,
		Self:    ctx.SignedIn && ctx.UserID == post.UserID,
		Author:  user.Name,
		ID:      post.ID,
		Title:   post.Title,
		Content: post.Content,
		Date:    post.CreatedAt.Format(timeFormat),
	}
}

func (router *Router) postContext(r *http.Request) PostContext {
	userName := mux.Vars(r)["user"]
	postID := mux.Vars(r)["post"]
	id, _ := strconv.ParseUint(postID, 10, 64)
	return router.postContextWithID(r, userName, uint(id))
}

func (router *Router) post(w http.ResponseWriter, r *http.Request) {
	ctx := router.postContext(r)
	if err := postTemplate.Execute(w, ctx); err != nil {
		log.Errorln("Failed to render post:", err)
	}
}

func (router *Router) termsOfService(w http.ResponseWriter, r *http.Request) {
	if err := termsOfServiceTemplate.Execute(w, router.defaultContext(r)); err != nil {
		log.Errorln("Failed to render terms of service:", err)
	}
}

func (router *Router) privacyPolicy(w http.ResponseWriter, r *http.Request) {
	if err := privacyPolicyTemplate.Execute(w, router.defaultContext(r)); err != nil {
		log.Errorln("Failed to render privacy policy:", err)
	}
}
