package oom

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Azure/kdebug/pkg/base"
)

const (
	logPath         = "/var/log/kern.log"
	cgroupOOMKeyStr = "Memory cgroup out of memory"
	outOfMemoryKey  = "Out of memory"
)

var helpLink = []string{
	"https://www.kernel.org/doc/gorman/html/understand/understand016.html",
	"https://stackoverflow.com/questions/18845857/what-does-anon-rss-and-total-vm-mean",
	"https://medium.com/tailwinds-navigator/kubernetes-tip-how-does-oomkilled-work-ba71b135993b",
}

var oomRegex = regexp.MustCompile("^(.*:.{2}:.{2}) .* process (.*) \\((.*)\\) .* anon-rss:(.*), file-rss.* oom_score_adj:(.*)")

type OOMChecker struct {
	kernLogPath string
}

func (c *OOMChecker) Name() string {
	return "OOM"
}

func New() *OOMChecker {
	//todo: support other logpath
	return &OOMChecker{
		kernLogPath: logPath,
	}
}

func (c *OOMChecker) Check(ctx *base.CheckContext) ([]*base.CheckResult, error) {
	var results []*base.CheckResult
	oomResult, err := c.checkOOM(ctx)
	if err != nil {
		return nil, err
	}
	results = append(results, oomResult)
	return results, nil
}

func (c *OOMChecker) checkOOM(ctx *base.CheckContext) (*base.CheckResult, error) {
	result := &base.CheckResult{
		Checker: c.Name(),
	}
	//todo:support other os
	if !ctx.Environment.HasFlag("linux") {
		result.Description = fmt.Sprint("Skip oom check in non-linux os")
		return result, nil
	}
	oomInfos, err := c.getAndParseOOMLog()
	if err != nil {
		return nil, err
	} else if len(oomInfos) > 0 {
		result.Error = strings.Join(oomInfos, "\n")
		result.Description = "Detect process oom killed"
		result.HelpLinks = helpLink
	} else {
		result.Description = "No OOM found in recent kernlog."
	}
	return result, nil
}
func (c *OOMChecker) getAndParseOOMLog() ([]string, error) {
	file, err := os.Open(c.kernLogPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var oomInfos []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		tmp := scanner.Text()
		//todo: more sophisticated OOM context
		//pattern match. https://github.com/torvalds/linux/blob/551acdc3c3d2b6bc97f11e31dcf960bc36343bfc/mm/oom_kill.c#L1120, https://github.com/torvalds/linux/blob/551acdc3c3d2b6bc97f11e31dcf960bc36343bfc/mm/oom_kill.c#L895
		if strings.Contains(tmp, cgroupOOMKeyStr) || strings.Contains(tmp, outOfMemoryKey) {
			oomInfo, err := parseOOMContent(tmp)
			if err != nil {
				return nil, err
			} else {
				oomInfos = append(oomInfos, oomInfo)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return oomInfos, nil
}

func parseOOMContent(content string) (string, error) {
	match := oomRegex.FindStringSubmatch(content)
	if len(match) != 6 {
		err := fmt.Errorf("Can't parse oom content:%s \n", content)
		return "", err
	} else {
		return fmt.Sprintf("progress:[%s %s] is OOM kill at time [%s]. [rss:%s] [oom_score_adj:%s]\n", match[2], match[3], match[1], match[4], match[5]), nil
	}
}
