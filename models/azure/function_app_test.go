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

import "testing"

func TestIsFunctionAppKind(t *testing.T) {
	cases := map[string]bool{
		"functionapp":                            true,
		"functionapp,linux":                      true,
		"functionapp,linux,container,kubernetes": true,
		"functionapp,workflowapp":                true,
		"FunctionApp, Linux":                     true,
		"app":                                    false,
		"app,linux":                              false,
		"app,linux,container":                    false,
		"app,container,windows":                  false,
		"hyperV":                                 false,
		"":                                       false,
	}
	for kind, want := range cases {
		if got := IsFunctionAppKind(kind); got != want {
			t.Errorf("IsFunctionAppKind(%q) = %v, want %v", kind, got, want)
		}
	}
}
