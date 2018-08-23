package router

import (
	"net/http"
	"time"

	"github.com/lnsp/microlog/internal/tokens"
)

type emailContext struct {
	Context
	Success bool
}

func (router *Router) defaultContext(r *http.Request) *Context {
	sessionCookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return &Context{
			SignedIn:     false,
			HeadControls: true,
		}
	}
	_, uid, ok := tokens.VerifySessionToken(router.SessionSecret, sessionCookie.Value)
	if !ok {
		log.Errorln("Received invalid token:", err)
		return &Context{
			SignedIn:     false,
			HeadControls: true,
		}
	}
	return &Context{
		SignedIn:     true,
		UserID:       uid,
		HeadControls: true,
	}
}

func (router *Router) confirm(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query()["token"][0]
	email, userID, ok := tokens.VerifyEmailToken(router.EmailSecret, token, tokens.PurposeConfirmation)
	ctx := emailContext{
		Context: *router.defaultContext(r),
		Success: false,
	}
	if !ok {
		router.render(confirmTemplate, w, ctx)
		return
	}
	if err := router.Data.ConfirmIdentity(userID, email); err != nil {
		log.Errorln("Failed to confirm identity:", err)
		router.render(confirmTemplate, w, ctx)
		return
	}
	ctx.Success = true
	router.render(confirmTemplate, w, ctx)
}

func (router *Router) forgot(w http.ResponseWriter, r *http.Request) {
	ctx := emailContext{
		Context: *router.defaultContext(r),
		Success: false,
	}
	router.render(forgotTemplate, w, ctx)
}

func (router *Router) forgotSubmit(w http.ResponseWriter, r *http.Request) {
	ctx := emailContext{
		Context: *router.defaultContext(r),
		Success: true,
	}
	email := r.FormValue("email")
	id, err := router.Data.GetIdentityByEmail(email)
	if err == nil && id.Confirmed {
		if err := router.Email.SendPasswordReset(id.UserID, email, router.PublicAddress+resetURLFormat); err != nil {
			log.Errorln("Failed to send reset email:", err)
			ctx.Success = false
			ctx.ErrorMessage = "Unexpected internal error, please try again."
		}
	}
	router.render(forgotTemplate, w, ctx)
}

func (router *Router) reset(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query()["token"][0]
	_, _, ok := tokens.VerifyEmailToken(router.EmailSecret, token, tokens.PurposeReset)
	ctx := emailContext{
		Context: *router.defaultContext(r),
		Success: false,
	}
	if !ok {
		router.render(resetTemplate, w, ctx)
		return
	}
	ctx.Success = true
	router.render(resetTemplate, w, ctx)
}

func (router *Router) resetSubmit(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query()["token"][0]
	email, userID, ok := tokens.VerifyEmailToken(router.EmailSecret, token, tokens.PurposeReset)
	ctx := emailContext{
		Context: *router.defaultContext(r),
		Success: false,
	}
	if !ok {
		router.render(resetTemplate, w, ctx)
		return
	}
	ctx.Success = true
	var (
		password        = r.FormValue("password")
		passwordConfirm = r.FormValue("password_confirm")
	)
	if password != passwordConfirm {
		ctx.ErrorMessage = "Passwords do not match."
		router.render(resetTemplate, w, ctx)
		return
	}
	if !router.Data.ValidatePassword(password) {
		ctx.ErrorMessage = "Password must be at min 8 characters long."
		router.render(resetTemplate, w, ctx)
		return
	}
	if err := router.Data.ResetPassword(userID, email, []byte(password)); err != nil {
		log.Errorln("Failed to reset password:", err)
		ctx.ErrorMessage = "Unexpected internal error, please try again."
		router.render(resetTemplate, w, ctx)
		return
	}
	ctx.ErrorMessage = "You can now log in with your new password."
	ctx.HeadControls = false
	router.render(loginTemplate, w, ctx)
}

func (router *Router) login(w http.ResponseWriter, r *http.Request) {
	ctx := router.defaultContext(r)
	ctx.HeadControls = false
	router.render(loginTemplate, w, ctx)
}

func (router *Router) loginSubmit(w http.ResponseWriter, r *http.Request) {
	var (
		email    = r.FormValue("email")
		password = r.FormValue("password")
	)
	id, confirmed, err := router.Data.HasUser(email, []byte(password))
	if err != nil {
		ctx := router.defaultContext(r)
		ctx.ErrorMessage = "User identity does not exist."
		router.render(loginTemplate, w, ctx)
		return
	}
	if !confirmed {
		ctx := router.defaultContext(r)
		ctx.ErrorMessage = "User identity is not confirmed."
		router.render(loginTemplate, w, ctx)
		return
	}
	user, err := router.Data.GetUser(id)
	if err != nil {
		ctx := router.defaultContext(r)
		ctx.ErrorMessage = "Unexpected internal error, please try again."
		router.render(loginTemplate, w, ctx)
		return
	}
	token, err := tokens.CreateSessionToken(router.SessionSecret, user.Name, user.ID)
	if err != nil {
		log.Errorln("Could not sign token", err)
		ctx := router.defaultContext(r)
		ctx.ErrorMessage = "Unexpected internal error, please try again."
		router.render(loginTemplate, w, ctx)
		return
	}
	cookie := http.Cookie{
		Path:    "/",
		Name:    sessionCookieName,
		Value:   token,
		Expires: time.Now().Add(time.Hour),
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (router *Router) signup(w http.ResponseWriter, r *http.Request) {
	ctx := router.defaultContext(r)
	ctx.HeadControls = false
	router.render(signupTemplate, w, ctx)
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
		router.render(signupTemplate, w, ctx)
		return
	}

	userID, err := router.Data.AddUser(name, email, []byte(password))
	if err != nil {
		ctx.ErrorMessage = "Internal error occured, please try again."
		router.render(signupTemplate, w, ctx)
		return
	}

	if err := router.Email.SendConfirmation(userID, email, router.PublicAddress+confirmURLFormat); err != nil {
		ctx.ErrorMessage = "Internal error occured, please try again."
		router.render(signupTemplate, w, ctx)
		return
	}

	log.Debugln("User", name, "just signed up")
	router.render(signupSuccessTemplate, w, ctx)
}

func (router *Router) delete(w http.ResponseWriter, r *http.Request) {
	ctx := router.defaultContext(r)
	if !ctx.SignedIn {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}
	router.render(profileDeleteTemplate, w, ctx)
}

func (router *Router) deleteSubmit(w http.ResponseWriter, r *http.Request) {
	ctx := router.defaultContext(r)
	if !ctx.SignedIn {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}
	router.Data.DeleteUser(ctx.UserID)
	http.Redirect(w, r, "/auth/logout", http.StatusSeeOther)
}
