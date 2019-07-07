package ec2

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestInstanceFamily(t *testing.T) {
	tests := []struct {
		input  string
		loose  bool
		expect string
	}{
		{"t3.small", false, "t3"},
		{"t3a.small", false, "t3a"},
		{"t3a.small", true, "t3"},
		{"r5ad.xlarge", false, "r5ad"},
		{"r5ad.xlarge", true, "r5"},
	}

	for _, test := range tests {
		require.Equal(t, test.expect, instanceFamily(test.input, test.loose))
	}
}

func TestImageCreationDate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		Input  *string
		Expect time.Time
	}{
		{nil, time.Time{}},
		{aws.String("2019-05-15T12:59:50.000Z"), time.Date(2019, 5, 15, 12, 59, 50, 0, time.UTC)},
	}

	for _, test := range tests {
		actual, _ := imageCreationDate(test.Input)
		require.Equal(t, test.Expect, actual)
	}
}
