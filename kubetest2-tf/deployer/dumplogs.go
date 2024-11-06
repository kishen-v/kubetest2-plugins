package deployer

import (
	"bytes"
	"fmt"
	"github.com/ppc64le-cloud/kubetest2-plugins/pkg/providers/common"
	"os"
	"os/exec"
	"path/filepath"

	"k8s.io/klog/v2"
)

var commandFilename = map[string]string{
	fmt.Sprintf("%s.log", common.CommonProvider.Runtime): fmt.Sprintf("journalctl -xeu %s --no-pager", common.CommonProvider.Runtime),

	"dmesg.log":    "dmesg",
	"kernel.log":   "sudo journalctl --no-pager --output=short-precise -k",
	"services.log": "sudo systemctl list-units -t service --no-pager --no-legend --all"}

func (d *deployer) DumpClusterLogs() error {
	var stdErr bytes.Buffer
	klog.Infof("Collecting cluster logs under %s", d.logsDir)
	// create a directory based on the generated path: _rundir/dump-cluster-logs
	if _, err := os.Stat(d.logsDir); os.IsNotExist(err) {
		if err := os.Mkdir(d.logsDir, os.ModePerm); err != nil {
			klog.Errorf("cannot create a directory in path %q. Err: %v", d.logsDir, err)
			return err
		}
	} else if err == nil {
		klog.Errorf("%q already exists. Please clean up directory", d.logsDir)
		return err
	} else {
		return fmt.Errorf("an error occured while obtaining directory stats. Err: %v", err)
	}
	outfile, err := os.Create(filepath.Join(d.logsDir, "cluster-info.log"))
	if err != nil {
		klog.Errorf("Failed to create a log file. Err: %v", err)
		return err
	}
	defer outfile.Close()
	command := []string{
		"kubectl",
		"cluster-info",
		"dump",
	}
	klog.Infof("About to run: %s", command)
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdout = outfile
	cmd.Stderr = &stdErr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("couldn't use kubectl to dump cluster info: %v. StdErr: %s", err, stdErr.String())
	}
	// Todo: Include provider specific logic in this section. (Includes node level information/CRI/Services, etc.)
	for _, machineIP := range d.machineIPs {
		klog.Infof("Collecting node level information from machine %s", machineIP)
		for logFile, command := range commandFilename {
			outfile, err := os.Create(filepath.Join(d.logsDir, fmt.Sprintf("%s-%s.log", machineIP, logFile)))
			if err != nil {
				klog.Errorf("Failed to create a log file. Err: %v", err)
				return err
			}
			klog.V(1).Infof("Remotely executing command: %s", command)
			cmd := exec.Command("ssh", fmt.Sprintf("root@%s", machineIP), command)
			cmd.Stdout = outfile
			cmd.Stderr = &stdErr
			err = cmd.Run()
			if err != nil {
				klog.Errorf("An error occurred while obtaining logs from node: %v. StdErr: %s", err, stdErr.String())
				return err
			}
			outfile.Close()
		}
	}
	klog.Infof("Successfully collected cluster logs under %s", d.logsDir)
	return nil
}
