/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package conditions

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"

	appsv1alpha1 "github.com/NVIDIA/k8s-nim-operator/api/apps/v1alpha1"
)

func TestWarnIfNGCAPIKeyUnset(t *testing.T) {
	obj := &appsv1alpha1.NIMService{ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"}}

	t.Run("emits warning when authSecret empty", func(t *testing.T) {
		recorder := record.NewFakeRecorder(1)
		WarnIfNGCAPIKeyUnset(context.Background(), recorder, obj, "")
		select {
		case event := <-recorder.Events:
			if !strings.Contains(event, "Warning") || !strings.Contains(event, appsv1alpha1.NGCAPIKeyUnsetReason) {
				t.Fatalf("event %q missing Warning/%s", event, appsv1alpha1.NGCAPIKeyUnsetReason)
			}
			if !strings.Contains(event, appsv1alpha1.NGCAPIKeyUnsetWarning) {
				t.Fatalf("event %q missing warning message", event)
			}
		default:
			t.Fatal("expected warning event when authSecret is empty")
		}
	})

	t.Run("emits warning on related objects", func(t *testing.T) {
		recorder := record.NewFakeRecorder(2)
		pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test-pod", Namespace: "default"}}
		WarnIfNGCAPIKeyUnset(context.Background(), recorder, obj, "", pod)
		got := 0
		for i := 0; i < 2; i++ {
			select {
			case event := <-recorder.Events:
				got++
				if !strings.Contains(event, appsv1alpha1.NGCAPIKeyUnsetReason) {
					t.Fatalf("event %q missing %s", event, appsv1alpha1.NGCAPIKeyUnsetReason)
				}
			default:
				t.Fatalf("expected 2 events, got %d", got)
			}
		}
	})

	t.Run("skips when authSecret set", func(t *testing.T) {
		recorder := record.NewFakeRecorder(1)
		pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test-pod", Namespace: "default"}}
		WarnIfNGCAPIKeyUnset(context.Background(), recorder, obj, "ngc-api-secret", pod)
		select {
		case event := <-recorder.Events:
			t.Fatalf("expected no event when authSecret is set, got %q", event)
		default:
		}
	})

	t.Run("nil recorder is safe", func(t *testing.T) {
		WarnIfNGCAPIKeyUnset(context.Background(), nil, obj, "")
		WarnIfNGCAPIKeyUnset(context.Background(), nil, &corev1.Pod{}, "secret")
	})
}
