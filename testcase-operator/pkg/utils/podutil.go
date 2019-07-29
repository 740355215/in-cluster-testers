package utils

import (
	corev1 "k8s.io/api/core/v1"
)

func IsPodRunning(pod *corev1.Pod) bool {
	if pod.Status.Phase != corev1.PodRunning {
		return false
	}

	return true
}

func IsPodFailed(pod *corev1.Pod) bool {
	if pod.Status.Phase != corev1.PodFailed {
		return false
	}

	return true
}

func IsPodSucceeded(pod *corev1.Pod) bool {
	if pod.Status.Phase != corev1.PodSucceeded {
		return false
	}

	return true
}
