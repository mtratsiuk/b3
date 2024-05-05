package b3

import (
	"fmt"
	"log/slog"
)

type Params struct {
	Log      *slog.Logger
	Verbose  bool
	RootPath string
}

type App struct {
	log    *slog.Logger
	params Params
	config Config
}

func NewApp(params Params) (App, error) {
	cfg, err := NewConfig(params.RootPath)

	if err != nil {
		return App{}, fmt.Errorf("failed to create b3 app: %v", err)
	}

	return App{
		log:    params.Log,
		params: params,
		config: cfg,
	}, nil
}
