package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/funkygao/go-helix"
	"github.com/funkygao/go-helix/store/zk"
	"github.com/funkygao/gocli"
)

type Init struct {
	Ui  cli.Ui
	Cmd string
}

func (this *Init) Run(args []string) (exitCode int) {
	var (
		node   string
		reinit bool
	)
	cmdFlags := flag.NewFlagSet("init", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&node, "addNode", "", "")
	cmdFlags.BoolVar(&reinit, "reinit", false, "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	// create the admin instance and connect
	admin := zk.NewZkHelixAdmin(zkSvr)
	must(admin.Connect())

	if reinit {
		must(admin.DropCluster(cluster))
	}

	if node != "" {
		if ok, err := admin.IsClusterSetup(cluster); !ok || err != nil {
			this.Ui.Errorf("cluster %s not setup", cluster)
			return 1
		}

		must(admin.AddNode(cluster, node))
		this.Ui.Infof("node %s added to cluster %s", node, cluster)
		return
	}

	// create the cluster
	must(admin.AddCluster(cluster))
	this.Ui.Infof("cluster[%s] added", cluster)

	must(admin.AllowParticipantAutoJoin(cluster, true))
	this.Ui.Output("enable partition auto join")

	// define the resource and partition
	resourceOption := helix.DefaultAddResourceOption(partitions, stateModel)
	resourceOption.RebalancerMode = helix.RebalancerModeFullAuto
	must(admin.AddResource(cluster, resource, resourceOption))
	this.Ui.Outputf("resource[%s] partitions:%d model:%s added to cluster[%s]", resource,
		partitions, stateModel, cluster)

	this.Ui.Warnf("%s/bin/run-helix-controller.sh --zkSvr %s --cluster %s", helixInstallBase, zkSvr, cluster)

	return
}

func (*Init) Synopsis() string {
	return "Initialize redis cluster"
}

func (this *Init) Help() string {
	help := fmt.Sprintf(`
Usage: %s init [options]

    %s

Options:

    -addNode host_port

    -reinit

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
