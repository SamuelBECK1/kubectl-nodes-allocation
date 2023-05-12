package cmd
import (
	"errors"
	"fmt"
	"math"
	"github.com/spf13/cobra"
	"io"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
        "k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/clientcmd"
        corev1 "k8s.io/api/core/v1"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type nodeAllocationCmd struct {
	out io.Writer
}

func NewNodeAllocationCommand(streams genericclioptions.IOStreams) *cobra.Command {
	helloWorldCmd := &nodeAllocationCmd{
		out: streams.Out,
	}

	cmd := &cobra.Command{
		Use:          "node-allocation",
		Short:        "list nodes with their commitments",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.New("this command does not accept arguments")
			}
			return helloWorldCmd.run()
		},
	}
	return cmd
}

func (sv *nodeAllocationCmd) run() error {
	err := getNodeAllocation(sv.out)
	if err != nil {
		return err
	}

	return nil
}

func getNodeAllocation(w io.Writer) error {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err := kubeConfig.ClientConfig()
	_=err
	clientset, err := kubernetes.NewForConfig(config)
        _=err
	nodes,_:= clientset.CoreV1().Nodes().List(metav1.ListOptions{})
        fmt.Fprintf(w,"NODE\tMEM REQUEST\tCPU REQUEST\tNODE MEM CAPACITY\tNODE CPU CAPACITY\tPERCENT MEM REQUESTED\tPERCENT CPU REQUESTED\n")
        for _,n := range nodes.Items {
                nodres,_ := genResource(n.Status.Capacity)
        	pl,_:=getNonTerminatedPodList(clientset,n.Name)
		res,_:=computeResource(pl)
	        fmt.Fprintf(w,
                    "%s\t%d\t%d\t%d\t%d\t%d%%\t%d%%\n",
                    n.Name,
                    (int64)(math.Ceil(res["memory"]/1000/1024)),
                    (int64)(res["cpu"]),
                    (int64)(math.Ceil(nodres["memory"]/1000/1024)),
                    (int64)(nodres["cpu"]),
                    (int64)(math.Ceil(100*res["memory"]/nodres["memory"])),
                    (int64)(math.Ceil(100*res["cpu"]/nodres["cpu"])),
                )
	}
	return nil
}

func computeResource(podList []corev1.Pod) (map[string]float64, error){
	var memory float64 = 0
        var cpu float64 = 0
	for _, p  := range podList {
		for _,c := range p.Spec.Containers {
			rm,_ := genResource(c.Resources.Requests)
			memory = memory + rm["memory"]
			cpu = cpu + rm["cpu"]
 		}
	}
	result := make(map[string]float64)
        result["memory"] = memory
        result["cpu"] = cpu
        return result, nil
}

func genResource(resourceList corev1.ResourceList) (map[string]float64, error){
	result := make(map[string]float64)
	for name, quantity := range resourceList {
		result[name.String()] = float64(quantity.MilliValue())
	}
	return result, nil
}

func getNonTerminatedPodList(coreClient kubernetes.Interface, node string) ([]corev1.Pod, error) {
    fieldSelector, _ := fields.ParseSelector("spec.nodeName=" + node + ",status.phase!=" + string(corev1.PodSucceeded) + ",status.phase!=" + string(corev1.PodFailed))
    podListAllNs,_ := coreClient.CoreV1().Pods("").List(metav1.ListOptions{FieldSelector: fieldSelector.String()})
    return podListAllNs.Items, nil
}
