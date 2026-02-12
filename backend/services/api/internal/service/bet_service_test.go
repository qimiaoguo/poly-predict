package service

import "testing"

func TestFindOutcomeIndex(t *testing.T) {
	labels := []string{"Yes", "No"}

	tests := []struct {
		name    string
		outcome string
		want    int
	}{
		{"lowercase yes matches Yes", "yes", 0},
		{"lowercase no matches No", "no", 1},
		{"uppercase YES matches Yes", "YES", 0},
		{"exact case Yes", "Yes", 0},
		{"exact case No", "No", 1},
		{"mixed case yEs", "yEs", 0},
		{"invalid outcome", "maybe", -1},
		{"empty outcome", "", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findOutcomeIndex(labels, tt.outcome)
			if got != tt.want {
				t.Errorf("findOutcomeIndex(%v, %q) = %d, want %d", labels, tt.outcome, got, tt.want)
			}
		})
	}
}

func TestFindOutcomeIndex_EmptyLabels(t *testing.T) {
	got := findOutcomeIndex(nil, "yes")
	if got != -1 {
		t.Errorf("findOutcomeIndex(nil, \"yes\") = %d, want -1", got)
	}
}
