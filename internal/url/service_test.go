package url

import "testing"

func TestValidateURL(t *testing.T) {
	test := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Valid HTTPS URL",
			input:   "https://example.com",
			wantErr: false,
		},
		{
			name:    "valid HTTP URL with path",
			input:   "http://example.com/products",
			wantErr: false,
		},
		{
			name:    "unsupported FTP scheme",
			input:   "ftp://example.com",
			wantErr: true,
		},
		{
			name:    "unsupported file scheme",
			input:   "file:///etc/passwd",
			wantErr: true,
		},
		{
			name:    "missing scheme",
			input:   "example.com",
			wantErr: true,
		},
		{
			name:    "missing hostname",
			input:   "https:///products",
			wantErr: true,
		},
		{
			name:    "empty URL",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateURL(tt.input); (err == nil) == tt.wantErr {
				t.Errorf("validate URL(%q) error: %v, wantErr: %v", tt.input, err, tt.wantErr)
			}

		})
	}
}
