package dataplane

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/nginx/nginx-gateway-fabric/v2/internal/framework/helpers"
)

func TestSortPathRules_RegexByDescendingLength(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	rules := []PathRule{
		{Path: "/.*", PathType: PathTypeRegularExpression},
		{Path: "/perform.*", PathType: PathTypeRegularExpression},
		{Path: "/perform_logout.*", PathType: PathTypeRegularExpression},
	}

	sortPathRules(rules)

	// Expected order: longest first (descending length for regex)
	g.Expect(rules[0].Path).To(Equal("/perform_logout.*"))
	g.Expect(rules[1].Path).To(Equal("/perform.*"))
	g.Expect(rules[2].Path).To(Equal("/.*"))
}

func TestSortPathRules_MixedTypes(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	rules := []PathRule{
		{Path: "/.*", PathType: PathTypeRegularExpression},
		{Path: "/api", PathType: PathTypePrefix},
		{Path: "/perform.*", PathType: PathTypeRegularExpression},
		{Path: "/exact", PathType: PathTypeExact},
	}

	sortPathRules(rules)

	// The current sort groups regex paths with non-regex using alphabetical comparison
	// Since alphabetically /.* < /api < /exact < /perform.*, the regex rules are interleaved
	// The key is that when BOTH rules are regex, longer ones come first
	// In this mixed case, the order is: /.* (regex), /api (prefix), /exact (exact), /perform.* (regex)
	g.Expect(rules[0].Path).To(Equal("/.*"))
	g.Expect(rules[1].Path).To(Equal("/api"))
	g.Expect(rules[2].Path).To(Equal("/exact"))
	g.Expect(rules[3].Path).To(Equal("/perform.*"))
}

func TestSortPathRules_SameLengthRegex(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	rules := []PathRule{
		{Path: "/bbb.*", PathType: PathTypeRegularExpression},
		{Path: "/aaa.*", PathType: PathTypeRegularExpression},
	}

	sortPathRules(rules)

	// Same length: alphabetical for deterministic ordering
	g.Expect(rules[0].Path).To(Equal("/aaa.*"))
	g.Expect(rules[1].Path).To(Equal("/bbb.*"))
}

func TestSortPathRules_NonRegexUnchanged(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	rules := []PathRule{
		{Path: "/zebra", PathType: PathTypePrefix},
		{Path: "/apple", PathType: PathTypePrefix},
		{Path: "/apple", PathType: PathTypeExact},
	}

	sortPathRules(rules)

	// Non-regex paths: alphabetical by path, then by PathType
	g.Expect(rules[0].Path).To(Equal("/apple"))
	g.Expect(rules[0].PathType).To(Equal(PathTypeExact))
	g.Expect(rules[1].Path).To(Equal("/apple"))
	g.Expect(rules[1].PathType).To(Equal(PathTypePrefix))
	g.Expect(rules[2].Path).To(Equal("/zebra"))
}

func TestSortPathRules_EmptySlice(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	rules := []PathRule{}
	sortPathRules(rules)
	g.Expect(rules).To(BeEmpty())
}

func TestSortPathRules_SingleElement(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	rules := []PathRule{
		{Path: "/test.*", PathType: PathTypeRegularExpression},
	}
	sortPathRules(rules)
	g.Expect(rules).To(HaveLen(1))
	g.Expect(rules[0].Path).To(Equal("/test.*"))
}

func TestSort(t *testing.T) {
	t.Parallel()
	// timestamps
	earlier := metav1.Now()
	later := metav1.NewTime(earlier.Add(1 * time.Second))

	earlierTimestampMeta := &metav1.ObjectMeta{
		Name:              "hr1",
		Namespace:         "test",
		CreationTimestamp: earlier,
	}
	laterTimestampMeta := &metav1.ObjectMeta{
		Name:              "hr2",
		Namespace:         "test",
		CreationTimestamp: later,
	}
	laterTimestampButAlphabeticallyFirstMeta := &metav1.ObjectMeta{
		Name:              "hr3",
		Namespace:         "a-test",
		CreationTimestamp: later,
	}

	pathOnly := MatchRule{
		Match:  Match{},
		Source: earlierTimestampMeta,
	}
	twoHeadersEarlierTimestamp := MatchRule{
		Match: Match{
			Headers: []HTTPHeaderMatch{
				{
					Name:  "header1",
					Value: "value1",
				},
				{
					Name:  "header2",
					Value: "value2",
				},
			},
		},
		Source: earlierTimestampMeta,
	}
	twoHeadersOneParam := MatchRule{
		Match: Match{
			Headers: []HTTPHeaderMatch{
				{
					Name:  "header1",
					Value: "value1",
				},
				{
					Name:  "header2",
					Value: "value2",
				},
			},
			QueryParams: []HTTPQueryParamMatch{
				{
					Name:  "key1",
					Value: "value1",
				},
			},
		},
		Source: earlierTimestampMeta,
	}
	threeHeaders := MatchRule{
		Match: Match{
			Headers: []HTTPHeaderMatch{
				{
					Name:  "header1",
					Value: "value1",
				},
				{
					Name:  "header2",
					Value: "value2",
				},
				{
					Name:  "header3",
					Value: "value3",
				},
			},
		},
		Source: earlierTimestampMeta,
	}
	methodEarlierTimestamp := MatchRule{
		Match: Match{
			Method: helpers.GetPointer("POST"),
		},
		Source: earlierTimestampMeta,
	}
	methodLaterTimestamp := MatchRule{
		Match: Match{
			Method: helpers.GetPointer("POST"),
		},
		Source: earlierTimestampMeta,
	}
	twoHeadersLaterTimestamp := MatchRule{
		Match: Match{
			Headers: []HTTPHeaderMatch{
				{
					Name:  "header1",
					Value: "value1",
				},
				{
					Name:  "header2",
					Value: "value2",
				},
			},
		},
		Source: laterTimestampMeta,
	}
	twoHeadersLaterTimestampButAlphabeticallyBefore := MatchRule{
		Match: Match{
			Headers: []HTTPHeaderMatch{
				{
					Name:  "header1",
					Value: "value1",
				},
				{
					Name:  "header2",
					Value: "value2",
				},
			},
		},
		Source: laterTimestampButAlphabeticallyFirstMeta,
	}

	rules := []MatchRule{
		methodLaterTimestamp,
		pathOnly,
		twoHeadersEarlierTimestamp,
		twoHeadersOneParam,
		threeHeaders,
		methodEarlierTimestamp,
		twoHeadersLaterTimestamp,
		twoHeadersLaterTimestampButAlphabeticallyBefore,
	}

	sortedRules := []MatchRule{
		methodEarlierTimestamp,
		methodLaterTimestamp,
		threeHeaders,
		twoHeadersOneParam,
		twoHeadersEarlierTimestamp,
		twoHeadersLaterTimestampButAlphabeticallyBefore,
		twoHeadersLaterTimestamp,
		pathOnly,
	}

	sortMatchRules(rules)

	g := NewWithT(t)
	g.Expect(cmp.Diff(sortedRules, rules)).To(BeEmpty())
}
