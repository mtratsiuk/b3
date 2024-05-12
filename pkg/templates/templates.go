package templates

import (
	"embed"
	"html/template"
	"io"
	"time"

	"github.com/mtratsiuk/b3/pkg/config"
)

//go:embed *.html
var viewsFs embed.FS

//go:embed base.css
var baseCss template.CSS

//go:embed base.js
var baseJs template.JS

type Templates struct {
	config config.Config
	post   *template.Template
	home   *template.Template
}

func New(cfg config.Config) (Templates, error) {
	post, err := template.ParseFS(viewsFs, "base.html", "components.html", "post.html")
	if err != nil {
		return Templates{}, err
	}

	home, err := template.ParseFS(viewsFs, "base.html", "components.html", "home.html")
	if err != nil {
		return Templates{}, err
	}

	t := Templates{cfg, post, home}
	return t, nil
}

type BaseData[T any] struct {
	Title    string
	Css      template.CSS
	Js       template.JS
	Config   config.Config
	PageData T
}

type PostData struct {
	Title     string
	PostHtml  template.HTML
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (t Templates) RenderPost(wr io.Writer, data PostData) error {
	return t.post.ExecuteTemplate(wr, "post.html", BaseData[PostData]{
		Title:    data.Title,
		Css:      baseCss,
		Js:       baseJs,
		Config:   t.config,
		PageData: data,
	})
}

type HomeData struct {
	Title string
	Posts []HomePostData
}

type HomePostData struct {
	Id          string
	Url         string
	Title       template.HTML
	Description template.HTML
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (t Templates) RenderHome(wr io.Writer, data HomeData) error {
	return t.home.ExecuteTemplate(wr, "home.html", BaseData[HomeData]{
		Title:    data.Title,
		Css:      baseCss,
		Js:       baseJs,
		Config:   t.config,
		PageData: data,
	})
}
