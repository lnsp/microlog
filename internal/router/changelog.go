package router

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/russross/blackfriday"
)

type changelogContext struct {
	Context
	Changelog template.HTML
}

const changelogPath = "./CHANGELOG.md"

var changelogHTML template.HTML
var changelogOnce sync.Once

func (router *Router) changelog(w http.ResponseWriter, r *http.Request) {
	changelogOnce.Do(func() {
		bytes, err := ioutil.ReadFile(changelogPath)
		if err != nil {
			log.Errorln("Failed to read changelog:", err)
			return
		}
		changelogHTML = template.HTML(blackfriday.MarkdownCommon(bytes))
	})
	ctx := changelogContext{
		Context:   *router.defaultContext(r),
		Changelog: changelogHTML,
	}
	router.render(changelogTemplate, w, ctx)
}
