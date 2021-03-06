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

package hub

import (
	"fmt"
	"strconv"
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
)

// CreateHubConfig will create the hub configMaps
func (hc *Creater) createHubConfig(createHub *v1.Hub, hubContainerFlavor *ContainerFlavor) map[string]*components.ConfigMap {
	configMaps := make(map[string]*components.ConfigMap)

	hubConfig := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: createHub.Name, Name: "hub-config"})
	hubConfig.AddData(map[string]string{
		"PUBLIC_HUB_WEBSERVER_HOST": "localhost",
		"PUBLIC_HUB_WEBSERVER_PORT": "443",
		"HUB_WEBSERVER_PORT":        "8443",
		"IPV4_ONLY":                 "0",
		"RUN_SECRETS_DIR":           "/tmp/secrets",
		"HUB_VERSION":               createHub.Spec.HubVersion,
		"HUB_PROXY_NON_PROXY_HOSTS": "solr",
	})

	configMaps["hub-config"] = hubConfig

	hubDbConfig := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: createHub.Name, Name: "hub-db-config"})
	hubDbConfig.AddData(map[string]string{
		"HUB_POSTGRES_ADMIN": "blackduck",
		"HUB_POSTGRES_USER":  "blackduck_user",
		"HUB_POSTGRES_PORT":  "5432",
		"HUB_POSTGRES_HOST":  "postgres",
	})

	configMaps["hub-db-config"] = hubDbConfig

	hubConfigResources := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: createHub.Name, Name: "hub-config-resources"})
	hubConfigResources.AddData(map[string]string{
		"webapp-mem":    hubContainerFlavor.WebappHubMaxMemory,
		"jobrunner-mem": hubContainerFlavor.JobRunnerHubMaxMemory,
		"scan-mem":      hubContainerFlavor.ScanHubMaxMemory,
	})

	configMaps["hub-config-resources"] = hubConfigResources

	hubDbConfigGranular := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: createHub.Name, Name: "hub-db-config-granular"})
	hubDbConfigGranular.AddData(map[string]string{"HUB_POSTGRES_ENABLE_SSL": "false"})

	configMaps["hub-db-config-granular"] = hubDbConfigGranular

	postgresBootstrap := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: createHub.Name, Name: "postgres-bootstrap"})
	var backupInSeconds int
	switch createHub.Spec.BackupUnit {
	case "Minute(s)":
		backupInSeconds, _ = strconv.Atoi(createHub.Spec.BackupInterval)
		backupInSeconds = backupInSeconds * 60
	case "Hour(s)":
		backupInSeconds, _ = strconv.Atoi(createHub.Spec.BackupInterval)
		backupInSeconds = backupInSeconds * 60 * 60
	case "Week(s)":
		backupInSeconds, _ = strconv.Atoi(createHub.Spec.BackupInterval)
		backupInSeconds = backupInSeconds * 60 * 60 * 7
	default:
		if strings.EqualFold(createHub.Spec.BackupInterval, "") {
			backupInSeconds = 1
		}
		backupInSeconds = backupInSeconds * 60 * 60
	}

	postgresBootstrap.AddData(map[string]string{"pgbootstrap.sh": fmt.Sprintf(`#!/bin/bash
    if [ ! -f /data/bds/backup/%s.sql ] && [ -f /data/bds/backup/%s.sql ]; then
			echo "clone data file found"
			while true; do
				if psql -c "SELECT 1" &>/dev/null; then
					echo "Migrating the data"
      		psql < /data/bds/backup/%s.sql
      		break
    		else
      		echo "unable to execute the SELECT 1"
      		sleep 10
    		fi
  		done
		fi;

		if [ -f /data/bds/backup/%s.sql ]; then
			echo "backup data file found"
			while true; do
				if psql -c "SELECT 1" &>/dev/null; then
					echo "Migrating the data"
      		psql < /data/bds/backup/%s.sql
      		break
    		else
      		echo "unable to execute the SELECT 1"
      		sleep 10
    		fi
  		done
		fi;

		if [ "%s" == "Yes" ]; then
			while true; do
			  echo "Dump the data"
				sleep %d;
				pg_dumpall -w > /data/bds/backup/%s.sql;
			done
		fi`, createHub.Name, createHub.Spec.DbPrototype, createHub.Spec.DbPrototype, createHub.Name, createHub.Name, createHub.Spec.BackupSupport, backupInSeconds, createHub.Name)})

	configMaps["postgres-bootstrap"] = postgresBootstrap

	postgresInit := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: createHub.Name, Name: "postgres-init"})
	postgresInit.AddData(map[string]string{"pginit.sh": `#!/bin/bash
		echo "executing bds init script"
    sh /usr/share/container-scripts/postgresql/pgbootstrap.sh &
    run-postgresql`})

	configMaps["postgres-init"] = postgresInit

	return configMaps
}
