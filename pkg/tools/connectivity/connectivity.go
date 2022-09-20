package connectivity

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Azure/kdebug/pkg/base"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/homedir"
	"k8s.io/kubectl/pkg/cmd/debug"
	"k8s.io/kubectl/pkg/cmd/util"
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

	err := c.checkWithPod(ctx.KubeClient)
	if err != nil {
		return err
	}

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
	// kubectlCommand := fmt.Sprintf("kubectl debug -ti %s --image %s -- %s", c.podName, c.image, c.command)

	// args := strings.Fields(kubectlCommand)
	// cmd := logAndExec(args[0], args[1:]...)
	// cmd.Stdin = os.Stdin
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	// if err := cmd.Run(); err != nil {
	// 	return err
	// }

	cf := genericclioptions.NewConfigFlags(false)
	if home := homedir.HomeDir(); home != "" {
		kubeConfigPath := filepath.Join(home, ".kube", "config")
		cf.KubeConfig = &kubeConfigPath
	}

	// bcmd := cmd.NewDefaultKubectlCommandWithArgs(cmd.KubectlOptions{
	// 	PluginHandler: cmd.NewDefaultPluginHandler(plugin.ValidPluginFilenamePrefixes),
	// 	Arguments:     strings.Fields("kubectl debug -ti signalr-00000000-0000-0000-0000-000000000002-fd78779f9-nr9zz --image busybox -- wget --spider www.google.com"),
	// 	ConfigFlags:   cf,
	// 	IOStreams:     genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
	// })
	// bcmd.Run()

	fac := util.NewFactory(cf)
	stream := genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
	opt := debug.NewDebugOptions(stream)
	opt.Args = strings.Fields(DefaultCommand)
	opt.Interactive = true
	opt.TTY = true
	opt.Image = DefaultContainerImage
	opt.Attach = true
	ns, _, err := fac.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}
	opt.Namespace = ns
	opt.TargetNames = []string{c.podName}

	pcmd := &cobra.Command{
		Use: "kubectl",
	}
	acmd := &cobra.Command{
		Use: "debug (POD | TYPE[[.VERSION].GROUP]/NAME) [ -- COMMAND [args...] ]",
	}
	pcmd.AddCommand(acmd)
	print(acmd.Parent().CommandPath())

	err = opt.Run(fac, acmd)
	if err != nil {
		print(err)
		return err
	}
	return nil
}
