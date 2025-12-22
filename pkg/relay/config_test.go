package relay

import (
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		cfg     *Config
		wantErr string
	}{
		{
			name: "success",
			cfg: &Config{
				ProjectID: "test-project",
			},
		},
		{
			name: "missing_project_id",
			cfg: &Config{
				ProjectID: "",
			},
			wantErr: "PROJECT_ID is required",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.cfg.Validate()
			if err != nil {
				if tc.wantErr == "" {
					t.Errorf("Config.Validate() unexpected error: %v", err)
				} else if err.Error() != tc.wantErr {
					t.Errorf("Config.Validate() error %q, want %q", err, tc.wantErr)
				}
			} else if tc.wantErr != "" {
				t.Errorf("Config.Validate() expected error %q, got nil", tc.wantErr)
			}
		})
	}
}
