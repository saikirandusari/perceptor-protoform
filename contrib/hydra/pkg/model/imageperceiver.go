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

package model

import (
	"encoding/json"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ImagePerceiverConfigMap struct {
	PerceptorHost             string
	PerceptorPort             int
	AnnotationIntervalSeconds int
	DumpIntervalMinutes       int
}

type ImagePerceiver struct {
	PodName string
	Image   string
	Port    int32
	CPU     resource.Quantity
	Memory  resource.Quantity

	ConfigMapName  string
	ConfigMapMount string
	ConfigMapPath  string
	Config         ImagePerceiverConfigMap

	ReplicaCount       int32
	ServiceName        string
	ServiceAccountName string
}

func NewImagePerceiver(replicaCount int32, serviceAccountName string) *ImagePerceiver {
	memory, err := resource.ParseQuantity("2Gi")
	if err != nil {
		panic(err)
	}
	cpu, err := resource.ParseQuantity("500m")
	if err != nil {
		panic(err)
	}

	return &ImagePerceiver{
		PodName:            "perceptor-image-perceiver",
		Image:              "gcr.io/gke-verification/blackducksoftware/image-perceiver:latest",
		Port:               4000,
		CPU:                cpu,
		Memory:             memory,
		ConfigMapName:      "openshift-perceiver-config",
		ConfigMapMount:     "/etc/perceiver",
		ConfigMapPath:      "perceiver.yaml",
		ReplicaCount:       replicaCount,
		ServiceName:        "image-perceiver",
		ServiceAccountName: serviceAccountName,
	}
}

func (ip *ImagePerceiver) container() *v1.Container {
	return &v1.Container{
		Name:            "pod-perceiver",
		Image:           ip.Image,
		ImagePullPolicy: "Always",
		Command:         []string{},
		Ports: []v1.ContainerPort{
			v1.ContainerPort{
				ContainerPort: ip.Port,
				Protocol:      "TCP",
			},
		},
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    ip.CPU,
				v1.ResourceMemory: ip.Memory,
			},
		},
		VolumeMounts: []v1.VolumeMount{
			v1.VolumeMount{
				Name:      ip.ConfigMapName,
				MountPath: ip.ConfigMapMount,
			},
		},
	}
}

func (ip *ImagePerceiver) ReplicationController() *v1.ReplicationController {
	return &v1.ReplicationController{
		ObjectMeta: v1meta.ObjectMeta{Name: ip.PodName},
		Spec: v1.ReplicationControllerSpec{
			Replicas: &ip.ReplicaCount,
			Selector: map[string]string{"name": ip.PodName},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: v1meta.ObjectMeta{Labels: map[string]string{"name": ip.PodName}},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						v1.Volume{
							Name: ip.ConfigMapName,
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{Name: ip.ConfigMapName},
								},
							},
						},
					},
					Containers:         []v1.Container{*ip.container()},
					ServiceAccountName: ip.ServiceAccountName,
					// TODO: RestartPolicy?  terminationGracePeriodSeconds? dnsPolicy?
				}}}}
}

func (ip *ImagePerceiver) Service() *v1.Service {
	return &v1.Service{
		ObjectMeta: v1meta.ObjectMeta{
			Name: ip.ServiceName,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name: ip.ServiceName,
					Port: ip.Port,
				},
			},
			Selector: map[string]string{"name": ip.ServiceName}}}
}

func (ip *ImagePerceiver) ConfigMap() *v1.ConfigMap {
	jsonBytes, err := json.Marshal(ip.Config)
	if err != nil {
		panic(err)
	}
	return MakeConfigMap(ip.ConfigMapName, ip.ConfigMapPath, string(jsonBytes))
}