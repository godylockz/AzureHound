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

package rest

import (
	"encoding/base64"
	"testing"
)

func unsignedToken(payload string) string {
	return "e30." + base64.RawURLEncoding.EncodeToString([]byte(payload)) + ".sig"
}

func TestParseAud(t *testing.T) {
	cases := map[string]string{
		"https://management.azure.com/":              "https://management.azure.com",
		"https://management.core.windows.net/":       "https://management.azure.com",
		"https://management.core.usgovcloudapi.net/": "https://management.usgovcloudapi.net",
		"https://graph.microsoft.com":                "https://graph.microsoft.com",
	}
	for aud, want := range cases {
		if got, err := ParseAud(unsignedToken(`{"aud":"` + aud + `"}`)); err != nil {
			t.Errorf("ParseAud(%q) returned error %v", aud, err)
		} else if got != want {
			t.Errorf("ParseAud(%q) = %q, want %q", aud, got, want)
		}
	}
}
