package ansible

import (
	"context"
	"fmt"
	"path/filepath"

	"k8s.io/klog/v2"

	"github.com/apenella/go-ansible/v2/pkg/execute"
	ansiblePlaybook "github.com/apenella/go-ansible/v2/pkg/playbook"

	"github.com/ppc64le-cloud/kubetest2-plugins/data"
)

const (
	ansibleDataDir = "k8s-ansible"
)

func Playbook(dir, inventory, playbook string, extraVars map[string]string) error {
	if err := unpackAnsible(dir); err != nil {
		return fmt.Errorf("failed to unpack the ansible code: %v", err)
	}
	extraVarsMap := make(map[string]interface{}, len(extraVars))
	for key, val := range extraVars {
		extraVarsMap[key] = val
	}
	klog.Infof("Running ansible playbook: %s", playbook)
	playbookCmd := ansiblePlaybook.NewAnsiblePlaybookCmd(
		ansiblePlaybook.WithPlaybooks(filepath.Join(dir, playbook)),
		ansiblePlaybook.WithPlaybookOptions(
			&ansiblePlaybook.AnsiblePlaybookOptions{
				ExtraVars: extraVarsMap,
				Inventory: inventory,
			}),
	)
	return execute.NewDefaultExecute(
		execute.WithCmd(playbookCmd),
		execute.WithErrorEnrich(ansiblePlaybook.NewAnsiblePlaybookErrorEnrich()),
	).Execute(context.Background())
}

func unpackAnsible(dir string) error {
	return data.Unpack(dir, ansibleDataDir)
}
