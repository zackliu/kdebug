package connectivity

import (
	"os"
	"os/exec"
	"strings"

	"github.com/Azure/kdebug/pkg/base"
	log "github.com/sirupsen/logrus"
)

type ConnectivityTool struct {
	pid     string
	podName string
	command string
}

const (
	DefaultCommand = "wget -q --spider www.google.com"
)

func New() *ConnectivityTool {
	return &ConnectivityTool{}
}

func (c *ConnectivityTool) Name() string {
	return "Connectivity"
}

func logAndExec(name string, args ...string) *exec.Cmd {
	log.Infof("Exec %s %+v", name, args)
	return exec.Command(name, args...)
}

func (c *ConnectivityTool) Run(ctx *base.ToolContext) error {
	c.ParseParameters(ctx)

	if len(c.pid) > 0 {
		c.checkWithPid()
	} else if len(c.podName) > 0 {
		c.checkWithPod()
	} else {
		return "error"
	}

	cmd := logAndExec("tcpdump", strings.Split(tcpdumpArgs, " ")...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

func (c *ConnectivityTool) ParseParameters(ctx *base.ToolContext) {
	c.pid = ctx.Connectivity.Pid
	c.podName = ctx.Connectivity.PodName
	c.command = ctx.Connectivity.Command
}

func (c *ConnectivityTool) checkWithPid() error {
	_, err := logAndExec("nsenter", "-n", "-t", c.pid).Output()

	if err != nil {
		return err
	}

	cmd := logAndExec(DefaultCommand)
	if err := cmd.Run(); err != nil {
		return err
	}

}

func (c *ConnectivityTool) checkWithPod() {
}
