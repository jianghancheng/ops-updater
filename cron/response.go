package cron

import (
	"github.com/Cepave/ops-common/model"
	"github.com/Cepave/ops-updater/g"
	"log"
	"path"
	"strings"
	"github.com/toolkits/file"
)

func HandleHeartbeatResponse(respone *model.HeartbeatResponse) {
	if respone.ErrorMessage != "" {
		log.Println("receive error message:", respone.ErrorMessage)
		return
	}

	das := respone.DesiredAgents
	if das == nil || len(das) == 0 {
		return
	}

	for _, da := range das {
		//workdir目录发生改变，做出相应修改
		workdir := path.Join(g.SelfDir, "Agents")
		da.FillAttrs(workdir)

		if g.Config().DesiredAgent == "" || g.Config().DesiredAgent == da.Name {
			log.Println("in response fun")
			HandleDesiredAgent(da)
		}
	}
}

func HandleDesiredAgent(da *model.DesiredAgent) {
	if da.Is_execute == "false" {
		log.Println("Is_excute is false")
		return
	}

	//获取当前版本和状态和期待版本和状态
	versionFile := path.Join(da.AgentDir, ".version")
	if !file.IsExist(versionFile) {
		log.Printf("WARN: %s is nonexistent", versionFile)
		return
	}
	version, err := file.ToTrimString(versionFile)
	out, err := ControlStatus(da.AgentVersionDir)
	status := ""

	if err == nil && strings.Contains(out, "started") {
		status = "start"
	} else {
		status = "stop"
	}
	//当版本和状态一样则不做任何动作
	if status == da.Cmd && version == da.Version {
		log.Println("Same Version and Cmd")
		log.Printf("Version:%s  Desired Version:%s Status:%s Desired Status:%s", version, da.Version, status, da.Cmd)
		return
	}
	//有一种可能：如果有人误把正在运行的agent给删除了怎么办？？？？解决：1.错误log(手动kill进程，拷贝当前运行agent文件夹并开启agent)
    //如果业务方需要自己安装agent,则把agent放到指定目录手动开启，并生成.version文件写入启动版本

	//版本一样期待状态是start时直接开启，期待状态stop使用下面StopDesiredAgent()
	if version == da.Version && da.Cmd == "start" {
		log.Println("Same version and agent will start")
		if err := ControlStartIn(da.AgentVersionDir); err != nil {
			return
		}
		log.Println("Same version and agent start already")
	}
	//更新新版本时，
	if da.Cmd == "start" {
		StartDesiredAgent(da)
	} else if da.Cmd == "stop" {
		log.Println("in stop process")
		StopDesiredAgent(da)
	} else {
		log.Println("unknown cmd", da)
	}
}
