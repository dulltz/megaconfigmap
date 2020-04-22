package combiner

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCombiner_sortContents(t *testing.T) {
	tests := []struct {
		name    string
		args    *corev1.ConfigMapList
		want    []string
		wantErr bool
	}{
		{
			name: "valid",
			args: &corev1.ConfigMapList{Items: []corev1.ConfigMap{
				{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{OrderLabel: "1"}}, Data: map[string]string{PartialItemKey: "b"}},
				{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{OrderLabel: "0"}}, Data: map[string]string{PartialItemKey: "a"}},
				{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{OrderLabel: "2"}}, Data: map[string]string{PartialItemKey: "c"}},
			}},
			want:    []string{"a", "b", "c"},
			wantErr: false,
		},
		{
			name: "invalid: out-of-index but not panic",
			args: &corev1.ConfigMapList{Items: []corev1.ConfigMap{
				{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{OrderLabel: "2"}}, Data: map[string]string{PartialItemKey: "b"}},
				{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{OrderLabel: "1"}}, Data: map[string]string{PartialItemKey: "a"}},
				{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{OrderLabel: "3"}}, Data: map[string]string{PartialItemKey: "c"}},
			}},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid: unexpected value",
			args: &corev1.ConfigMapList{Items: []corev1.ConfigMap{
				{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{OrderLabel: "1"}}, Data: map[string]string{PartialItemKey: "b"}},
				{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{OrderLabel: "0"}}, Data: map[string]string{PartialItemKey: "a"}},
				{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{OrderLabel: "two"}}, Data: map[string]string{PartialItemKey: "c"}},
			}},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid: orderLabel not found",
			args: &corev1.ConfigMapList{Items: []corev1.ConfigMap{
				{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{OrderLabel: "1"}}, Data: map[string]string{PartialItemKey: "b"}},
				{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"aaa": "0"}}, Data: map[string]string{PartialItemKey: "a"}},
				{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{OrderLabel: "2"}}, Data: map[string]string{PartialItemKey: "c"}},
			}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &Combiner{}
			got, err := c.sortContents(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("sortContents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sortContents() got = %v, want %v", got, tt.want)
			}
		})
	}
}
