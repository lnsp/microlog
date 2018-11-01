package router

import "net/http"

func (router *Router) termsOfService(w http.ResponseWriter, r *http.Request) {
	router.render(termsOfServiceTemplate, w, router.defaultContext(r))
}

func (router *Router) privacyPolicy(w http.ResponseWriter, r *http.Request) {
	router.render(privacyPolicyTemplate, w, router.defaultContext(r))
}
