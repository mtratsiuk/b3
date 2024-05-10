package app

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mtratsiuk/b3/pkg/config"
	"github.com/mtratsiuk/b3/pkg/templates"
	"github.com/mtratsiuk/b3/pkg/timestamper"
	"github.com/yuin/goldmark"
)

type Params struct {
	Log      *slog.Logger
	Verbose  bool
	RootPath string
}

type App struct {
	log         *slog.Logger
	params      Params
	config      config.Config
	outDirPath  string
	timestamper timestamper.Timestamper
	templates   templates.Templates
}

type Post struct {
	Id          PostId
	FilePath    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Title       template.HTML
	Description template.HTML
}

type PostId string

func New(params Params) (App, error) {
	cfg, err := config.New(params.RootPath)
	if err != nil {
		return App{}, fmt.Errorf("app.New: failed to create config: %v", err)
	}
	params.Log.Debug(fmt.Sprintf("app.New: created config: %v", cfg))

	tmplts, err := templates.New()
	if err != nil {
		return App{}, fmt.Errorf("app.New: failed to load templates: %v", err)
	}

	return App{
		log:         params.Log,
		params:      params,
		config:      cfg,
		outDirPath:  filepath.Join(params.RootPath, cfg.OutPath),
		timestamper: timestamper.NewGit(),
		templates:   tmplts,
	}, nil
}

func (app *App) ResolveRelativePath(path string) string {
	return filepath.Join(app.params.RootPath, path)
}

func (app *App) Build() (map[PostId]*Post, error) {
	posts := make(map[PostId]*Post, 0)

	if err := os.MkdirAll(app.outDirPath, os.ModePerm); err != nil {
		return posts, fmt.Errorf("app.Build: failed to create out directory: %v", err)
	}

	for _, pg := range app.config.Posts {
		glob := app.ResolveRelativePath(pg)

		matches, err := filepath.Glob(glob)

		if err != nil {
			return posts, fmt.Errorf("app.Build: failed to match glob pattern '%v': %v", glob, err)
		}

		for _, p := range matches {
			app.log.Debug(fmt.Sprintf("app.Build: processing post match: %v", p))

			filename := filepath.Base(p)
			title, _ := strings.CutSuffix(filename, filepath.Ext(filename))

			post := Post{}
			post.Id = PostId(title)
			post.FilePath = p

			createdAt, err := app.timestamper.CreatedAt(p)
			if err != nil {
				app.log.Warn(fmt.Sprintf("app.Build: failed to read CreatedAt time: %v", err))
			}
			post.CreatedAt = createdAt

			updatedAt, err := app.timestamper.UpdatedAt(p)
			if err != nil {
				app.log.Warn(fmt.Sprintf("app.Build: failed to read UpdatedAt time: %v", err))
			}
			post.UpdatedAt = updatedAt

			err = app.renderPost(&post)
			if err != nil {
				return posts, fmt.Errorf("app.Build: failed to render post %v: %v", post, err)
			}
			app.log.Debug(fmt.Sprintf("app.Build: rendered post: %v", post))

			posts[post.Id] = &post
		}
	}

	if err := app.renderHome(posts); err != nil {
		return posts, fmt.Errorf("app.Build: failed to render home page: %v", err)
	}

	return posts, nil
}

func (app *App) renderPost(post *Post) error {
	in, err := os.ReadFile(post.FilePath)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := goldmark.Convert(in, &buf); err != nil {
		return err
	}
	html := buf.String()

	title, err := getPostTitleHtml(html)
	if err != nil {
		return err
	}
	post.Title = template.HTML(title)

	description, err := getPostDescriptionHtml(html)
	if err != nil {
		return err
	}
	post.Description = template.HTML(description)

	data := templates.PostData{
		Title:     string(post.Id),
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
		PostHtml:  template.HTML(html),
	}

	out, err := os.Create(filepath.Join(app.outDirPath, string(post.Id)+".html"))
	if err != nil {
		return err
	}
	defer out.Close()

	return app.templates.RenderPost(out, data)
}

func (app *App) renderHome(posts map[PostId]*Post) error {
	data := templates.HomeData{}
	data.Title = "b3" // TODO: use config
	data.Posts = make([]templates.HomePostData, 0)

	for _, p := range posts {
		data.Posts = append(data.Posts, templates.HomePostData{
			Title:       p.Title,
			Description: p.Description,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
			Url:         fmt.Sprintf("%v.html", p.Id), // TODO: strip .html for github pages build
		})
	}

	app.log.Debug(fmt.Sprintf("renderHome: data: %v", data))

	out, err := os.Create(filepath.Join(app.outDirPath, "index.html"))
	if err != nil {
		return err
	}
	defer out.Close()

	return app.templates.RenderHome(out, data)
}

// TODO: walk ast
func getPostTitleHtml(html string) (string, error) {
	left := strings.Index(html, "<h")
	right := strings.Index(html, "</h")

	if left == -1 || right == -1 {
		return "", fmt.Errorf("getPostTitleHtml: expected post to have at least one heading element")
	}

	return html[left+4 : right], nil
}

// TODO: walk ast
func getPostDescriptionHtml(html string) (string, error) {
	left := strings.Index(html, "<p>")
	right := strings.Index(html, "</p>")

	if left == -1 || right == -1 {
		return "", fmt.Errorf("getPostDescriptionHtml: expected post to have at least one paragraph element")
	}

	return html[left+3 : right], nil
}
