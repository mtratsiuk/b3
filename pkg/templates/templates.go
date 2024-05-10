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
	post *template.Template
	home *template.Template
}

func New() (Templates, error) {
	post, err := template.ParseFS(viewsFs, "base.html", "post.html")
	if err != nil {
		return Templates{}, err
	}

	home, err := template.ParseFS(viewsFs, "base.html", "home.html")
	if err != nil {
		return Templates{}, err
	}

	t := Templates{post, home}
	return t, nil
}

type PostData struct {
	Title     string
	PostHtml  template.HTML
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (t Templates) RenderPost(wr io.Writer, data PostData) error {
	return t.post.ExecuteTemplate(wr, "post.html", data)
}

type HomeData struct {
	Title string
	Posts []HomePostData
}

type HomePostData struct {
	Url         string
	Title       template.HTML
	Description template.HTML
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (t Templates) RenderHome(wr io.Writer, data HomeData) error {
	return t.home.ExecuteTemplate(wr, "home.html", data)
}
