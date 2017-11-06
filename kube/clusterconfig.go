/*
Copyright 2017 Kinvolk GmbH

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kube

import (
	"fmt"

	"k8s.io/client-go/rest"
	// "k8s.io/client-go/tools/clientcmd"
)

const DefaultCgroupRoot string = "/sys/fs/cgroup/unified/cgnet"

func BuildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		// TODO: does it make sense to allow this?
		// If host of the out-of-cluster cgnet process is a cluster-node we
		// could install bpf programs there but not on other nodes
		//
		// return clientcmd.BuildConfigFromFlags("", kubeconfig)
		return nil, fmt.Errorf("running out-of-cluster is not supported.")
	}
	return rest.InClusterConfig()
}

func GetCgroupRoot(_ *rest.Config) (string, error) {
	return DefaultCgroupRoot, nil
}
