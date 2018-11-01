package router

import (
	"html/template"
	"net/http"
	"os"
	"regexp"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/lnsp/microlog/gateway/internal/email"
	"github.com/lnsp/microlog/gateway/internal/models"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/json"
	"github.com/tdewolff/minify/svg"
	"github.com/tdewolff/minify/xml"
)

const (
	timeFormat          = "Monday, 2. January at 15:04"
	sessionCookieName   = "session_token"
	dashboardPostsLimit = 5
	dashboardUsersLimit = 5
)

var log = &logrus.Logger{
	Out:       os.Stderr,
	Formatter: new(logrus.JSONFormatter),
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
	reportTemplate         = template.Must(template.ParseFiles("./web/templates/base.html", "./web/templates/report.html"))
	notFoundTemplate       = template.Must(template.ParseFiles("./web/templates/base.html", "./web/templates/notfound.html"))
	confirmTemplate        = template.Must(template.ParseFiles("./web/templates/base.html", "./web/templates/confirm.html"))
	resetTemplate          = template.Must(template.ParseFiles("./web/templates/base.html", "./web/templates/reset.html"))
	forgotTemplate         = template.Must(template.ParseFiles("./web/templates/base.html", "./web/templates/forgot.html"))
	changelogTemplate      = template.Must(template.ParseFiles("./web/templates/base.html", "./web/templates/changelog.html"))
	termsOfServiceTemplate = template.Must(template.ParseFiles("./web/templates/base.html", "./web/templates/legal/terms-of-service.html"))
	privacyPolicyTemplate  = template.Must(template.ParseFiles("./web/templates/base.html", "./web/templates/legal/privacy-policy.html"))
	moderationTemplate     = template.Must(template.ParseFiles("./web/templates/base.html", "./web/templates/moderation.html"))
)

type Config struct {
	SessionSecret []byte
	EmailClient   *email.Client
	DataSource    *models.DataSource
	PublicAddress string
	Minify        bool
}

func New(cfg Config) http.Handler {
	router := &Router{
		Data:          cfg.DataSource,
		SessionSecret: cfg.SessionSecret,
		EmailClient:   cfg.EmailClient,
		PublicAddress: cfg.PublicAddress,
	}
	serveMux := mux.NewRouter()
	serveMux.HandleFunc("/favicon.ico", router.favicon).Methods("GET")
	serveMux.HandleFunc("/auth/login", router.login).Methods("GET")
	serveMux.HandleFunc("/auth/login", router.loginSubmit).Methods("POST")
	serveMux.HandleFunc("/auth/forgot", router.forgot).Methods("GET")
	serveMux.HandleFunc("/auth/forgot", router.forgotSubmit).Methods("POST")
	serveMux.HandleFunc("/auth/signup", router.signup).Methods("GET")
	serveMux.HandleFunc("/auth/signup", router.signupSubmit).Methods("POST")
	serveMux.HandleFunc("/auth/logout", router.logout).Methods("GET")
	serveMux.HandleFunc("/auth/confirm", router.confirm).Methods("GET")
	serveMux.HandleFunc("/auth/reset", router.reset).Methods("GET")
	serveMux.HandleFunc("/auth/reset", router.resetSubmit).Methods("POST")
	serveMux.HandleFunc("/auth/delete", router.delete).Methods("GET")
	serveMux.HandleFunc("/auth/delete", router.deleteSubmit).Methods("POST")
	serveMux.HandleFunc("/", router.dashboard).Methods("GET")
	serveMux.HandleFunc("/changelog", router.changelog).Methods("GET")
	serveMux.HandleFunc("/profile", router.profileRedirect).Methods("GET")
	serveMux.HandleFunc("/profile/edit", router.profileEdit).Methods("GET")
	serveMux.HandleFunc("/profile/edit", router.profileEditSubmit).Methods("POST")
	serveMux.HandleFunc("/post", router.postNew).Methods("GET")
	serveMux.HandleFunc("/post", router.postSubmit).Methods("POST")
	serveMux.HandleFunc("/legal/privacy-policy", router.privacyPolicy).Methods("GET")
	serveMux.HandleFunc("/legal/terms-of-service", router.termsOfService).Methods("GET")
	serveMux.HandleFunc("/moderate", router.Moderate).Methods("GET")
	serveMux.HandleFunc("/moderate/delete/{report}", router.ModerateDelete).Methods("GET")
	serveMux.HandleFunc("/moderate/close/{report}", router.ModerateClose).Methods("GET")
	serveMux.HandleFunc("/{user}", router.profile).Methods("GET")
	serveMux.HandleFunc("/{user}/{post}", router.postRedirect).Methods("GET")
	serveMux.HandleFunc("/{user}/{post}/", router.post).Methods("GET")
	serveMux.HandleFunc("/{user}/{post}/edit", router.postEdit).Methods("GET")
	serveMux.HandleFunc("/{user}/{post}/delete", router.postDelete).Methods("GET")
	serveMux.HandleFunc("/{user}/{post}/report", router.report).Methods("GET")
	serveMux.HandleFunc("/{user}/{post}/report", router.reportSubmit).Methods("POST")
	serveMux.HandleFunc("/{user}/{post}/like", router.like).Methods("GET")
	return serveMux
}

var minifier *minify.M

func init() {
	minifier = minify.New()
	minifier.AddFunc("text/css", css.Minify)
	minifier.AddFunc("text/html", html.Minify)
	minifier.AddFunc("image/svg+xml", svg.Minify)
	minifier.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	minifier.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
	minifier.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)
}

type Context struct {
	ErrorMessage string
	HeadControls bool
	SignedIn     bool
	UserID       uint
	Moderator    bool
}

type Router struct {
	EmailClient   *email.Client
	Data          *models.DataSource
	SessionSecret []byte
	EmailSecret   []byte
	PublicAddress string
	Minification  bool
}

func (router *Router) render(tmp *template.Template, w http.ResponseWriter, ctx interface{}) {
	mw := minifier.Writer("text/html", w)
	defer mw.Close()
	if err := tmp.Execute(mw, ctx); err != nil {
		log.WithFields(logrus.Fields{
			"name": tmp.Name(),
		}).WithError(err).Error("failed to render template")
	}
}

func (router *Router) favicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/static/img/microlog.png")
}

func (router *Router) renderNotFound(w http.ResponseWriter, r *http.Request, topic string) {
	ctx := struct {
		Context
		Topic string
	}{Context: *router.defaultContext(r), Topic: topic}
	router.render(notFoundTemplate, w, ctx)
}
