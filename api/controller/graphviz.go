package controller

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/container"
	"github.com/mylxsw/glacier/web"
	"github.com/mylxsw/graphviz-server/config"
)

type GraphvizController struct {
	cc container.Container
}

func NewGraphvizController(cc container.Container) web.Controller {
	return &GraphvizController{cc: cc}
}
func (g GraphvizController) Register(router *web.Router) {
	router.Group("/graphviz", func(router *web.Router) {
		router.Post("/", g.createImageDef)
	})

	router.Get("/preview/{id}", g.getImage)
}

var supportFileTypes = []string{"svg", "svgz", "webp", "png", "bmp", "jpg", "jpeg", "pdf", "gif"}

func (g GraphvizController) createImageDef(ctx web.Context, conf *config.Config) web.Response {
	graphDef := ctx.Body()
	fileType := strings.ToLower(ctx.InputWithDefault("type", "svg"))
	if !in(fileType, supportFileTypes) {
		return ctx.JSONError(fmt.Sprintf("invalid type, only support: %s", strings.Join(supportFileTypes, ",")), http.StatusUnprocessableEntity)
	}

	finger, err := g.rebuildImageFromDefinition(conf, graphDef, fileType)
	if err != nil {
		return ctx.JSONError(err.Error(), http.StatusInternalServerError)
	}

	return ctx.JSON(web.M{"url": fmt.Sprintf("/api/preview/%s.%s", finger, fileType)})
}

func (g GraphvizController) rebuildImageFromDefinition(conf *config.Config, definition []byte, filetype string) (string, error) {
	finger := fmt.Sprintf("%x", md5.Sum(definition))
	if !fileExist(filepath.Join(conf.TempDir, "graphviz", finger)) {
		sourceFilepath := filepath.Join(conf.TempDir, "sources", finger+".dot")
		if !fileExist(sourceFilepath) {
			// 生成用户输入的临时文件
			_ = os.MkdirAll(filepath.Join(conf.TempDir, "sources"), os.ModePerm)
			if err := ioutil.WriteFile(sourceFilepath, definition, os.ModePerm); err != nil {
				return "", fmt.Errorf("write temp file failed for source: %w", err)
			}
		}

		if err := g.rebuildImage(conf, finger, filetype); err != nil {
			return "", fmt.Errorf("rebuild image failed: %w", err)
		}
	}

	return finger, nil
}

func (g GraphvizController) rebuildImage(conf *config.Config, finger string, filetype string) error {
	outputFilepath := filepath.Join(conf.TempDir, "graphviz", finger+"."+filetype)
	sourceFilepath := filepath.Join(conf.TempDir, "sources", finger+".dot")

	if !fileExist(sourceFilepath) {
		return errors.New("source file not exist")
	}

	// 创建图片文件
	_ = os.MkdirAll(filepath.Join(conf.TempDir, "graphviz"), os.ModePerm)
	dotCmd := exec.Command(conf.DotBin, "-T"+filetype, "-o", outputFilepath, sourceFilepath)
	stdout, err := dotCmd.Output()
	if err != nil {
		_ = os.Remove(sourceFilepath)
		return err
	}

	if log.DebugEnabled() {
		log.Debugf("dot command output: %s", stdout)
	}

	return nil
}

func (g GraphvizController) getImage(ctx web.Context, conf *config.Config) web.Response {
	id := ctx.PathVar("id")
	segs := strings.SplitN(id, ".", 2)
	if len(segs) != 2 {
		return ctx.JSONError("invalid file", http.StatusUnprocessableEntity)
	}

	destFilepath := filepath.Join(conf.TempDir, "graphviz", id)
	if !fileExist(destFilepath) {
		if err := g.rebuildImage(conf, segs[0], segs[1]); err != nil {
			return ctx.JSONError(err.Error(), http.StatusInternalServerError)
		}
	}

	return ctx.Redirect(fmt.Sprintf("/resources/%s", id), http.StatusFound)
}

// fileExist 判断文件是否存在
func fileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}

// i 判断元素是否在字符串数组中
func in(val string, items []string) bool {
	for _, item := range items {
		if item == val {
			return true
		}
	}

	return false
}
