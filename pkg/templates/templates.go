package templates

import (
	"embed"
	"html/template"
	"io"
	"time"
)

//go:embed *.html
var viewsFs embed.FS

type Templates struct {
	views *template.Template
}

func New() (Templates, error) {
	views, err := template.ParseFS(viewsFs, "*.html")

	if err != nil {
		return Templates{}, err
	}

	t := Templates{views}

	return t, nil
}

type PostViewModel struct {
	Title     string
	PostHtml  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (t Templates) RenderPost(wr io.Writer, model PostViewModel) error {
	return t.views.ExecuteTemplate(wr, "post.html", model)
}

type HomeViewModel struct {
	Title string
}

func (t Templates) RenderHome(wr io.Writer, model PostViewModel) error {
	return t.views.ExecuteTemplate(wr, "home.html", model)
}
