package app

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/mtratsiuk/b3/pkg/cdn"
	"github.com/mtratsiuk/b3/pkg/config"
	"github.com/mtratsiuk/b3/pkg/templates"
	"github.com/mtratsiuk/b3/pkg/timestamper"
	"github.com/mtratsiuk/b3/pkg/utils"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
)

type Params struct {
	Log      *slog.Logger
	Verbose  bool
	RootPath string
	Prod     bool
	DryRun   bool
}

type App struct {
	log         *slog.Logger
	params      Params
	config      config.Config
	outDirPath  string
	timestamper timestamper.Timestamper
	templates   templates.Templates
	cdn         cdn.Cdn
}

type Post struct {
	Id           PostId
	FilePath     string
	HtmlFilePath string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Title        template.HTML
	Description  template.HTML
}

type PostId string

type Posts = map[PostId]*Post

func New(params Params) (App, error) {
	app := App{}

	cfg, err := config.New(params.RootPath)
	if err != nil {
		return App{}, fmt.Errorf("app.New: failed to create config: %v", err)
	}
	params.Log.Debug(fmt.Sprintf("app.New: created config: %v", cfg))

	tmplts, err := templates.New(cfg)
	if err != nil {
		return App{}, fmt.Errorf("app.New: failed to load templates: %v", err)
	}

	if cfg.DotEnvPath != "" {
		if err := utils.LoadDotEnv(filepath.Join(params.RootPath, cfg.DotEnvPath)); err != nil {
			return App{}, fmt.Errorf("app.New: failed to load env variables from dotenv file: %v", err)
		}
		params.Log.Debug("app.New: loaded dotenv file")
	}

	app.log = params.Log
	app.params = params
	app.config = cfg
	app.outDirPath = filepath.Join(params.RootPath, cfg.OutDirPath)
	app.timestamper = timestamper.NewGit()
	app.templates = tmplts

	if cfg.AssetsToUploadRegexp != "" {
		cdn, err := cdn.New()
		if err != nil {
			return App{}, fmt.Errorf("app.New: failed to create cdn: %v", err)
		}
		app.cdn = cdn
	}

	return app, nil
}

func (app *App) ResolveRelativePath(path string) string {
	return filepath.Join(app.params.RootPath, path)
}

func (app *App) Build() (Posts, error) {
	if err := os.MkdirAll(app.outDirPath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("app.Build: failed to create out directory: %v", err)
	}

	if err := app.copyAssets(); err != nil {
		return nil, fmt.Errorf("app.Build: failed to copy assets to out directory: %v", err)
	}

	posts, err := app.renderPosts()
	if err != nil {
		return nil, fmt.Errorf("app.Build: failed to render posts: %v", err)
	}

	if err := app.renderHome(posts); err != nil {
		return nil, fmt.Errorf("app.Build: failed to render home page: %v", err)
	}

	return posts, nil
}

func (app *App) Cdn() error {
	if err := app.uploadAssets(); err != nil {
		return fmt.Errorf("app.Cdn: failed to upload assets to cdn: %v", err)
	}

	return nil
}

func (app *App) renderPosts() (Posts, error) {
	posts := make(Posts, 0)

	for _, pg := range app.config.PostsGlob {
		glob := app.ResolveRelativePath(pg)

		matches, err := filepath.Glob(glob)

		if err != nil {
			return posts, fmt.Errorf("renderPosts: failed to match glob pattern '%v': %v", glob, err)
		}

		for _, p := range matches {
			app.log.Debug(fmt.Sprintf("renderPosts: processing post match: %v", p))

			filename := filepath.Base(p)
			title, _ := strings.CutSuffix(filename, filepath.Ext(filename))

			post := Post{}
			post.Id = PostId(title)
			post.FilePath = p

			createdAt, err := app.timestamper.CreatedAt(p)
			if err != nil {
				app.log.Warn(fmt.Sprintf("renderPosts: failed to read CreatedAt time: %v", err))
			}
			post.CreatedAt = createdAt

			updatedAt, err := app.timestamper.UpdatedAt(p)
			if err != nil {
				app.log.Warn(fmt.Sprintf("renderPosts: failed to read UpdatedAt time: %v", err))
			}
			post.UpdatedAt = updatedAt

			err = app.renderPost(&post)
			if err != nil {
				return posts, fmt.Errorf("renderPosts: failed to render post %v: %v", post, err)
			}
			app.log.Debug(fmt.Sprintf("renderPosts: rendered post: %v", post))

			posts[post.Id] = &post
		}
	}

	return posts, nil
}

func (app *App) renderPost(post *Post) error {
	in, err := os.ReadFile(post.FilePath)
	if err != nil {
		return err
	}

	md := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert(in, &buf); err != nil {
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
		Title:       title,
		Description: utils.TrimText(utils.StripHtml(description), app.config.TrimPostOgDescriptionsAt),
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
		PostHtml:    template.HTML(html),
	}

	postOutDirPath := filepath.Join(app.outDirPath, strings.TrimPrefix(filepath.Dir(post.FilePath), filepath.Clean(app.params.RootPath)))
	if err := os.MkdirAll(postOutDirPath, os.ModePerm); err != nil {
		return err
	}
	post.HtmlFilePath = filepath.Join(postOutDirPath, string(post.Id)+".html")

	out, err := os.Create(post.HtmlFilePath)
	if err != nil {
		return err
	}
	defer out.Close()

	return app.templates.RenderPost(out, data)
}

func (app *App) renderHome(posts map[PostId]*Post) error {
	data := templates.HomeData{}
	data.Title = app.config.DocTitle
	data.Description = app.config.DocDescription
	data.Posts = make([]templates.HomePostData, 0)

	for _, p := range posts {
		url := filepath.Join(".", strings.TrimPrefix(p.HtmlFilePath, filepath.Clean(app.outDirPath)))

		if app.params.Prod && app.config.StripHtmlExtInProdLinks {
			url, _ = strings.CutSuffix(url, ".html")
		}

		data.Posts = append(data.Posts, templates.HomePostData{
			Id:          string(p.Id),
			Title:       p.Title,
			Description: p.Description,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
			Url:         url,
		})
	}

	slices.SortFunc(data.Posts, func(a, b templates.HomePostData) int {
		createdAtCmp := b.CreatedAt.Compare(a.CreatedAt)

		if createdAtCmp == 0 {
			return strings.Compare(b.Id, a.Id)
		}

		return createdAtCmp
	})

	app.log.Debug(fmt.Sprintf("renderHome: data: %v", data))

	out, err := os.Create(filepath.Join(app.outDirPath, "index.html"))
	if err != nil {
		return err
	}
	defer out.Close()

	return app.templates.RenderHome(out, data)
}

func (app *App) uploadAssets() error {
	if app.config.AssetsToUploadRegexp == "" {
		app.log.Debug("uploadAssets: nothing to do, `assets_to_upload_regexp` is not defined")
		return nil
	}

	for _, pg := range app.config.PostsGlob {
		glob := app.ResolveRelativePath(pg)

		matches, err := filepath.Glob(glob)

		if err != nil {
			return fmt.Errorf("uploadAssets: failed to match glob pattern '%v': %v", glob, err)
		}

		for _, m := range matches {
			app.log.Debug(fmt.Sprintf("uploadAssets: processing file %v", m))

			post, err := os.ReadFile(m)
			if err != nil {
				return fmt.Errorf("uploadAssets: failed read post file: %v", err)
			}

			updatedPost, err := app.uploadAssetsForPost(string(post), filepath.Dir(m))
			if err != nil {
				return fmt.Errorf("uploadAssets: failed process post %v: %v", m, err)
			}

			if err := os.WriteFile(m, []byte(updatedPost), 0644); err != nil {
				return fmt.Errorf("uploadAssets: failed to write updated post %v: %v", m, err)
			}
		}
	}

	return nil
}

var imageRe = regexp.MustCompile(`!\[.*\]\((.*)\)`)

func (app *App) uploadAssetsForPost(content, postDirPath string) (string, error) {
	uploadRe := regexp.MustCompile(app.config.AssetsToUploadRegexp)
	updatedContent := content

	for _, match := range imageRe.FindAllStringSubmatch(content, -1) {
		imageMd := match[0]
		imagePath := match[1]
		upload := uploadRe.MatchString(imagePath)

		if !upload {
			app.log.Debug(fmt.Sprintf("uploadAssetsForPost: skipping asset: %v", imageMd))
			continue
		}

		app.log.Debug(fmt.Sprintf("uploadAssetsForPost: processing asset: %v", imageMd))

		if app.params.DryRun {
			continue
		}

		publicUrl, err := app.cdn.UploadAsset(filepath.Join(postDirPath, imagePath))
		if err != nil {
			return content, fmt.Errorf("uploadAssetsForPost: failed to upload asset: %v: %v", imageMd, err)
		}

		app.log.Debug(fmt.Sprintf("uploadAssetsForPost: uploaded asset: %v to %v", imageMd, publicUrl))

		updatedMd := strings.ReplaceAll(imageMd, imagePath, publicUrl)
		updatedContent = strings.ReplaceAll(updatedContent, imageMd, updatedMd)
	}

	return updatedContent, nil
}

func (app *App) copyAssets() error {
	for _, dir := range app.config.AssetsDirPath {
		if err := utils.CopyDir(app.ResolveRelativePath(dir), filepath.Join(app.outDirPath, dir)); err != nil {
			return err
		}
	}

	return nil
}

// TODO: walk ast
func getPostTitleHtml(html string) (string, error) {
	left := strings.Index(html, "<h")
	right := strings.Index(html, "</h")

	if left == -1 || right == -1 {
		return "", fmt.Errorf("getPostTitleHtml: expected post to have at least one heading element")
	}

	return html[strings.Index(html[left:], ">")+1 : right], nil
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
