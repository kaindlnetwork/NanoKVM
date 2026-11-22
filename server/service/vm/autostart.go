package vm

import (
	"NanoKVM-Server/proto"
	"NanoKVM-Server/utils"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var autostartDirectory = "/etc/kvm/autostart"

func (s *Service) GetAutostart(c *gin.Context) {
	var rsp proto.Response

	var files []string
	err := filepath.Walk(autostartDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			files = append(files, info.Name())
		}

		return nil
	})
	if err != nil {
		rsp.ErrRsp(c, -1, "get autostart directory fail")
		return
	}
	rsp.OkRspWithData(c, &proto.GetAutostartRsp{
		Files: files,
	})

	log.Debugf("get autostart total %d", len(files))
}

func (s *Service) UploadAutostart(c *gin.Context) {
	var req proto.UploadAutostartReq
	var rsp proto.Response

	fileName := c.Param("name")

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "parse form request fail")
		return
	}

	if _, err := os.Stat(autostartDirectory); err != nil {
		_ = os.MkdirAll(autostartDirectory, 0o700)
	}

	// Autostart entries are executed at boot, so they keep the exec bit.
	if err := utils.ConfinedWriteFile(autostartDirectory, fileName, []byte(req.Content), 0o700); err != nil {
		if err == utils.ErrUnsafeFileName {
			rsp.ErrRsp(c, -1, "invalid file name")
			return
		}
		rsp.ErrRsp(c, -1, "write content fail")
		return
	}

	rsp.OkRspWithData(c, fileName)

	log.Debugf("upload autostart %s success", fileName)
}

func (s *Service) DeleteAutostart(c *gin.Context) {
	var rsp proto.Response

	fileName := c.Param("name")

	if err := utils.ConfinedRemove(autostartDirectory, fileName); err != nil {
		log.Errorf("delete autostart file %s fail", fileName)
		rsp.ErrRsp(c, -3, "remove file fail")
		return
	}

	rsp.OkRsp(c)
	log.Debugf("delete autostart %s success", fileName)
}

func (s *Service) GetAutostartContent(c *gin.Context) {
	var rsp proto.Response
	fileName := c.Param("name")
	content, err := utils.ConfinedReadFile(autostartDirectory, fileName)
	if err != nil {
		rsp.ErrRsp(c, -1, "read file fail")
		return
	}

	rsp.OkRspWithData(c, string(content))
	log.Debugf("get autostart content %s success", fileName)
}
