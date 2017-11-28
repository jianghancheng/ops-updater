package cron

import (
	"fmt"
	"github.com/Cepave/ops-common/model"
	"github.com/Cepave/ops-common/utils"
	"github.com/toolkits/file"
	//	"io/ioutil"
	"log"
	"os/exec"
	"path"
	"strings"
	"time"
)

func StartDesiredAgent(da *model.DesiredAgent) {
	if err := InsureDesiredAgentDirExists(da); err != nil {
		log.Println("InsureDesiredAgentDirExists wrong")
		return
	}

	if err := InsureNewVersionFiles(da); err != nil {
		log.Println("InsureNewVersionFiles wrong")
		return
	}

	if err := Untar(da); err != nil {
		log.Println("Untar error")
		return
	}

	if err := StopAgentOf(da.AgentDir, da.Version); err != nil {
		log.Println("StopAgentOf wrong")
		return
	}

	if err := ControlStartIn(da.AgentVersionDir); err != nil {
		log.Println("ControlStartIn wrong")
		return
	}
	file.WriteString(path.Join(da.AgentDir, ".version"), da.Version)
}

func Untar(da *model.DesiredAgent) error {
	cmd := exec.Command("tar", "zxf", da.TarballFilename)
	cmd.Dir = da.AgentVersionDir
	err := cmd.Run()
	if err != nil {
		log.Println("tar zxf", da.TarballFilename, "fail", err)
		return err
	}

	return nil
}

func ControlStartIn(workdir string) error {
	out, err := ControlStatus(workdir)

	if err == nil && strings.Contains(out, "started") {

		return nil
	}

	_, err = ControlStart(workdir)
	if err != nil {
		return err
	}

	time.Sleep(time.Second * 3)

	out, err = ControlStatus(workdir)
	if err == nil && strings.Contains(out, "started") {
		return nil
	}

	log.Println("agent does not start when we use control start in func")
	return err
}

func InsureNewVersionFiles(da *model.DesiredAgent) error {
	if FilesReady(da) {
		return nil
	}
	//自己搭建nginx web服务，把agent 的tar包和MD5包放在10.161.161.204上面，使用wget下载tar,md5文件到每台服务器
	TarballFilename := da.TarballFilename
	TarballUrl := fmt.Sprintf("http://10.161.161.204:3001/%s", TarballFilename)
	downloadTarballCmd := exec.Command("wget", "--no-check-certificate", "--auth-no-challenge", TarballUrl, "-O", da.TarballFilename)
	downloadTarballCmd.Dir = da.AgentVersionDir
	err := downloadTarballCmd.Run()
	if err != nil {
		log.Println("wget tarball fail %s", err)
		return err
	}
	Md5Filename := da.Md5Filename
	Md5Url := fmt.Sprintf("http://10.161.161.204:3001/%s", Md5Filename)
	downloadMd5Cmd := exec.Command("wget", "--no-check-certificate", "--auth-no-challenge", Md5Url, "-O", da.Md5Filename)
	downloadMd5Cmd.Dir = da.AgentVersionDir
	err = downloadMd5Cmd.Run()
	if err != nil {
		log.Println("wget tar.md5 fail %s", err)
		return err
	}

	if utils.Md5sumCheck(da.AgentVersionDir, da.Md5Filename) {
		return nil
	} else {
		return fmt.Errorf("md5sum -c fail")
	}
}

func FilesReady(da *model.DesiredAgent) bool {
	if !file.IsExist(da.Md5Filepath) {
		return false
	}

	if !file.IsExist(da.TarballFilepath) {
		return false
	}

	if !file.IsExist(da.ControlFilepath) {
		return false
	}

	return utils.Md5sumCheck(da.AgentVersionDir, da.Md5Filename)
}

func InsureDesiredAgentDirExists(da *model.DesiredAgent) error {
	err := file.InsureDir(da.AgentDir)
	if err != nil {
		log.Println("insure dir", da.AgentDir, "fail", err)
		return err
	}

	err = file.InsureDir(da.AgentVersionDir)
	if err != nil {
		log.Println("insure dir", da.AgentVersionDir, "fail", err)
	}
	return err
}
