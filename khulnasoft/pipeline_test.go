package khulnasoft

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewHiddenJob(t *testing.T) {
	tests := []struct {
		name      string
		label     string
		stage     string
		image     string
		scripts   []string
		erroneous bool
	}{
		{
			name:      "with valid parameters",
			label:     ".label",
			stage:     "stage1",
			image:     "image1",
			erroneous: false,
		},
		{
			name:      "with invalid label",
			label:     "do not start with dot",
			stage:     "stage1",
			image:     "image1",
			erroneous: true,
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			j, err := newHiddenJob(spec.label, spec.stage, spec.image, spec.scripts)

			if spec.erroneous {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
				require.Equal(t, spec.label, j.Label())
				require.Equal(t, spec.stage, j.Stage)
				require.Equal(t, spec.image, j.Image)
				require.Equal(t, spec.scripts, j.Scripts)
			}
		})
	}
}
