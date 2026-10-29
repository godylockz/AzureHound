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

import (
	"encoding/json"
	"testing"
)

func TestDefinitionUnmarshal_RecurrenceIntervalNumberOrString(t *testing.T) {
	payload := []byte(`{
		"triggers": {
			"fixed": {"type": "Recurrence", "recurrence": {"frequency": "Hour", "interval": 1}},
			"parameterized": {"type": "Recurrence", "recurrence": {"frequency": "Minute", "interval": "@parameters('interval')"}}
		}
	}`)

	var d Definition
	if err := json.Unmarshal(payload, &d); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if got := d.Triggers["fixed"].Recurrence.Interval; got != float64(1) {
		t.Errorf("expected numeric interval 1, got %#v", got)
	}
	if got := d.Triggers["parameterized"].Recurrence.Interval; got != "@parameters('interval')" {
		t.Errorf("expected expression interval, got %#v", got)
	}
}
