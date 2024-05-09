package app

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/mtratsiuk/b3/pkg/config"
	"github.com/mtratsiuk/b3/pkg/timestamper"
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
	timestamper timestamper.Timestamper
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

	return App{
		log:         params.Log,
		params:      params,
		config:      cfg,
		timestamper: timestamper.NewGit(),
	}, nil
}

func (app *App) Build() error {
	posts := make(map[PostId]Post, 0)

	for _, pg := range app.config.Posts {
		glob := filepath.Join(app.params.RootPath, pg)

		matches, err := filepath.Glob(glob)

		if err != nil {
			return fmt.Errorf("app.Build: failed to match glob pattern '%v': %v", glob, err)
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

			// TODO: render post

			posts[post.Id] = post

			app.log.Debug(fmt.Sprintf("app.Build: created post: %v", post))
		}
	}

	// TODO: render home page

	return nil
}
