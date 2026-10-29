// Copyright (C) 2022 Specter Ops, Inc.
//
// This file is part of AzureHound.
//
// AzureHound is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// AzureHound is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package azure

import "strings"

type FunctionApp struct {
	Entity

	ExtendedLocation ExtendedLocation      `json:"extendedLocation,omitempty"`
	Identity         ManagedIdentity       `json:"identity,omitempty"`
	Kind             string                `json:"kind,omitempty"`
	Location         string                `json:"location,omitempty"`
	Name             string                `json:"name,omitempty"`
	Properties       FunctionAppProperties `json:"properties,omitempty"`
	Tags             map[string]string     `json:"tags,omitempty"`
	Type             string                `json:"type,omitempty"`
}

func (s FunctionApp) ResourceGroupName() string {
	parts := strings.Split(s.Id, "/")
	if len(parts) > 4 {
		return parts[4]
	} else {
		return ""
	}
}

func (s FunctionApp) ResourceGroupId() string {
	parts := strings.Split(s.Id, "/")
	if len(parts) > 5 {
		return strings.Join(parts[:5], "/")
	} else {
		return ""
	}
}

// IsFunctionAppKind reports whether an App Service site kind describes a function app. Kind is a
// comma-separated list such as "functionapp,linux" or "functionapp,workflowapp" (Logic Apps Standard).
// https://github.com/Azure/app-service-linux-docs/blob/master/Things_You_Should_Know/kind_property.md
func IsFunctionAppKind(kind string) bool {
	for _, value := range strings.Split(kind, ",") {
		if strings.EqualFold(strings.TrimSpace(value), "functionapp") {
			return true
		}
	}
	return false
}
