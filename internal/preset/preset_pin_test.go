package preset

import "testing"

func TestResolvedInstall_Substitution(t *testing.T) {
	cases := []struct {
		name string
		tool Tool
		os   string
		want string
	}{
		{
			name: "no pin returns raw",
			tool: Tool{Source: "mise", InstallMac: "mise use -g node@lts"},
			os:   "darwin",
			want: "mise use -g node@lts",
		},
		{
			name: "mise @lts replaced",
			tool: Tool{Source: "mise", InstallMac: "mise use -g node@lts", Pin: "22.11.0"},
			os:   "darwin",
			want: "mise use -g node@22.11.0",
		},
		{
			name: "mise @latest replaced",
			tool: Tool{Source: "mise", InstallMac: "mise use -g pnpm@latest", Pin: "9.15.0"},
			os:   "darwin",
			want: "mise use -g pnpm@9.15.0",
		},
		{
			name: "npm scoped without ver gets pin appended",
			tool: Tool{Source: "npm", InstallMac: "npm install -g @openai/codex", Pin: "0.50.1"},
			os:   "darwin",
			want: "npm install -g @openai/codex@0.50.1",
		},
		{
			name: "npm unscoped without ver gets pin appended",
			tool: Tool{Source: "npm", InstallMac: "npm install -g yarn", Pin: "1.22.22"},
			os:   "darwin",
			want: "npm install -g yarn@1.22.22",
		},
		{
			name: "linux command used when os=linux",
			tool: Tool{
				Source:       "mise",
				InstallMac:   "brew install mise",
				InstallLinux: "mise use -g node@lts",
				Pin:          "22.11.0",
			},
			os:   "linux",
			want: "mise use -g node@22.11.0",
		},
		{
			name: "curl source leaves command alone",
			tool: Tool{
				Source:     "curl",
				InstallMac: "curl -fsSL https://astral.sh/uv/install.sh | sh",
				Pin:        "0.5.0",
			},
			os:   "darwin",
			want: "curl -fsSL https://astral.sh/uv/install.sh | sh",
		},
		{
			name: "brew source leaves command alone (no @ver pin)",
			tool: Tool{
				Source:     "brew",
				InstallMac: "brew install mise",
				Pin:        "2026.5.6",
			},
			os:   "darwin",
			want: "brew install mise",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.tool.ResolvedInstall(tc.os)
			if got != tc.want {
				t.Errorf("\n got: %q\nwant: %q", got, tc.want)
			}
		})
	}
}
