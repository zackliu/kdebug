package base

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"

	"github.com/Azure/kdebug/pkg/env"
)

type CheckContext struct {
	// TODO: Add user input here
	Pod struct {
		Name      string
		Namespace string
	}

	// TODO: Add shared dependencies here, for example, kube-client
	Environment env.Environment
	KubeClient  *kubernetes.Clientset
}

type ToolContext struct {
	Tcpdump          Tcpdump
	VmRebootDetector VMRebootDetector
	Netexec          Netexec
	KubeConfigFlag   *genericclioptions.ConfigFlags
}

type VMRebootDetector struct {
	CheckDays int `short:"d" long:"checkdays" description:"Days you want to look back to search for reboot events. Default is 1."`
}

type Tcpdump struct {
	Source      string `long:"source" description:"The source of the connection. Format: <ip>:<port>. Watch all sources if not assigned."`
	Destination string `long:"destination" description:"The destination of the connection. Format: <ip>:<port>. Watch all destination if not assigned."`
	Host        string `long:"host" description:"The host(either src or dst) of the connection. Format: <ip>:<port>. Watch if not assigned."`
	Pid         string `short:"p" long:"pid" description:"Attach into a specific pid's network namespace. Use current namespace if not assigned"`
	TcpOnly     bool   `long:"tcponly" description:"Only watch tcp connections"`
}

type Netexec struct {
	Pid       string `long:"pid" description:"Attach into a specific pid's network namespace."`
	PodName   string `long:"pod" description:"Attach into a specific pod's network namespace. Caution: The command will use ephemeral debug container to attach a container with 'ghcr.io/azure/kdebug:main' to the target pod."`
	Namespace string `long:"namespace" description:"the namespace of the pod."`
	Command   string `long:"command" description:"Customize the command to be run in container namespace. Leave it blank to run default check."`
	Image     string `long:"image" description:"Customize the image to be used to run command when using --netexec.pod. Leave it blank to use busybox."`
}

type CheckResult struct {
	Checker         string
	Error           string
	Description     string
	Recommendations []string
	Logs            []string
	HelpLinks       []string
}

func (r *CheckResult) Ok() bool {
	return r.Error == ""
}
