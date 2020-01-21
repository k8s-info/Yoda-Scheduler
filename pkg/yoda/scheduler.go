package yoda

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

const (
	Name  = "yoda"
)

var _ framework.FilterPlugin = &Yoda{}

type Args struct {
	KubeConfig string `json:"kubeconfig,omitempty"`
	Master     string `json:"master,omitempty"`
}

type Yoda struct {
	args   *Args
	handle framework.FrameworkHandle
}

func (y *Yoda) Name() string {
	return Name
}

func (y *Yoda) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, node *nodeinfo.NodeInfo) *framework.Status {
	klog.V(3).Infof("filter pod: %v, node: %v", pod.Name, node.Node().Name)
	if NodeHasGPU(node){
		if NodeGPUHealth(node){
			if PodNeedLevel(pod){
				if !PodFitsLevel(pod,node){
					return framework.NewStatus(framework.Unschedulable, "Node:"+node.Node().Name+" GPU Low Level")
				}
			}
			if PodNeedMemory(pod){
				if !PodFitsMemory(pod,node){
					return framework.NewStatus(framework.Unschedulable, "Node:"+node.Node().Name+" GPUNoMemory")
				}
			}
			return framework.NewStatus(framework.Success, "")
		}
		return framework.NewStatus(framework.Unschedulable, "Node:"+node.Node().Name+" GPU Unhealthy")
	}
	return framework.NewStatus(framework.Unschedulable, "Node:"+node.Node().Name+" No GPU")
}

func New(configuration *runtime.Unknown, f framework.FrameworkHandle) (framework.Plugin, error) {
	args := &Args{}
	if err := framework.DecodeInto(configuration, args); err != nil {
		return nil, err
	}
	klog.V(3).Infof("get plugin config args: %+v", args)
	return &Yoda{
		args: args,
		handle: f,
	}, nil
}