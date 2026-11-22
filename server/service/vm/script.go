package vm

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/utils"
)

// maxScriptSize caps uploaded scripts; they are read into memory.
const maxScriptSize = 1 << 20

// ScriptDirectory is a variable so tests can point it at a temp dir.
var ScriptDirectory = "/etc/kvm/scripts"

func (s *Service) GetScripts(c *gin.Context) {
	var rsp proto.Response

	var files []string
	err := filepath.Walk(ScriptDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && isScript(info.Name()) {
			files = append(files, info.Name())
		}

		return nil
	})
	if err != nil {
		rsp.ErrRsp(c, -1, "get scripts failed")
		return
	}

	rsp.OkRspWithData(c, &proto.GetScriptsRsp{
		Files: files,
	})

	log.Debugf("get scripts total %d", len(files))
}

func (s *Service) UploadScript(c *gin.Context) {
	var rsp proto.Response

	_, header, err := c.Request.FormFile("file")
	if err != nil {
		rsp.ErrRsp(c, -1, "bad request")
		return
	}

	if !isScript(header.Filename) {
		rsp.ErrRsp(c, -2, "invalid arguments")
		return
	}

	if _, err = utils.SafeFileName(header.Filename); err != nil {
		rsp.ErrRsp(c, -2, "invalid file name")
		return
	}

	if _, err = os.Stat(ScriptDirectory); err != nil {
		_ = os.MkdirAll(ScriptDirectory, 0o700)
	}

	// Written through os.Root: an existing symlink with this name cannot
	// redirect the upload outside the script directory.
	src, err := header.Open()
	if err != nil {
		rsp.ErrRsp(c, -2, "save failed")
		return
	}
	content, err := io.ReadAll(io.LimitReader(src, maxScriptSize+1))
	_ = src.Close()
	if err != nil || len(content) > maxScriptSize {
		rsp.ErrRsp(c, -2, "save failed")
		return
	}
	if err = utils.ConfinedWriteFile(ScriptDirectory, header.Filename, content, 0o700); err != nil {
		rsp.ErrRsp(c, -2, "save failed")
		return
	}

	data := &proto.UploadScriptRsp{
		File: header.Filename,
	}
	rsp.OkRspWithData(c, data)

	log.Debugf("upload script %s success", header.Filename)
}

func (s *Service) RunScript(c *gin.Context) {
	var req proto.RunScriptReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	if !isScript(req.Name) {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}
	// Only a regular file directly in ScriptDirectory runs; a symlink could
	// point anywhere once the path leaves os.Root for exec.
	if err := utils.ConfinedRegularFile(ScriptDirectory, req.Name); err != nil {
		rsp.ErrRsp(c, -1, "invalid file name")
		return
	}
	scriptPath := filepath.Join(ScriptDirectory, req.Name)

	// Execute the file directly instead of through "sh -c": the name never
	// reaches a shell, so it cannot inject commands.
	//
	// Running an uploaded script is the purpose of this endpoint (it needs
	// the terminal grant). scriptPath is one validated file name inside
	// ScriptDirectory, so gosec's G204 does not apply.
	var cmd *exec.Cmd
	if strings.HasSuffix(strings.ToLower(req.Name), ".py") {
		cmd = exec.Command("python", scriptPath) // #nosec G204 -- validated path, see above
	} else if hasShebang(req.Name) {
		cmd = exec.Command(scriptPath) // #nosec G204 -- validated path, see above
	} else {
		// Same fallback the shell applied before: scripts without "#!" run in sh.
		cmd = exec.Command("sh", scriptPath) // #nosec G204 -- validated path, see above
	}

	var output []byte
	var err error

	if req.Type == "foreground" {
		output, err = cmd.CombinedOutput()
	} else {
		cmd.Stdout = nil
		cmd.Stderr = nil
		go func() {
			err := cmd.Run()
			if err != nil {
				log.Errorf("run script %s in background failed: %s", req.Name, err)
			}
		}()
	}

	if err != nil {
		log.Errorf("run script %s failed: %s", req.Name, err.Error())
		rsp.ErrRsp(c, -2, "run script failed")
		return
	}

	rsp.OkRspWithData(c, &proto.RunScriptRsp{
		Log: string(output),
	})

	log.Debugf("run script %s success", req.Name)
}

func (s *Service) DeleteScript(c *gin.Context) {
	var req proto.DeleteScriptReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	if !isScript(req.Name) {
		rsp.ErrRsp(c, -2, "invalid file name")
		return
	}

	if err := utils.ConfinedRemove(ScriptDirectory, req.Name); err != nil {
		log.Errorf("delete script %s failed: %s", req.Name, err)
		rsp.ErrRsp(c, -3, "delete failed")
		return
	}

	rsp.OkRsp(c)
	log.Debugf("delete script %s success", req.Name)
}

func isScript(name string) bool {
	nameLower := strings.ToLower(name)
	if strings.HasSuffix(nameLower, ".sh") || strings.HasSuffix(nameLower, ".py") {
		return true
	}

	return false
}

func hasShebang(name string) bool {
	data, err := utils.ConfinedReadFile(ScriptDirectory, name)
	return err == nil && len(data) >= 2 && data[0] == '#' && data[1] == '!'
}
