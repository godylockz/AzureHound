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

func TestListRoleAssignments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockAzureClient(ctrl)
	mockChannel := make(chan client.AzureResult[azure.UnifiedRoleAssignment], 3)
	mockChannel <- client.AzureResult[azure.UnifiedRoleAssignment]{Ok: azure.UnifiedRoleAssignment{
		Entity: azure.Entity{Id: "tenant"}, DirectoryScopeId: "/",
	}}
	mockChannel <- client.AzureResult[azure.UnifiedRoleAssignment]{Ok: azure.UnifiedRoleAssignment{
		Entity: azure.Entity{Id: "app"}, DirectoryScopeId: "/app-object-id", DirectoryScope: azure.Application{AppId: "app-id"},
	}}
	mockChannel <- client.AzureResult[azure.UnifiedRoleAssignment]{Ok: azure.UnifiedRoleAssignment{
		Entity: azure.Entity{Id: "admin-unit"}, DirectoryScopeId: "/administrativeUnits/admin-unit-id",
	}}
	close(mockChannel)
	mockClient.EXPECT().TenantInfo().Return(azure.Tenant{}).AnyTimes()
	mockClient.EXPECT().ListAzureADRoleAssignments(gomock.Any(), gomock.Any()).Return(mockChannel).Times(1)

	roles := make(chan interface{}, 1)
	roles <- AzureWrapper{Data: models.Role{}}
	close(roles)

	result, ok := <-listRoleAssignments(context.Background(), mockClient, roles)
	if !ok {
		t.Fatalf("failed to receive from channel")
	}
	roleAssignments, ok := result.(AzureWrapper).Data.(models.RoleAssignments)
	if !ok {
		t.Fatalf("failed type assertion: got %T, want %T", result.(AzureWrapper).Data, models.RoleAssignments{})
	}

	want := map[string]string{
		"tenant":     "/",
		"app":        "/app-id",
		"admin-unit": "/administrativeUnits/admin-unit-id",
	}
	for _, assignment := range roleAssignments.RoleAssignments {
		if got := assignment.DirectoryScopeId; got != want[assignment.Id] {
			t.Errorf("assignment %s: got scope %q, want %q", assignment.Id, got, want[assignment.Id])
		}
	}
	if len(roleAssignments.RoleAssignments) != len(want) {
		t.Errorf("got %d role assignments, want %d", len(roleAssignments.RoleAssignments), len(want))
	}
}
