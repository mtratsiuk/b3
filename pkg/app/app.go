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
	Id        PostId
	FilePath  string
	CreatedAt time.Time
	UpdatedAt time.Time
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
			app.log.Debug(fmt.Sprintf("app.Build: finished processing post: %v", post))
		}
	}

	// TODO: render home page

	return posts, nil
}

func (app *App) renderPost(post *Post) error {
	out, err := os.Create(filepath.Join(app.outDirPath, string(post.Id) + ".html"))
	if err != nil {
		return err
	}
	defer out.Close()

	in, err := os.ReadFile(post.FilePath)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := goldmark.Convert(in, &buf); err != nil {
		return err
	}

	data := templates.PostData{
		Title: string(post.Id),
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
		PostHtml: template.HTML(buf.String()),
	}

	app.templates.RenderPost(out, data)

	return nil
}
