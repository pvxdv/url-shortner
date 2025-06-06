package random

import "testing"

func TestNewRandomString(t *testing.T) {
	type args struct {
		length int
	}

	tests := []struct {
		name    string
		args    args
		wantLen int
	}{
		{
			name:    "Positive length 10",
			args:    args{length: 10},
			wantLen: 10,
		},
		{
			name:    "Positive length 20",
			args:    args{length: 20},
			wantLen: 20,
		},
		{
			name:    "Positive length 30",
			args:    args{length: 30},
			wantLen: 30,
		},
		{
			name:    "Positive length 40",
			args:    args{length: 40},
			wantLen: 40,
		},
		{
			name:    "Zero length",
			args:    args{length: 0},
			wantLen: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRandomString(tt.args.length); len(got) != tt.wantLen {
				t.Errorf("NewRandomString() = %s, len: %d want %d", got, len(got), tt.wantLen)
			}
		})
	}
}
