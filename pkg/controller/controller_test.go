package controller

import (
	"testing"

	"github.com/shaardie/k8s-restarter/pkg/config"
)

type testSelectable struct {
	Namespace string
	Labels    map[string]string
}

func (ts testSelectable) GetNamespace() string {
	return ts.Namespace
}
func (ts testSelectable) GetLabels() map[string]string {
	return ts.Labels
}

func Test_shouldSelect(t *testing.T) {
	type args struct {
		s            selectable
		matcher      config.Matcher
		defaultValue bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Matcher not enabled",
			args: args{
				s:            testSelectable{},
				matcher:      config.Matcher{Enabled: false},
				defaultValue: true,
			},
			want: true,
		},
		{
			name: "Namespace match",
			args: args{
				s: testSelectable{Namespace: "namespace"},
				matcher: config.Matcher{
					Enabled:   true,
					Selectors: []config.Selector{{Namespace: "namespace"}},
				},
				defaultValue: false,
			},
			want: true,
		},
		{
			name: "Namespace does not match",
			args: args{
				s: testSelectable{Namespace: "other namespace"},
				matcher: config.Matcher{
					Enabled:   true,
					Selectors: []config.Selector{{Namespace: "namespace"}},
				},
				defaultValue: true,
			},
			want: false,
		},
		{
			name: "Labels match",
			args: args{
				s: testSelectable{
					Labels: map[string]string{
						"label1": "label1",
						"label2": "label2",
						"label3": "label3",
					},
				},
				matcher: config.Matcher{
					Enabled: true,
					Selectors: []config.Selector{
						{
							MatchLabels: map[string]string{
								"label1": "label1",
								"label2": "label2",
							},
						},
					},
				},
				defaultValue: false,
			},
			want: true,
		},
		{
			name: "Labels does not match",
			args: args{
				s: testSelectable{
					Labels: map[string]string{
						"label1": "label1",
						"label3": "label3",
					},
				},
				matcher: config.Matcher{
					Enabled: true,
					Selectors: []config.Selector{
						{
							MatchLabels: map[string]string{
								"label1": "label1",
								"label2": "label2",
							},
						},
					},
				},
				defaultValue: true,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldSelect(tt.args.s, tt.args.matcher, tt.args.defaultValue); got != tt.want {
				t.Errorf("shouldSelect() = %v, want %v", got, tt.want)
			}
		})
	}
}
