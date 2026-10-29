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

package cmd

import (
	"context"
	"sort"
	"testing"

	"github.com/SpecterOps/AzureHound/v2/client"
	"github.com/SpecterOps/AzureHound/v2/client/mocks"
	"github.com/SpecterOps/AzureHound/v2/models"
	"github.com/SpecterOps/AzureHound/v2/models/azure"
	"go.uber.org/mock/gomock"
)

func init() {
	setupLogger()
}

// Kinds returned by the Microsoft.Web/sites list endpoint; every site must land in exactly one collector.
var siteKinds = map[string]string{
	"web-windows":           "app",
	"web-linux":             "app,linux",
	"web-linux-container":   "app,linux,container",
	"web-windows-hyperv":    "hyperV",
	"web-windows-container": "app,container,windows",
	"func-windows":          "functionapp",
	"func-linux":            "functionapp,linux",
	"func-logic-standard":   "functionapp,workflowapp",
}

func singleSubscription() chan interface{} {
	subscriptions := make(chan interface{}, 1)
	subscriptions <- AzureWrapper{Data: models.Subscription{}}
	close(subscriptions)
	return subscriptions
}

func collectNames(t *testing.T, channel <-chan interface{}) []string {
	var names []string
	for result := range channel {
		wrapper, ok := result.(AzureWrapper)
		if !ok {
			t.Fatalf("failed type assertion: got %T, want %T", result, AzureWrapper{})
		}
		switch data := wrapper.Data.(type) {
		case models.WebApp:
			names = append(names, data.Name)
		case models.FunctionApp:
			names = append(names, data.Name)
		default:
			t.Fatalf("unexpected data type %T", wrapper.Data)
		}
	}
	sort.Strings(names)
	return names
}

func TestListWebAppsKinds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockAzureClient(ctrl)
	mockChannel := make(chan client.AzureResult[azure.WebApp], len(siteKinds))
	for name, kind := range siteKinds {
		mockChannel <- client.AzureResult[azure.WebApp]{Ok: azure.WebApp{Name: name, Kind: kind}}
	}
	close(mockChannel)
	mockClient.EXPECT().TenantInfo().Return(azure.Tenant{}).AnyTimes()
	mockClient.EXPECT().ListAzureWebApps(gomock.Any(), gomock.Any()).Return(mockChannel).Times(1)

	got := collectNames(t, listWebApps(context.Background(), mockClient, singleSubscription()))
	want := []string{"web-linux", "web-linux-container", "web-windows", "web-windows-container", "web-windows-hyperv"}
	if len(got) != len(want) {
		t.Fatalf("got web apps %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got web apps %v, want %v", got, want)
		}
	}
}

func TestListFunctionAppsKinds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockAzureClient(ctrl)
	mockChannel := make(chan client.AzureResult[azure.FunctionApp], len(siteKinds))
	for name, kind := range siteKinds {
		mockChannel <- client.AzureResult[azure.FunctionApp]{Ok: azure.FunctionApp{Name: name, Kind: kind}}
	}
	close(mockChannel)
	mockClient.EXPECT().TenantInfo().Return(azure.Tenant{}).AnyTimes()
	mockClient.EXPECT().ListAzureFunctionApps(gomock.Any(), gomock.Any()).Return(mockChannel).Times(1)

	got := collectNames(t, listFunctionApps(context.Background(), mockClient, singleSubscription()))
	want := []string{"func-linux", "func-logic-standard", "func-windows"}
	if len(got) != len(want) {
		t.Fatalf("got function apps %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got function apps %v, want %v", got, want)
		}
	}
}
