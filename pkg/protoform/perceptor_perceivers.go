/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
*/

package protoform

import (
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
)

func (i *Installer) perceiverReplicationController(imageName string, replicas int32) *components.ReplicationController {
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      imageName,
		Namespace: i.Config.Namespace,
	})
	rc.AddLabelSelectors(map[string]string{"name": imageName})

	return rc
}

// PodPerceiverReplicationController creates a replication controller for the pod perceiver
func (i *Installer) PodPerceiverReplicationController() (*components.ReplicationController, error) {
	name := i.Config.PodPerceiverImageName
	rc := i.perceiverReplicationController(name, 1)

	pod, err := i.perceiverPod(name, i.Config.ServiceAccounts["pod-perceiver"], "./pod-perceiver")
	if err != nil {
		return nil, fmt.Errorf("failed to create pod perceiver pod: %v", err)
	}
	rc.AddPod(pod)

	return rc, nil
}

// ImagePerceiverReplicationController creates a replication controller for the image perceiver
func (i *Installer) ImagePerceiverReplicationController() (*components.ReplicationController, error) {
	name := i.Config.ImagePerceiverImageName
	rc := i.perceiverReplicationController(name, 1)

	pod, err := i.perceiverPod(name, i.Config.ServiceAccounts["image-perceiver"], "./image-perceiver")
	if err != nil {
		return nil, fmt.Errorf("failed to create image perceiver pod: %v", err)
	}
	rc.AddPod(pod)

	return rc, nil
}

func (i *Installer) perceiverPod(imageName string, account string, cmd string) (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name:           imageName,
		ServiceAccount: account,
	})

	pod.AddLabels(map[string]string{"name": imageName})
	pod.AddContainer(i.perceiverContainer(imageName, cmd))

	vols, err := i.perceiverVolumes()

	if err != nil {
		return nil, err
	}

	for _, v := range vols {
		pod.AddVolume(v)
	}

	return pod, nil
}

func (i *Installer) perceiverContainer(imageName string, cmd string) *components.Container {
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:    imageName,
		Image:   fmt.Sprintf("%s/%s/%s:%s", i.Config.Registry, i.Config.ImagePath, imageName, i.Config.PerceiverImageVersion),
		Command: []string{cmd},
		Args:    []string{"/etc/perceiver/perceiver.yaml"},
		MinCPU:  i.Config.DefaultCPU,
		MinMem:  i.Config.DefaultMem,
	})

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: fmt.Sprintf("%d", i.Config.PerceiverPort),
		Protocol:      horizonapi.ProtocolTCP,
	})

	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "perceiver",
		MountPath: "/etc/perceiver",
	})
	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "logs",
		MountPath: "/tmp",
	})

	return container
}

func (i *Installer) perceiverVolumes() ([]*components.Volume, error) {
	vols := []*components.Volume{}

	vols = append(vols, components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "perceiver",
		MapOrSecretName: "perceiver",
	}))

	vol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "logs",
		Medium:     horizonapi.StorageMediumDefault,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create empty dir volume: %v", err)
	}
	vols = append(vols, vol)

	return vols, nil
}

func (i *Installer) perceiverService(imageName string) *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      imageName,
		Namespace: i.Config.Namespace,
	})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(i.Config.PerceiverPort),
		TargetPort: fmt.Sprintf("%d", i.Config.PerceiverPort),
		Protocol:   horizonapi.ProtocolTCP,
	})

	service.AddSelectors(map[string]string{"name": imageName})

	return service
}

// PodPerceiverService creates a service for the pod perceiver
func (i *Installer) PodPerceiverService() *components.Service {
	return i.perceiverService(i.Config.PodPerceiverImageName)
}

// ImagePerceiverService creates a service for the image perceiver
func (i *Installer) ImagePerceiverService() *components.Service {
	return i.perceiverService(i.Config.ImagePerceiverImageName)
}

// PerceiverConfigMap creates a config map for perceivers
func (i *Installer) PerceiverConfigMap() *components.ConfigMap {
	configMap := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      "perceiver",
		Namespace: i.Config.Namespace,
	})
	configMap.AddData(map[string]string{"perceiver.yaml": fmt.Sprint(`{"PerceptorHost": "`, i.Config.PerceptorImageName, `","PerceptorPort": "`, i.Config.PerceptorPort, `","AnnotationIntervalSeconds": "`, i.Config.AnnotationIntervalSeconds, `","DumpIntervalMinutes": "`, i.Config.DumpIntervalMinutes, `","Port": "`, i.Config.PerceiverPort, `","LogLevel": "`, i.Config.LogLevel, `"}`)})

	return configMap
}

func (i *Installer) perceiverServiceAccount(name string) *components.ServiceAccount {
	serviceAccount := components.NewServiceAccount(horizonapi.ServiceAccountConfig{
		Name:      name,
		Namespace: i.Config.Namespace,
	})

	return serviceAccount
}

// PodPerceiverServiceAccount creates a service account for the pod perceiver
func (i *Installer) PodPerceiverServiceAccount() *components.ServiceAccount {
	return i.perceiverServiceAccount(i.Config.ServiceAccounts["pod-perceiver"])
}

// ImagePerceiverServiceAccount creates a service account for the image perceiver
func (i *Installer) ImagePerceiverServiceAccount() *components.ServiceAccount {
	return i.perceiverServiceAccount(i.Config.ServiceAccounts["image-perceiver"])
}

// PodPerceiverClusterRole creates a cluster role for the pod perceiver
func (i *Installer) PodPerceiverClusterRole() *components.ClusterRole {
	clusterRole := components.NewClusterRole(horizonapi.ClusterRoleConfig{
		Name:       "pod-perceiver",
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		APIGroups: []string{"*"},
		Resources: []string{"pods"},
		Verbs:     []string{"get", "watch", "list", "update"},
	})

	return clusterRole
}

// ImagePerceiverClusterRole creates a cluster role for the image perceiver
func (i *Installer) ImagePerceiverClusterRole() *components.ClusterRole {
	clusterRole := components.NewClusterRole(horizonapi.ClusterRoleConfig{
		Name:       "image-perceiver",
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		APIGroups: []string{"*"},
		Resources: []string{"images"},
		Verbs:     []string{"get", "watch", "list", "update"},
	})

	return clusterRole
}

// PodPerceiverClusterRoleBinding creates a cluster role binding for the pod perceiver
func (i *Installer) PodPerceiverClusterRoleBinding(clusterRole *components.ClusterRole) *components.ClusterRoleBinding {
	clusterRoleBinding := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       "pod-perceiver",
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRoleBinding.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      i.Config.ServiceAccounts["pod-perceiver"],
		Namespace: i.Config.Namespace,
	})
	clusterRoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     clusterRole.GetName(),
	})

	return clusterRoleBinding
}

// ImagePerceiverClusterRoleBinding creates a cluster role binding for the image perceiver
func (i *Installer) ImagePerceiverClusterRoleBinding(clusterRole *components.ClusterRole) *components.ClusterRoleBinding {
	clusterRoleBinding := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       "image-perceiver",
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRoleBinding.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      i.Config.ServiceAccounts["image-perceiver"],
		Namespace: i.Config.Namespace,
	})
	clusterRoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     clusterRole.GetName(),
	})

	return clusterRoleBinding
}
