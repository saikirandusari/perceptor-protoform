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

// GetHubRequest will have the configuration to get the Black Duck Hub Model
type GetHubRequest struct {
	Hubs map[string]*Hub `json:"hubs"`
}

// CreateHubRequest will have the configuration to create the Black Duck Hub
type CreateHubRequest struct {
	Namespace        string `json:"namespace"`
	Flavor           string `json:"flavor"`
	DockerRegistry   string `json:"dockerRegistry"`
	DockerRepo       string `json:"dockerRepo"`
	HubVersion       string `json:"hubVersion"`
	AdminPassword    string `json:"adminPassword"`
	UserPassword     string `json:"userPassword"`
	PostgresPassword string `json:"postgresPassword"`
	IsRandomPassword bool   `json:"isRandomPassword"`
}

// Hub will have the configuration relation to hold the Black Duck Hub details
type Hub struct {
	Namespace        string `json:"namespace"`
	DockerRegistry   string `json:"dockerRegistry"`
	DockerRepo       string `json:"dockerRepo"`
	HubVersion       string `json:"hubVersion"`
	Flavor           string `json:"flavor"`
	AdminPassword    string `json:"adminPassword"`
	UserPassword     string `json:"userPassword"`
	PostgresPassword string `json:"postgresPassword"`
	IsRandomPassword bool   `json:"isRandomPassword"`
	Status           string `json:"status"`
	IPAddress        string `json:"ipAddress"`
}

// DeleteHubRequest will have the configuration to delete the Black Duck Hub
type DeleteHubRequest struct {
	Namespace string `json:"namespace"`
}
