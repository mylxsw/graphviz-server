package controller

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/container"
	"github.com/mylxsw/glacier/web"
	"github.com/mylxsw/graphviz-server/config"
)

type GraphvizController struct {
	cc   container.Container
	conf *config.Config
}

func NewGraphvizController(cc container.Container) web.Controller {
	conf := config.Get(cc)
	_ = os.MkdirAll(filepath.Join(conf.TempDir, "sources"), os.ModePerm)
	_ = os.MkdirAll(filepath.Join(conf.TempDir, "graphviz"), os.ModePerm)
	return &GraphvizController{cc: cc, conf: conf}
}

func (g GraphvizController) Register(router *web.Router) {
	router.Group("/graphviz", func(router *web.Router) {
		router.Post("/definition", g.AddImageDefinition)
		router.Get("/definition", g.GetDefinition)
		router.Any("/stream", g.RenderImageAsStream)
	})

	router.Get("/preview/{id}", g.LoadImage)
}

var supportFileTypes = []string{"svg", "svgz", "webp", "png", "bmp", "jpg", "jpeg", "pdf", "gif"}

// RenderImageAsStream 直接返回指定的 Graph 定义生成的图形二进制
func (g GraphvizController) RenderImageAsStream(ctx web.Context) web.Response {
	fileType := strings.ToLower(ctx.InputWithDefault("type", "svg"))
	if !in(fileType, supportFileTypes) {
		return ctx.JSONError(fmt.Sprintf("invalid type, only support: %s", strings.Join(supportFileTypes, ",")), http.StatusUnprocessableEntity)
	}

	var def = []byte(strings.TrimSpace(ctx.Input("def")))

	stream, err := g.buildImageAsStream(def, fileType)
	if err != nil {
		return ctx.JSONError(fmt.Sprintf("can not create image from definition: %v", err), http.StatusInternalServerError)
	}

	return ctx.Raw(func(w http.ResponseWriter) {
		switch fileType {
		case "svg", "svgz":
			w.Header().Set("Content-Type", "image/svg+xml")
		case "jpg", "jpeg":
			w.Header().Set("Content-Type", "image/jpeg")
		case "pdf":
			w.Header().Set("Content-Type", "application/pdf")
		default:
			w.Header().Set("Content-Type", "image/"+fileType)
		}
		_, _ = w.Write(stream)
	})
}

// GetDefinition 获取 id 对应的图片定义
func (g GraphvizController) GetDefinition(ctx web.Context) web.Response {
	id := strings.TrimSpace(ctx.Input("id"))
	if id == "" {
		return ctx.JSONError("no id provided", http.StatusUnprocessableEntity)
	}

	sourcePath := g.buildSourcePath(id + ".dot")
	if !fileExist(sourcePath) {
		return ctx.JSONError("no such file", http.StatusUnprocessableEntity)
	}

	file, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		return ctx.JSONError(err.Error(), http.StatusInternalServerError)
	}

	return ctx.JSON(web.M{
		"id":  id,
		"def": string(file),
	})
}

// AddImageDefinition 添加图片定义
func (g GraphvizController) AddImageDefinition(ctx web.Context) web.Response {
	graphDef := strings.TrimSpace(ctx.Input("def"))
	if graphDef == "" {
		return ctx.JSONError("invalid graphviz def", http.StatusUnprocessableEntity)
	}

	if !strings.HasPrefix(graphDef, "strict digraph") && !strings.HasPrefix(graphDef, "digraph") {
		return ctx.JSONError("invalid graphviz def", http.StatusUnprocessableEntity)
	}

	fileType := strings.ToLower(ctx.InputWithDefault("type", "svg"))
	if !in(fileType, supportFileTypes) {
		return ctx.JSONError(fmt.Sprintf("invalid type, only support: %s", strings.Join(supportFileTypes, ",")), http.StatusUnprocessableEntity)
	}

	fileID, err := g.updateImageDefinition(strings.TrimSpace(ctx.Input("id")), []byte(graphDef), fileType)
	if err != nil {
		return ctx.JSONError(err.Error(), http.StatusInternalServerError)
	}

	imagePreviewURL := fmt.Sprintf("/api/preview/%s.%s", fileID, fileType)

	resp := web.M{"preview": imagePreviewURL, "id": fileID}
	if fileType == "svg" {
		resp["preview-interact"] = fmt.Sprintf("/assets/index.html?url=%s", imagePreviewURL)
		resp["preview-sketch"] = fmt.Sprintf("/assets/index.html?url=%s&t=sketch&roughness=0", imagePreviewURL)
	}

	return ctx.JSON(resp)
}

func (g GraphvizController) buildSourcePath(name string) string {
	return filepath.Join(g.conf.TempDir, "sources", name)
}

func (g GraphvizController) buildTargetPath(name string) string {
	return filepath.Join(g.conf.TempDir, "graphviz", name)
}

func (g GraphvizController) createDefinitionFinger(def []byte, id string) string {
	if id != "" {
		return id
	}

	return fmt.Sprintf("%x", md5.Sum(def))
}

func (g GraphvizController) sourceExisted(name string) bool {
	return fileExist(g.buildSourcePath(name))
}

func (g GraphvizController) targetExisted(name string) bool {
	return fileExist(g.buildTargetPath(name))
}

func (g GraphvizController) updateImageDefinition(id string, definition []byte, filetype string) (string, error) {
	finger := g.createDefinitionFinger(definition, id)
	if err := ioutil.WriteFile(g.buildSourcePath(finger+".dot"), definition, os.ModePerm); err != nil {
		return "", fmt.Errorf("write temp file failed for source: %w", err)
	}

	return g.rebuildImage(finger, filetype)
}

func (g GraphvizController) buildImageAsStream(def []byte, filetype string) ([]byte, error) {
	finger := fmt.Sprintf("tmp-%s-%d", g.createDefinitionFinger(def, ""), time.Now().Unix())
	sourcePath := g.buildSourcePath(finger + ".dot")
	if err := ioutil.WriteFile(sourcePath, def, os.ModePerm); err != nil {
		return nil, err
	}
	defer os.Remove(sourcePath)

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	dotCmd := exec.CommandContext(ctx, g.conf.DotBin, "-T"+filetype, sourcePath)
	stdout, err := dotCmd.Output()
	if err != nil {
		return nil, err
	}

	return stdout, nil
}

func (g GraphvizController) rebuildImage(finger string, filetype string) (string, error) {
	if !g.sourceExisted(finger + ".dot") {
		return "", errors.New("sourcePath file not exist")
	}

	// 创建图片文件
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	sourcePath := g.buildSourcePath(finger + ".dot")
	targetPath := g.buildTargetPath(fmt.Sprintf("%s.%s", finger, filetype))

	dotCmd := exec.CommandContext(ctx, g.conf.DotBin, "-T"+filetype, "-o", targetPath, sourcePath)
	stdout, err := dotCmd.Output()
	if err != nil {
		_ = os.Remove(sourcePath)
		return "", err
	}

	if log.DebugEnabled() {
		log.Debugf("dot command output: %s", stdout)
	}

	return finger, nil
}

func (g GraphvizController) LoadImage(ctx web.Context) web.Response {
	id := ctx.PathVar("id")
	segs := strings.SplitN(id, ".", 2)
	if len(segs) != 2 {
		return ctx.JSONError("invalid file", http.StatusUnprocessableEntity)
	}

	finger, fileType := segs[0], segs[1]
	sourcePath := g.buildSourcePath(finger + ".dot")
	targetPath := g.buildTargetPath(id)

	if fileModTs(targetPath) < fileModTs(sourcePath) {
		_, err := g.rebuildImage(finger, fileType)
		if err != nil {
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

func fileModTs(path string) int64 {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}

	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return 0
	}

	return fi.ModTime().Unix()
}
