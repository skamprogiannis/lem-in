package parser_test

import (
	"lem-in/internal/graph"
	"lem-in/internal/parser"

	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseBuildsValidatedGraph(t *testing.T) {
	input := strings.Join([]string{
		"3",
		"# ordinary comment",
		"##start",
		"start 0 1",
		"middle 2 3",
		"##unknown",
		"##end",
		"end 4 5",
		"start-middle",
		"middle-end",
		"",
	}, "\r\n")

	got, err := parser.Parse(writeInput(t, input))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	want := &graph.Graph{
		NumAnts: 3,
		Start:   "start",
		End:     "end",
		Rooms: map[string]*graph.Room{
			"start":  {Name: "start", X: 0, Y: 1},
			"middle": {Name: "middle", X: 2, Y: 3},
			"end":    {Name: "end", X: 4, Y: 5},
		},
		Links: map[string][]string{
			"start":  {"middle"},
			"middle": {"start", "end"},
			"end":    {"middle"},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse graph = %#v, want %#v", got, want)
	}
}

func TestParseRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "empty input",
			input:   " \n\t\n",
			wantErr: "empty input file",
		},
		{
			name:    "malformed ant count",
			input:   "many\n",
			wantErr: "invalid number in ant count entry",
		},
		{
			name:    "ant count has multiple fields",
			input:   "1 2\n",
			wantErr: "invalid number in ant count entry",
		},
		{
			name:    "zero ants",
			input:   "0\n",
			wantErr: "invalid ant value provided",
		},
		{
			name: "zero ant count cannot be replaced by a later number",
			input: "0\n" +
				"1\n" +
				"##start\n" +
				"start 0 0\n" +
				"##end\n" +
				"end 1 0\n" +
				"start-end\n",
			wantErr: "invalid ant value provided",
		},
		{
			name:    "negative ants",
			input:   "-1\n",
			wantErr: "invalid ant value provided",
		},
		{
			name: "duplicate room",
			input: "1\n" +
				"##start\n" +
				"start 0 0\n" +
				"start 1 1\n",
			wantErr: "room start already exists",
		},
		{
			name: "malformed coordinate",
			input: "1\n" +
				"##start\n" +
				"start x 0\n",
			wantErr: "x value is malformed",
		},
		{
			name: "malformed y coordinate",
			input: "1\n" +
				"##start\n" +
				"start 0 y\n",
			wantErr: "y value is malformed",
		},
		{
			name: "malformed room",
			input: "1\n" +
				"##start\n" +
				"start 0\n",
			wantErr: "room lines must be of format",
		},
		{
			name: "reserved room name",
			input: "1\n" +
				"##start\n" +
				"Lstart 0 0\n",
			wantErr: "room names must not start with L",
		},
		{
			name: "unknown room in tunnel",
			input: "1\n" +
				"##start\n" +
				"start 0 0\n" +
				"##end\n" +
				"end 1 0\n" +
				"start-missing\n",
			wantErr: "room referenced by tunnel does not exist",
		},
		{
			name: "malformed tunnel",
			input: "1\n" +
				"##start\n" +
				"start 0 0\n" +
				"##end\n" +
				"end 1 0\n" +
				"start--end\n",
			wantErr: "malformed tunnel entry",
		},
		{
			name: "self link",
			input: "1\n" +
				"##start\n" +
				"start 0 0\n" +
				"##end\n" +
				"end 1 0\n" +
				"start-start\n",
			wantErr: "room cannot link to itself",
		},
		{
			name: "duplicate tunnel",
			input: "1\n" +
				"##start\n" +
				"start 0 0\n" +
				"##end\n" +
				"end 1 0\n" +
				"start-end\n" +
				"end-start\n",
			wantErr: "duplicate tunnel",
		},
		{
			name: "missing start",
			input: "1\n" +
				"room 0 0\n" +
				"##end\n" +
				"end 1 0\n" +
				"room-end\n",
			wantErr: "start room entry missing",
		},
		{
			name: "missing end",
			input: "1\n" +
				"##start\n" +
				"start 0 0\n" +
				"room 1 0\n" +
				"start-room\n",
			wantErr: "end room entry missing",
		},
		{
			name: "multiple start commands",
			input: "1\n" +
				"##start\n" +
				"start 0 0\n" +
				"##start\n",
			wantErr: "multiple start rooms are not allowed",
		},
		{
			name: "multiple end commands",
			input: "1\n" +
				"##end\n" +
				"end 0 0\n" +
				"##end\n",
			wantErr: "multiple end rooms are not allowed",
		},
		{
			name: "room after tunnels",
			input: "1\n" +
				"##start\n" +
				"start 0 0\n" +
				"##end\n" +
				"end 1 0\n" +
				"start-end\n" +
				"late 2 0\n",
			wantErr: "room declared after tunnels",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parser.Parse(writeInput(t, tc.input))
			if err == nil {
				t.Fatal("Parse accepted invalid input")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("Parse error = %q, want it to contain %q", err, tc.wantErr)
			}
		})
	}
}

func TestParseRejectsMissingFile(t *testing.T) {
	missingFile := filepath.Join(t.TempDir(), "missing.txt")

	_, err := parser.Parse(missingFile)
	if err == nil {
		t.Fatal("Parse accepted missing file")
	}
	if got, want := err.Error(), "input file not found"; got != want {
		t.Errorf("Parse error = %q, want %q", got, want)
	}
}

func writeInput(t *testing.T, input string) string {
	t.Helper()

	filePath := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(filePath, []byte(input), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}
	return filePath
}
