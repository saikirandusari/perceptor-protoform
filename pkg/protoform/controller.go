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
	"time"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"

	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	log "github.com/sirupsen/logrus"
)

// PodListController defines a controller that will list pods
type PodListController struct {
	namespaces []string
}

// NewPodListController will create a new ListPodController
func NewPodListController(ns []string) *PodListController {
	return &PodListController{namespaces: ns}
}

// Run will print to debug output the status of the pods that were started
func (l *PodListController) Run(resources horizonapi.ControllerResources, stopCh chan struct{}) error {
	client := resources.KubeClient
	for cnt := 0; cnt < 10; cnt++ {
		for _, n := range l.namespaces {
			pods, _ := client.Core().Pods(n).List(v1meta.ListOptions{})
			for _, pod := range pods.Items {
				log.Debugf("Pod = %v -> %v", pod.Name, pod.Status.Phase)
			}
		}
		log.Debug("***************")
		time.Sleep(10 * time.Second)
	}

	return nil
}
