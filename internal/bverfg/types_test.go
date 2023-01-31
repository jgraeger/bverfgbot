package bverfg_test

import (
	"testing"

	"github.com/jgraeger/bverfgbot/internal/bverfg"
	"github.com/stretchr/testify/assert"
)

func TestParseCaseRef(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		shouldFail  bool
		expectedRef bverfg.CaseReference
	}{
		{
			name:       "Parse valid 20th century case ref",
			input:      "1 BvR 205/58",
			shouldFail: false,
			expectedRef: bverfg.CaseReference{
				Senate:        1,
				Type:          bverfg.Verfassungsbeschwerde,
				RunningNumber: 205,
				Year:          1958,
			},
		},
		{
			name:       "Parse valid 21th century case ref",
			input:      "2 BvB 1/13",
			shouldFail: false,
			expectedRef: bverfg.CaseReference{
				Senate:        2,
				Type:          bverfg.Parteiverbotsverfahren,
				RunningNumber: 1,
				Year:          2013,
			},
		},
		{
			name:        "Error with invalid senate",
			input:       "3 BvR 3/19",
			shouldFail:  true,
			expectedRef: bverfg.CaseReference{},
		},
		{
			name:        "Error with invalid senate",
			input:       "3 BvR 3/19",
			shouldFail:  true,
			expectedRef: bverfg.CaseReference{},
		},
		{
			name:        "Error with invalid running number",
			input:       "1 BvR -717/16",
			shouldFail:  true,
			expectedRef: bverfg.CaseReference{},
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			parsed, err := bverfg.ParseCaseRef(tc.input)
			if tc.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedRef, parsed)
		})
	}
}

func TestCaseRefString(t *testing.T) {
	testCases := []struct {
		name     string
		input    bverfg.CaseReference
		expected string
	}{
		{
			name: "Test decision from 21st century",
			input: bverfg.CaseReference{
				Senate:        2,
				Type:          bverfg.KonkreteNormenkontrolle,
				RunningNumber: 4,
				Year:          2020,
			},
			expected: "2 BvL 4/20",
		},
		{
			name: "Test decision from 20st century",
			input: bverfg.CaseReference{
				Senate:        2,
				Type:          bverfg.VorkonstitutionelleFortgeltung,
				RunningNumber: 3,
				Year:          1956,
			},
			expected: "2 BvO 3/56",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, tc.input.String())
		})
	}

}
