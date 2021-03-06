package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/glacier/infra"
	"github.com/mylxsw/glacier/listener"
	"github.com/mylxsw/glacier/starter/application"
	"github.com/mylxsw/glacier/web"
	"github.com/mylxsw/graphviz-server/api"
	"github.com/mylxsw/graphviz-server/config"
	"github.com/urfave/cli"
	"github.com/urfave/cli/altsrc"
)

var Version = "1.0"
var GitCommit = "5dbef13fb456f51a5d29464d"

func main() {
	app := application.Create(fmt.Sprintf("%s (%s)", Version, GitCommit[:8]))
	app.AddFlags(altsrc.NewStringFlag(cli.StringFlag{
		Name:  "listen",
		Usage: "服务监听地址",
		Value: "127.0.0.1:19921",
	}))
	app.AddFlags(altsrc.NewStringFlag(cli.StringFlag{
		Name:  "dot-bin",
		Usage: "Dot 可执行文件路径",
		Value: "/usr/local/bin/dot",
	}))
	app.AddFlags(altsrc.NewStringFlag(cli.StringFlag{
		Name:  "tmpdir",
		Usage: "临时文件存储目录",
		Value: "/tmp",
	}))
	app.AddFlags(altsrc.NewBoolFlag(cli.BoolFlag{
		Name:  "debug",
		Usage: "是否使用调试模式，调试模式下，静态资源使用本地文件",
	}))

	app.WithHttpServer(listener.FlagContext("listen"))
	app.WebAppExceptionHandler(func(ctx web.Context, err interface{}) web.Response {
		log.Errorf("error: %v, call stack: %s", err, debug.Stack())
		return nil
	})

	app.Singleton(func(c infra.FlagContext) *config.Config {
		return &config.Config{
			Listen:  c.String("listen"),
			DotBin:  c.String("dot-bin"),
			TempDir: c.String("tmpdir"),
			Debug:   c.Bool("debug"),
		}
	})

	app.Provider(api.ServiceProvider{})

	if err := app.Run(os.Args); err != nil {
		log.Errorf("exit with error: %s", err)
	}
}
