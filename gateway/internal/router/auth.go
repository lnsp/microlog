package router

import (
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/lnsp/microlog/gateway/pkg/tokens"
	"github.com/lnsp/microlog/gateway/pkg/utils"
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
		log.WithFields(logrus.Fields{
			"token": sessionCookie.Value,
			"type":  "invalid token",
		}).WithError(err).Error("failed to create context")
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
	email, userID, err := router.EmailClient.Verify(token, tokens.PurposeConfirmation)
	ctx := emailContext{
		Context: *router.defaultContext(r),
		Success: false,
	}
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"token": token,
			"addr":  utils.RemoteHost(r),
			"type":  "invalid token",
		}).Debug("attempt to confirm identity")
		router.render(confirmTemplate, w, ctx)
		return
	}
	if err := router.Data.ConfirmIdentity(userID, email); err != nil {
		log.WithFields(logrus.Fields{
			"token": token,
			"email": email,
			"id":    userID,
			"addr":  utils.RemoteHost(r),
		}).WithError(err).Error("failed to confirm identity")
		router.render(confirmTemplate, w, ctx)
		return
	}
	ctx.Success = true
	log.WithFields(logrus.Fields{
		"token": token,
		"email": email,
		"id":    userID,
		"addr":  utils.RemoteHost(r),
	}).Debug("confirmed identity")
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
	id, err := router.Data.IdentityByEmail(email)
	if err == nil && id.Confirmed {
		if err := router.EmailClient.SendPasswordReset(id.UserID, email); err != nil {
			ctx.Success = false
			ctx.ErrorMessage = "Unexpected internal error, please try again."
			log.WithFields(logrus.Fields{
				"addr":  utils.RemoteHost(r),
				"email": email,
				"id":    id.UserID,
			}).WithError(err).Error("failed to send password reset email")
		} else {
			log.WithFields(logrus.Fields{
				"addr":  utils.RemoteHost(r),
				"email": email,
				"id":    id.UserID,
			}).Debug("requested password reset")
		}
	} else if err != nil {
		log.WithFields(logrus.Fields{
			"addr":  utils.RemoteHost(r),
			"email": email,
			"type":  "unknown identity",
		}).WithError(err).Debug("attempt to request password reset")
	} else if !id.Confirmed {
		log.WithFields(logrus.Fields{
			"addr":  utils.RemoteHost(r),
			"email": email,
			"type":  "unconfirmed identity",
		}).WithError(err).Debug("attempt to request password reset")
	}
	router.render(forgotTemplate, w, ctx)
}

func (router *Router) reset(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query()["token"][0]
	_, _, err := router.EmailClient.Verify(token, tokens.PurposeReset)
	ctx := emailContext{
		Context: *router.defaultContext(r),
		Success: false,
	}
	if err != nil {
		router.render(resetTemplate, w, ctx)
		return
	}
	ctx.Success = true
	router.render(resetTemplate, w, ctx)
}

func (router *Router) resetSubmit(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query()["token"][0]
	email, userID, err := router.EmailClient.Verify(token, tokens.PurposeReset)
	ctx := emailContext{
		Context: *router.defaultContext(r),
		Success: false,
	}
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"token": token,
			"addr":  utils.RemoteHost(r),
			"type":  "invalid token",
		}).Debug("attempt to reset password")
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
		log.WithFields(logrus.Fields{
			"id":    userID,
			"token": token,
			"addr":  utils.RemoteHost(r),
			"type":  "password mismatch",
		}).Debug("attempt to reset password")
		router.render(resetTemplate, w, ctx)
		return
	}
	if !router.Data.ValidatePassword(password) {
		ctx.ErrorMessage = "Password must be at min 8 characters long."
		log.WithFields(logrus.Fields{
			"id":    userID,
			"token": token,
			"addr":  utils.RemoteHost(r),
			"type":  "invalid password",
		}).Debug("attempt to reset password")
		router.render(resetTemplate, w, ctx)
		return
	}
	if err := router.Data.ResetPassword(userID, email, []byte(password)); err != nil {
		ctx.ErrorMessage = "Unexpected internal error, please try again."
		log.WithFields(logrus.Fields{
			"id":    userID,
			"email": email,
			"token": token,
			"addr":  utils.RemoteHost(r),
		}).WithError(err).Error("failed to reset password")
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
		log.WithFields(logrus.Fields{
			"email": email,
			"addr":  utils.RemoteHost(r),
		}).WithError(err).Debug("failed login attempt")
		router.render(loginTemplate, w, ctx)
		return
	}
	if !confirmed {
		ctx := router.defaultContext(r)
		ctx.ErrorMessage = "User identity is not confirmed."
		log.WithFields(logrus.Fields{
			"id":    id,
			"email": email,
			"addr":  utils.RemoteHost(r),
		}).Debug("unconfirmed login attempt")
		router.render(loginTemplate, w, ctx)
		return
	}
	user, err := router.Data.User(id)
	if err != nil {
		ctx := router.defaultContext(r)
		ctx.ErrorMessage = "Unexpected internal error, please try again."
		log.WithFields(logrus.Fields{
			"id":   id,
			"addr": utils.RemoteHost(r),
		}).WithError(err).Error("failed to login user")
		router.render(loginTemplate, w, ctx)
		return
	}
	token, err := tokens.CreateSessionToken(router.SessionSecret, user.Name, user.ID)
	if err != nil {
		ctx := router.defaultContext(r)
		ctx.ErrorMessage = "Unexpected internal error, please try again."
		log.WithFields(logrus.Fields{
			"id":   id,
			"addr": utils.RemoteHost(r),
		}).WithError(err).Error("failed to login user")
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
	log.WithFields(logrus.Fields{
		"id":   user.ID,
		"name": user.Name,
		"addr": utils.RemoteHost(r),
	}).Info("user login")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type signupContext struct {
	Context
	Name      string
	Email     string
	Password  string
	AcceptTOS bool
}

func (router *Router) signup(w http.ResponseWriter, r *http.Request) {
	ctx := &signupContext{
		Context: *router.defaultContext(r),
	}
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
		ctx = &signupContext{
			Name:      name,
			Email:     email,
			Password:  password,
			AcceptTOS: acceptTOS == "on",
			Context:   *router.defaultContext(r),
		}
		errMessage string
	)
	if password != passwordConfirm {
		errMessage = "Passwords do not match."
		ctx.Password = ""
	} else if acceptTOS != "on" {
		errMessage = "You have to accept the Terms of Service and Privacy Policy."
		ctx.AcceptTOS = false
	} else if !router.Data.ValidateEmail(email) {
		errMessage = "Email must be an egligible email address."
		ctx.Email = ""
	} else if !router.Data.ValidatePassword(password) {
		errMessage = "Password must have a minimum length of 8 characters."
		ctx.Password = ""
	} else if !router.Data.ValidateName(name) {
		errMessage = "Username must only consist of alphanumerics."
		ctx.Name = ""
	} else if router.Data.EmailExists(email) {
		errMessage = "Email already exists."
		ctx.Email = ""
	} else if router.Data.NameExists(name) {
		errMessage = "Name already exists."
		ctx.Name = ""
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

	if err := router.EmailClient.SendConfirmation(userID, email); err != nil {
		ctx.ErrorMessage = "Internal error occured, please try again."
		router.render(signupTemplate, w, ctx)
		return
	}

	log.WithFields(logrus.Fields{
		"id":    userID,
		"user":  name,
		"email": email,
		"addr":  utils.RemoteHost(r),
	}).Debug("new user signed up")
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
	logrus.WithFields(logrus.Fields{
		"id":   ctx.UserID,
		"addr": utils.RemoteHost(r),
	}).Info("deleted user account")
	http.Redirect(w, r, "/auth/logout", http.StatusSeeOther)
}