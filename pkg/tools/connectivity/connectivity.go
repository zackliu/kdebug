package connectivity

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Azure/kdebug/pkg/base"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

type ConnectivityTool struct {
	pid       string
	podName   string
	namespace string
	command   string
	image     string
}

const (
	DefaultCommand        = "wget --spider www.google.com"
	DefaultContainerImage = "busybox"
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
	if err := c.parseAndCheckParameters(ctx); err != nil {
		return err
	}

	if len(c.pid) > 0 {
		err := c.checkWithPid()
		if err != nil {
			return err
		}

		return nil
	}

	c.checkWithPod(ctx.KubeClient)

	return nil
}

func (c *ConnectivityTool) parseAndCheckParameters(ctx *base.ToolContext) error {
	if len(ctx.Connectivity.Pid) == 0 && len(ctx.Connectivity.PodName) == 0 {
		return fmt.Errorf("Either --connectivity.pid and --connectivity.pod should be set.")
	}
	if len(ctx.Connectivity.Pid) > 0 && len(ctx.Connectivity.PodName) > 0 {
		return fmt.Errorf("--connectivity.pid and --connectivity.pod can not be assigned together. Please set either of them.")
	}
	if len(ctx.Connectivity.PodName) > 0 {
		if ctx.KubeClient == nil {
			return fmt.Errorf("kubernetes client is not availble. Check kubeconfig.")
		}
	}

	c.pid = ctx.Connectivity.Pid
	c.podName = ctx.Connectivity.PodName
	if len(ctx.Connectivity.Command) > 0 {
		c.command = ctx.Connectivity.Command
	} else {
		c.command = DefaultCommand
	}

	if len(ctx.Connectivity.Image) > 0 {
		c.image = ctx.Connectivity.Image
	} else {
		c.image = DefaultContainerImage
	}

	return nil
}

func (c *ConnectivityTool) checkWithPid() error {
	_, err := logAndExec("nsenter", "-n", "-t", c.pid).Output()
	if err != nil {
		return err
	}

	args := strings.Fields(c.command)
	cmd := logAndExec(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (c *ConnectivityTool) checkWithPod(clientset *kubernetes.Clientset) error {
	kubectlCommand := fmt.Sprintf("kubectl debug -ti %s --image %s -- %s", c.podName, c.image, c.command)

	args := strings.Fields(kubectlCommand)
	cmd := logAndExec(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
