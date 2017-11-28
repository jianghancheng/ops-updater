package cron

import (
	"fmt"
	"github.com/Cepave/ops-common/model"
	"github.com/Cepave/ops-updater/g"
	f "github.com/toolkits/file"
	"log"
	"os/exec"
	"path"
	"strings"
	"time"
)

func BuildHeartbeatRequest(hostname string, agentDirs []string) model.HeartbeatRequest {
	//增加hostname 和ip同时给服务端
	host_and_ip := hostname + "||" + g.Config().Ip
	req := model.HeartbeatRequest{Hostname: host_and_ip}
	realAgents := []*model.RealAgent{}
	now := time.Now().Unix()
	for _, agentDir := range agentDirs {
		// 如果目录下没有.version，我们认为这根本不是一个agent
		//改变了agent的存放目录
		versionFile := path.Join(g.SelfDir, "Agents", agentDir, ".version")
		if !f.IsExist(versionFile) {
			log.Println(".version does not exist")
			continue
		}

		version, err := f.ToTrimString(versionFile)
		log.Printf("version is %s ", version)
		if err != nil {
			log.Printf("read %s/.version fail: %v", agentDir, err)
			continue
		}

		controlFile := path.Join(g.SelfDir, "Agents", agentDir, version, "control")
		if !f.IsExist(controlFile) {
			log.Printf("%s is nonexistent", controlFile)
			continue
		}
		cmd := exec.Command("./control", "status")
		cmd.Dir = path.Join(g.SelfDir, "Agents", agentDir, version)
		bs, err := cmd.CombinedOutput()
		log.Printf("agent status %s", strings.TrimSpace(string(bs)))
		status := ""
		if err != nil {
			status = fmt.Sprintf("exec `./control status` fail: %s", err)
			log.Println("error %s", status)
		} else {
			status = strings.TrimSpace(string(bs))

			// if strings.Contains(status,"UP")&& strings.Contains(status,"agent"){
			//     status="started"
			// }else{
			//     status="stoped"
			// }
		}

		realAgent := &model.RealAgent{
			Name:      agentDir,
			Version:   version,
			Status:    status,
			Timestamp: now,
		}

		realAgents = append(realAgents, realAgent)
	}

	req.RealAgents = realAgents
	return req
}

func ListAgentDirs() ([]string, error) {
	//由于agent目录改变，所以做相应调整
	agentDirs, err := f.DirsUnder(path.Join(g.SelfDir, "Agents"))
	if err != nil {
		log.Printf("list dirs under", g.SelfDir, "fail", err)
	}
	return agentDirs, err
}
